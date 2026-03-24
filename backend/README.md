# NexGen WebTracker WhatsApp Bot 🤖

Optimized Go bot for logistics tracking using `whatsmeow` + Worker Pools + Postgres (via SQLC).

## 🏗️ Architecture

- **Protocol**: `whatsmeow` (Multi-Device)
- **Concurrency**: Worker Pools for scalable message processing
- **Database**: Postgres (Pgx) with type-safe SQLC adapters
- **Architecture**: Clean Architecture (Domain, UseCase, Adapter)
- **Configuration**: Structured and validated with `cleanenv`
- **Validation**: Strict API request validation with `validator.v10`
- **Rendering**: Native Go `gg` library (No Chrome required)
- **Parsing**: Advanced Regex extraction with Gemini AI Fallback
- **Monitoring**: Built-in health check (with DB ping) and vitals monitoring
- **Containerization**: Official `Dockerfile` provided for multi-stage builds

## 📂 Project Structure

- `assets/`: Image and font assets
- `cmd/bot/`: Application entry point
- `internal/`: Private library code
  - `adapter/db/`: SQLC-generated Postgres adapters
  - `commands/`: Command dispatching logic
  - `config/`: Configuration management
  - `usecase/`: Business logic layer (Shipment & Config)
  - `logger/`: Structured logging
  - `models/`: Manifest and domain models
  - `parser/`: Manifest parsing (Regex + AI) - Now supports ID/Passport extraction
  - `shipment/`: Core logistics logic and status management
  - `utils/`: Receipt rendering, waybill generation, and helpers
  - `whatsapp/`: WhatsApp client and event handling
  - `worker/`: Async job workers

## 🚀 Deployment (VPS)

To build for a standard Linux VPS:

```bash
# Cross-compile for Linux
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -ldflags="-s -w" -o bot-linux ./cmd/bot
```

### Run with PM2

1. Upload `bot-linux`, `.env`, and `ecosystem.config.js` to your VPS.
2. Run with PM2:

   ```bash
   pm2 start ecosystem.config.js
   pm2 save
   ```

## 🧪 Local Dev (Windows)

```powershell
go run ./cmd/bot/main.go
```

## 📝 Features

- **Automated Manifest Parsing**: Extracts Sender, Receiver, Phone, Address, Email, and ID/Passport numbers.
- **Dynamic Receipts**: Generates high-quality branded receipts (JPEG) sent directly via WhatsApp.
- **High Performance**: Pgx connection pooling + SQLC for optimized query execution.
- **Rate Limiting**: Integrated rate limits for AI parsing to prevent API quota exhaustion.
- **Admin Commands**: `!stats`, `!edit`, `!delete` for control and correction.
