# Production WhatsApp Bot 🤖

Optimized Go bot using `whatsmeow` + Worker Pools + Direct Supabase API.

## 🏗️ Architecture

- **Protocol**: `whatsmeow` (Multi-Device)
- **Concurrency**: Worker Pools for scalable message processing
- **Database**: Supabase (PostgreSQL)
- **Rendering**: Native Go `gg` library (No Chrome required)
- **Monitoring**: Built-in health check and vitals monitoring

## 📂 Project Structure

- `assets/`: Image and font assets
- `cmd/bot/`: Application entry point
- `internal/`: Private library code
  - `app/`: App lifecycle and initialization
  - `commands/`: Command dispatching logic
  - `config/`: Configuration management
  - `health/`: Health server and monitoring
  - `logger/`: Structured logging
  - `models/`: Domain models
  - `parser/`: Manifest parsing (Regex + AI)
  - `supabase/`: Database interactions
  - `utils/`: Receipt rendering and helpers
  - `whatsapp/`: WhatsApp client and events
  - `worker/`: Async job workers

To run this on a standard VPS (Ubuntu/Debian):

```bash
# Set environment for cross-compilation
$env:GOOS="linux"; $env:GOARCH="amd64"
go build -ldflags="-s -w" -o bot-linux-amd64 main.go
```

*Note: `-s -w` strips debug symbols to reduce binary size.*

## 🚀 Run with PM2

1. Upload `bot-linux-amd64`, `.env` (or set sys vars), and `ecosystem.config.js` to your VPS.
2. Run with PM2:

   ```bash
   pm2 start ecosystem.config.js
   pm2 save
   ```

## 🧪 Local Test (Windows)

```bash
$env:SUPABASE_URL="https://your-project.supabase.co"
$env:SUPABASE_SERVICE_ROLE_KEY="your-service-role-key"
go run main.go
```
