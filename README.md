# Financial Monitoring System

A Telegram bot for tracking personal expenses: add purchases via text, receipt photo, or manual entry; automatically categorize line items; and view spending analytics for a selected period.

## Features

- **Purchase entry** — text input, receipt recognition from photos (Groq), manual line-item entry
- **Categorization** — keyword dictionary + LLM classifier for ambiguous items
- **Purchase list** — paginated browsing
- **Analytics** — summary for day, week, month, or half-year; detailed report with anomalies and AI-generated insights
- **Editing** — update name, quantity, and unit price of items in saved purchases
- **Organisations** — save frequently used stores

## Tech Stack

| Layer | Technologies |
|-------|--------------|
| Language | Go 1.25 |
| Bot | [telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api) |
| Database | PostgreSQL 16, [Bun](https://bun.uptrace.dev/) |
| Migrations | golang-migrate |
| AI | Groq API (Llama 3.3 / Llama 4 Scout) |

## Architecture

The project follows Clean Architecture principles:

```
cmd/bot/                  — entry point
internal/
  domain/                 — entities, repositories, gateway interfaces
  application/            — use cases, services, DI container
  infrastructure/         — PostgreSQL, Groq, config, logger
  presentation/telegram/  — bot command and callback handlers
```

## Bot Commands

| Command | Description |
|---------|-------------|
| `/start` | Welcome message and command list |
| `/text` | Add a purchase via text |
| `/photo` | Add a purchase from a receipt photo |
| `/manual` | Add a purchase manually, item by item |
| `/list` | List purchases |
| `/stats` | Spending analytics |
| `/edit` | Edit items in existing purchases |
| `/cancel` | Cancel the current operation |

## Make Commands

```bash
make build        # build the binary
make run          # build and run
make infra/up     # start PostgreSQL
make infra/down   # stop containers
make infra/remove # stop containers and remove volumes
make lint         # run golangci-lint
make vulncheck    # run govulncheck
```
