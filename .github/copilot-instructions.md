# Copilot / Agent Instructions

This repo is a very small Go-based Telegram bot. These instructions help AI coding agents become productive quickly.

## Big picture
- **Entry point:** [main.go](main.go#L1-L200) — single-file Go app that initializes a Telegram bot and uses long-polling (`GetUpdatesChan`) to handle messages.
- **Dependencies:** Managed in [go.mod](go.mod#L1-L20). The project depends on `github.com/go-telegram-bot-api/telegram-bot-api/v5` to interact with Telegram.
- **Runtime:** Simple synchronous loop that reads from the updates channel and replies immediately. No database, no external services beyond Telegram.

## What to look for / Why things are structured this way
- Single file simplicity: the project is intentionally minimal; changes should preserve small surface area unless you introduce a clear reason to refactor into packages.
- Long-polling pattern: code uses `GetUpdatesChan` for polling; if adding features with higher throughput, consider switching to webhooks or moving update-processing into worker goroutines.

## Secrets and configuration
- The current `main.go` contains a hard-coded bot token — treat this as sensitive. Replace in-source tokens with environment configuration immediately.
- Preferred configuration pattern: read `TELEGRAM_BOT_TOKEN` from environment and fail fast if absent. Example snippet:

```go
token := os.Getenv("TELEGRAM_BOT_TOKEN")
if token == "" {
    log.Fatal("TELEGRAM_BOT_TOKEN environment variable required")
}
bot, err := tgbotapi.NewBotAPI(token)
```

## Build / Run / Debug
- Quick run (dev): `go run main.go` — this is the simplest way to run locally.
- Build: `go build -o bin/go-price-monitor ./...` then run the produced binary.
- The workspace also includes a task that runs `go run ${file}` (see the VS Code task named `run go`).
- Go version: use the Go toolchain compatible with `go.mod` (declared `go 1.25.3`).

## Project conventions and patterns you should follow
- Keep logic minimal and explicit in small changes; prefer adding helper packages only when multiple handlers or types are needed.
- Logging uses the standard `log` package; continue this pattern unless a structured logger is added project-wide.
- Messages and log text may contain Russian; preserve i18n awareness when editing strings.

## Integration points & extension ideas
- Telegram API: all external calls go through `github.com/go-telegram-bot-api/telegram-bot-api/v5` (see [go.mod](go.mod#L1-L20)).
- To add persistence or external monitoring, add a small package (e.g., `internal/store`) and keep `main.go` responsible only for wiring.

## Safety & testing notes
- Don't commit secrets. If you find an API token in source, remove it and prompt the user to rotate credentials.
- There are no tests currently — when adding behavior, include small unit tests and a runbook for running the bot locally with a test token.

## Useful files
- [main.go](main.go#L1-L200): bot setup and update loop.
- [go.mod](go.mod#L1-L20): dependency list and Go version.

## If you change behavior
- When you add message handlers, check for blocking operations inside the update loop and move them to goroutines or worker pools.
- Add clear instructions in the repository README if you introduce new setup steps (env vars, DB, external services).

If any section is unclear or you'd like more examples (for example, a `config` helper or a safer startup template), tell me which part to expand.
