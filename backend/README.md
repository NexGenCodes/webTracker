# NexGen WebTracker WhatsApp Bot 🤖

Optimized Go bot for logistics tracking using `whatsmeow` + Worker Pools + Local SQLite.

## 🏗️ Architecture

- **Protocol**: `whatsmeow` (Multi-Device)
- **Concurrency**: Worker Pools for scalable message processing
- **Database**: SQLite (Local storage optimized for 1GB RAM)
- **Rendering**: Native Go `gg` library (No Chrome required)
- **Parsing**: Advanced Regex extraction with Gemini AI Fallback
- **Monitoring**: Built-in health check and vitals monitoring

## 📂 Project Structure

- `assets/`: Image and font assets
- `cmd/bot/`: Application entry point
- `internal/`: Private library code
  - `commands/`: Command dispatching logic
  - `config/`: Configuration management
  - `localdb/`: SQLite database client and operations
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
- **Local Database**: Fast, embedded SQLite storage with automatic WAL mode management.
- **Rate Limiting**: Integrated rate limits for AI parsing to prevent API quota exhaustion.
- **Admin Commands**: `!stats`, `!edit`, `!delete` for control and correction.
