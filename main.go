package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	tg "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type Subscription struct {
	ID        int    `db:"id"`
	ChatID    int64  `db:"chat_id"`
	URL       string `db:"url"`
	LastValue string `db:"last_value"`
}

func startChecking(db *sqlx.DB, bot *tg.BotAPI) {
	ticker := time.NewTicker(time.Minute * 5)
	var client = &http.Client{Timeout: 5 * time.Second}

	for range ticker.C {
		var subs []Subscription
		err := db.Select(&subs, "SELECT * FROM subscriptions")
		if err != nil {
			log.Println("не удалось выполнить запрос: ", err)
		}

		for _, sub := range subs {
			resp, err := client.Get(sub.URL)
			if err != nil {
				log.Println("не удалось выполнить запрос: ", err)
				continue
			}

			currentStatus := resp.Status
			log.Printf("Сайт %s ответил %s", sub.URL, currentStatus)
			resp.Body.Close()

			if sub.LastValue != currentStatus {
				_, err := db.Exec("UPDATE subscriptions SET last_value = $1 WHERE chat_id = $2 AND url = $3", currentStatus, sub.ChatID, sub.URL)
				if err != nil {
					log.Printf("ошибка изменения значений в БД для сайта %s:%v\n", sub.URL, err)
					bot.Send(tg.NewMessage(sub.ChatID, fmt.Sprintf("Произошла ошибка при обновлении статуса сайта %s", sub.URL)))
					continue
				}

				sub.LastValue = currentStatus
				bot.Send(tg.NewMessage(sub.ChatID, fmt.Sprintf("Сайт %s изменил свой статус на %s", sub.URL, currentStatus)))
			}
		}
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("ошибка загрузки .env файла")
	}

	dsn := os.Getenv("DB_DSN")
	token := os.Getenv("TELEGRAM_TOKEN")

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln("ошибка подключения к БД:", err)
	}
	log.Println("Успешное подключение к базе данных!")

	schema := `
    CREATE TABLE IF NOT EXISTS subscriptions (
        id SERIAL PRIMARY KEY,
        chat_id BIGINT NOT NULL,
        url TEXT NOT NULL,
        last_value TEXT DEFAULT ''
    );`
	db.MustExec(schema)

	bot, err := tg.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	go startChecking(db, bot)
	log.Printf("Авторизован под аккаунтом %s", bot.Self.UserName)
	bot.Debug = true

	u := tg.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := update.Message.Text
		chatID := update.Message.Chat.ID

		if len(text) > 5 && text[:4] == "/add" {
			url := text[5:]

			_, err := db.Exec("INSERT INTO subscriptions (chat_id, url) VALUES ($1, $2)", chatID, url)
			if err != nil {
				log.Println("ошибка записи в БД:", err)
				bot.Send(tg.NewMessage(chatID, "Произошла ошибка при сохранении ссылки."))
				continue
			}

			bot.Send(tg.NewMessage(chatID, "Ссылка добавлена! Я начну за ней следить."))
			continue
		}

		msg := tg.NewMessage(chatID, "Пришли мне ссылку в формате: /add https://example.com")
		bot.Send(msg)
	}
}
