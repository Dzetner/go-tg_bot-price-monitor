package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type AviasalesResponse struct {
	Data []struct {
		Price       float64 `json:"price"`
		DepartureAt string  `json:"departure_at"`
	} `json:"data"`
}

type Subscription struct {
	ID        int    `db:"id"`
	ChatID    int64  `db:"chat_id"`
	URL       string `db:"url"`
	LastValue string `db:"last_value"`
}

func startChecking(db *sqlx.DB, bot *tg.BotAPI) {
	ticker := time.NewTicker(time.Minute * 5)
	client := &http.Client{Timeout: 10 * time.Second}
	apiToken := os.Getenv("TRAVELPAYOUTS_TOKEN")

	for range ticker.C {
		var subs []Subscription
		if err := db.Select(&subs, "SELECT * FROM subscriptions"); err != nil {
			log.Println("ошибка получения подписок:", err)
			continue
		}

		for _, sub := range subs {
			log.Printf("Начинаю проверку для: %s", sub.URL)
			parts := strings.Split(sub.URL, "-")
			if len(parts) != 2 {
				continue
			}
			origin, destination := parts[0], parts[1]

			apiURL := fmt.Sprintf("https://api.travelpayouts.com/aviasales/v3/prices_for_dates?origin=%s&destination=%s&token=%s&currency=rub&unique=true", origin, destination, apiToken)

			resp, err := client.Get(apiURL)
			if err != nil {
				log.Printf("ошибка запроса к API (%s): %v", sub.URL, err)
				continue
			}

			var data AviasalesResponse
			if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
				log.Printf("Ошибка JSON для %s: %v", sub.URL, err)
				resp.Body.Close()
				continue
			}
			resp.Body.Close()

			log.Printf("Данные от API для %s: %+v", sub.URL, data)

			var minPrice float64
			var departureDate string
			for _, flight := range data.Data {
				if minPrice == 0 || flight.Price < minPrice {
					minPrice = flight.Price
					departureDate = flight.DepartureAt
				}
			}
			if minPrice == 0 {
				log.Printf("Билеты для %s не найдены", sub.URL)
				continue
			}

			currentPriceStr := fmt.Sprintf("%.0f", minPrice)

			if sub.LastValue != currentPriceStr {
				displayDate := departureDate
				if len(departureDate) > 10 {
					displayDate = departureDate[:10]
				}
				msgText := fmt.Sprintf("Билет %s за %s руб.\nДата вылета: %s", sub.URL, currentPriceStr, displayDate)
				if sub.LastValue != "" {
					msgText += fmt.Sprintf("\n Старая цена: %s руб", sub.LastValue)
				}

				bot.Send(tg.NewMessage(sub.ChatID, msgText))

				db.Exec("UPDATE subscriptions SET last_value = $1 WHERE id = $2", currentPriceStr, sub.ID)
				sub.LastValue = currentPriceStr
			}
		}
	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Предупреждение: .env не загружен, используем системные переменные")
	}

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("Ошибка: переменная DB_DSN не установлена")
	}

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к БД: %v", err)
	}
	log.Println("Успешное подключение к базе данных!")

	schema := `CREATE TABLE IF NOT EXISTS subscriptions (
		id SERIAL PRIMARY KEY,
		chat_id BIGINT NOT NULL,
		url TEXT NOT NULL,
		last_value TEXT DEFAULT ''
	);`
	db.MustExec(schema)

	bot, _ := tg.NewBotAPI(os.Getenv("TELEGRAM_TOKEN"))
	go startChecking(db, bot)

	u := tg.NewUpdate(0)
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		text := update.Message.Text
		if text == "/list" {
			var subs []Subscription
			db.Select(&subs, "SELECT * FROM subscriptions WHERE chat_id = $1", chatID)

			if len(subs) == 0 {
				bot.Send(tg.NewMessage(chatID, "У тебя пока нет активных подписок. Используй /add, чтобы добавить!"))
				continue
			}

			resp := "Твои подписки:\n\n"
			for _, s := range subs {
				status := s.LastValue
				if status == "" {
					status = "ожидает проверки"
				}
				resp += fmt.Sprintf("%s — %s руб.\n", s.URL, status)
			}
			bot.Send(tg.NewMessage(chatID, resp))
			continue
		}
		if strings.HasPrefix(text, "/del") {
			args := strings.Fields(text)
			if len(args) != 2 {
				bot.Send(tg.NewMessage(chatID, "Используй: /del MOW-HRG"))
				continue
			}
			direction := strings.ToUpper(args[1])
			_, err := db.Exec("DELETE FROM subscriptions WHERE chat_id = $1 AND url = $2", chatID, direction)
			if err != nil {
				bot.Send(tg.NewMessage(chatID, "Ошибка при удалении."))
				continue
			}
			bot.Send(tg.NewMessage(chatID, "Подписка на "+direction+" удалена."))
			continue
		}
		if strings.HasPrefix(text, "/add") {
			args := strings.Fields(text)
			if len(args) != 3 {
				bot.Send(tg.NewMessage(chatID, "Используй: /add MOW HRG"))
				continue
			}
			direction := fmt.Sprintf("%s-%s", strings.ToUpper(args[1]), strings.ToUpper(args[2]))
			var count int
			db.Get(&count, "SELECT count(*) FROM subscriptions WHERE chat_id=$1 AND url=$2", chatID, direction)
			if count > 0 {
				bot.Send(tg.NewMessage(chatID, "Вы уже подписаны на это направление!"))
				continue
			}
			db.Exec("INSERT INTO subscriptions (chat_id, url) VALUES ($1, $2)", chatID, direction)
			bot.Send(tg.NewMessage(chatID, "Подписка оформлена на "+direction))
			continue
		}
		bot.Send(tg.NewMessage(chatID, "Пришли мне: /add MOW HRG"))
	}
}
