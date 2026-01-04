package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Вставьте сюда токен, который дал @BotFather
	bot, err := tgbotapi.NewBotAPI("8291846849:AAFebxk0ilgN0TgL-7CXGGrib5SqjoQSSNU")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Авторизован под аккаунтом %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Привет! Я твой будущий монитор цен.")
			bot.Send(msg)
		}
	}
}
