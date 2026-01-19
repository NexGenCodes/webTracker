#!/bin/bash
# AWS EC2 Free Tier Setup Script for WhatsApp Bot
# Instance: t2.micro/t3.micro (1GB RAM)
# OS: Ubuntu 22.04 LTS

set -e  # Exit on error

echo "================================================"
echo "AWS EC2 Setup for WhatsApp Bot (1GB RAM)"
echo "================================================"

# 1. Create 2GB Swap File (CRITICAL for 1GB RAM)
echo ""
echo "[1/5] Creating 2GB swap file..."
sudo fallocate -l 2G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
echo "✓ Swap file created and enabled"
free -h

# 2. Update System
echo ""
echo "[2/5] Updating system packages..."
sudo apt update && sudo apt upgrade -y

# 3. Install Font Dependencies (Required for gg)
echo ""
echo "[3/5] Installing Font dependencies..."
sudo apt install -y fontconfig libfreetype6

# Refresh font cache
sudo fc-cache -f -v

# 5. Install Go 1.21+
echo ""
echo "[5/5] Installing Go..."
GO_VERSION="1.21.6"
wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
rm go${GO_VERSION}.linux-amd64.tar.gz

# Add Go to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
export PATH=$PATH:/usr/local/go/bin

# Verify Go installation
go version
echo "✓ Go installed successfully"

echo ""
echo "================================================"
echo "✓ Setup Complete!"
echo "================================================"
echo ""
echo "Next Steps:"
echo "1. Clone your repository: git clone <your-repo>"
echo "2. Navigate to backend: cd webTracker/backend"
echo "3. Install dependencies: go mod download"
echo "4. Create .env file with your credentials"
echo "5. Build: go build -o bot cmd/bot/main.go"
echo "6. Run: ./bot"
echo ""
echo "Memory Status:"
free -h
echo ""
echo "Swap Status:"
swapon --show
