# Production WhatsApp Bot 🤖

Optimized Go bot using `whatsmeow` + Worker Pools + Direct Supabase API.

## 🏗️ Architecture

- **Protocol**: `whatsmeow` (Multi-Device)
- **Concurrency**: 3 Workers consuming a buffered channel (Limit: 100)
- **Database**: Direct `net/http` calls to Supabase (No heavy SDKs)
- **Storage**: `sqlite3` (auth.db) for session persistence

## 📦 Build for Production (Linux/AMD64)

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
