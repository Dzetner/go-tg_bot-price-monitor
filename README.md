# Telegram Price Monitor Bot

Бот на Go, который по расписанию делает HTTP‑запросы к заданным URL и присылает уведомления в Telegram при изменении статуса ответа.

## Возможности

- Добавление ссылки для мониторинга командой  
  `/add https://example.com`
- Периодические проверки всех ссылок каждые 5 минут.
- Сохранение подписок пользователей в PostgreSQL.
- Уведомление, если HTTP‑статус сайта изменился.

## Стек технологий

- Go 1.25+
- Telegram Bot API (`github.com/go-telegram-bot-api/telegram-bot-api/v5`)
- PostgreSQL (`github.com/lib/pq`, `github.com/jmoiron/sqlx`)
- `godotenv` для загрузки переменных окружения из `.env`

## Подготовка

1. Создать `.env` в корне проекта:
   ```env
   TELEGRAM_TOKEN=ваш_токен_бота
   DB_DSN=postgres://myuser:mypassword@localhost:5432/monitor_db?sslmode=disable
2. Убедиться, что установлены Docker и Docker Compose.

## Запуск базы данных (Docker)

- `docker compose up -d`
- База поднимется с пользователем myuser, паролем mypassword, базой monitor_db на порту 5432.

## Запуск бота

- `go run ./cmd/bot`
- При старте бот подключается к PostgreSQL, создаёт таблицу subscriptions, авторизуется в Telegram и запускает фоновую проверку URL каждые 5 минут.

## Использование

- Отправьте боту: `/add https://example.com`
- Бот начнёт раз в 5 минут проверять HTTP‑статус.
- При изменении придёт сообщение: Сайт https://example.com изменил свой статус на 200 OK.
- Если сообщение не начинается с /add, бот отвечает подсказкой с корректным форматом.

## Идеи для развития

- Мониторинг цен (парсинг HTML/JSON, а не только статуса) на какой-то продукт
