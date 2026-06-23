#!/bin/bash
set -e

BOT_BINARY="${BOT_BINARY:-/tmp/bot-new}"
INSTALL_DIR="${INSTALL_DIR:-/opt/webtracker}"
SERVICE_USER="${SERVICE_USER:-root}"

echo "=== WebTracker Deploy ==="

# 1. Install new binary
echo "Installing binary..."
sudo cp "$BOT_BINARY" "$INSTALL_DIR/bot"
sudo chmod +x "$INSTALL_DIR/bot"

# 2. Create/update API service
echo "Creating webtracker-api.service..."
sudo tee /etc/systemd/system/webtracker-api.service > /dev/null <<'SERVICEEOF'
[Unit]
Description=WebTracker HTTP API
After=network.target redis-server.service
Wants=redis-server.service
[Service]
Type=simple
User=root
WorkingDirectory=/opt/webtracker
ExecStart=/opt/webtracker/bot --mode=api
Restart=always
RestartSec=10
StartLimitIntervalSec=300
StartLimitBurst=5
EnvironmentFile=/opt/webtracker/.env
[Install]
WantedBy=multi-user.target
SERVICEEOF

# 3. Update bot service to use --mode=bot
echo "Updating webtracker-bot.service..."
sudo tee /etc/systemd/system/webtracker-bot.service > /dev/null <<'SERVICEEOF'
[Unit]
Description=WebTracker WhatsApp Bot
After=network.target redis-server.service
Wants=redis-server.service
[Service]
Type=simple
User=root
WorkingDirectory=/opt/webtracker
ExecStart=/opt/webtracker/bot --mode=bot
Restart=always
RestartSec=10
StartLimitIntervalSec=300
StartLimitBurst=5
EnvironmentFile=/opt/webtracker/.env
[Install]
WantedBy=multi-user.target
SERVICEEOF

# 4. Reload systemd and restart services
echo "Reloading systemd..."
sudo systemctl daemon-reload

echo "Enabling both services..."
sudo systemctl enable webtracker-api.service
sudo systemctl enable webtracker-bot.service

echo "Starting webtracker-api.service..."
sudo systemctl restart webtracker-api.service

echo "Starting webtracker-bot.service..."
sudo systemctl restart webtracker-bot.service

echo "=== Done ==="
echo "API status:"
sudo systemctl status webtracker-api.service --no-pager | head -5
echo ""
echo "Bot status:"
sudo systemctl status webtracker-bot.service --no-pager | head -5
