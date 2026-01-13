# Aviasales Telegram Price Monitor Bot

A high-performance **Go-based** service for automated flight price tracking via the **Aviasales (Travelpayouts) API**. The bot operates as a background worker, analyzing price fluctuations and delivering instant Telegram notifications [file:1][web:8].

## Key Features

- **Concurrent Background Worker**: Leverages **Goroutines** to handle background API polling and user commands simultaneously, ensuring zero latency for the end user [file:1][web:16].
- **State-Tracking Notification System**: Implements logic to compare real-time prices with historical database records, notifying users only when a price change is detected [file:1].
- **Scalable Architecture**: Database-driven design enables the bot to serve multiple users concurrently, maintaining individual subscription lists for each `chat_id` [file:1].

## Tech Stack

- **Language**: Go (Golang) 1.25 [file:1].
- **Database**: **PostgreSQL** (with `sqlx` for efficient data mapping).
- **Integrations**:
  - `telegram-bot-api/v5`: Telegram Bot API wrapper.
  - `joho/godotenv`: Environment configuration.
- **API**: Aviasales Data API V3 (**REST/HTTP**) [web:1].


## Installation & Setup

### 1. Clone the Repository
git clone https://github.com/Dzetner/go-tg_bot-price-monitor.git
cd go-tg_bot-price-monitor

### 2. Configuration
Create a .env file and provide your credentials:
TELEGRAM_TOKEN=123456:ABC-DEF1234...
TRAVELPAYOUTS_TOKEN=ваш_API_токен
DB_DSN=postgres://postgres:password@localhost:5432/bot_db?sslmode=disable

### 3. Infrastructure Deployment
Use Docker for rapid database deployment:
docker-compose up -d

### 4. Compilation & Execution
go run main.go

## Bot Commands

- /add [ORIGIN] [DEST] — Adds a new route for monitoring. Example: /add MOW HRG
- /list — Displays a list of all your active subscriptions with current prices
- /del [ORIGIN-DEST] — Stops monitoring and deletes the subscription. Example: /del MOW-HRG

## How It Works (Architecture)

1. Background Worker: Upon startup, a dedicated long-running Goroutine is launched to handle price monitoring.
2. Polling Cycle: Inside the goroutine, a time.Ticker initiates a check every 5 minutes:
   - The bot retrieves all active subscriptions from PostgreSQL.
   - For each route, it performs an HTTP request to the Aviasales API
3. State Comparison: If the price from the API differs from the value stored in the database:
   - The bot sends a notification to the user via Telegram.
   - The last_value field in the database is updated to track future changes.

## Development Roadmap

- Implementation of ticket searching for specific months.
- Adding price fluctuation charts directly within the Telegram chat.
- Integration with other aggregators for broader search results.
