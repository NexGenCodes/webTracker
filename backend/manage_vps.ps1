# WebTracker VPS Management Script
# Usage: .\manage_vps.ps1 <command>
# Commands: logs, status, restart, deploy

# Pro-level: Load VPS_IP from .env to avoid exposing it in the script
$EnvFile = ".env"
if (Test-Path $EnvFile) {
    $IP = Select-String -Path $EnvFile -Pattern '^VPS_IP="(.*)"' | ForEach-Object { $_.Matches.Groups[1].Value }
}

if (-not $IP) {
    Write-Host "Error: VPS_IP not found in .env file." -ForegroundColor Red
    exit
}

$USER = "ubuntu"
$KEY = "ssh.pem"
$REMOTE_PATH = "/opt/webtracker"

function Show-Usage {
    Write-Host "Usage: .\manage_vps.ps1 <command>" -ForegroundColor Cyan
    Write-Host "Commands:"
    Write-Host "  logs    - Watch real-time logs from the VPS"
    Write-Host "  status  - Check the bot service status"
    Write-Host "  restart - Restart the bot service"
    Write-Host "  deploy  - Build for Linux and push to VPS"
}

if ($args.Count -eq 0) {
    Show-Usage
    exit
}

$command = $args[0].ToLower()

switch ($command) {
    "logs" {
        Write-Host "--- Watching VPS Logs (Ctrl+C to stop) ---" -ForegroundColor Yellow
        ssh -i $KEY "$USER@$IP" "sudo journalctl -u webtracker-bot -f -n 50"
    }
    "status" {
        Write-Host "--- Bot Service Status ---" -ForegroundColor Yellow
        ssh -i $KEY "$USER@$IP" "sudo systemctl status webtracker-bot"
    }
    "restart" {
        Write-Host "--- Restarting Bot ---" -ForegroundColor Yellow
        ssh -i $KEY "$USER@$IP" "sudo systemctl restart webtracker-bot"
    }
    "deploy" {
        Write-Host "--- 1. Building for Linux (AMD64) ---" -ForegroundColor Yellow
        $env:GOOS = "linux"; $env:GOARCH = "amd64"
        go build -o bot-linux ./cmd/bot
        
        Write-Host "--- 2. Stopping Remote Service ---" -ForegroundColor Yellow
        ssh -i $KEY "$USER@$IP" "sudo systemctl stop webtracker-bot"
        
        Write-Host "--- 3. Uploading Binary ---" -ForegroundColor Yellow
        scp -i $KEY .\bot-linux "$USER@$IP`:$REMOTE_PATH/bot"
        
        Write-Host "--- 4. Restarting Service ---" -ForegroundColor Yellow
        ssh -i $KEY "$USER@$IP" "sudo systemctl start webtracker-bot"
        
        Write-Host "Done! Deployment successful." -ForegroundColor Green
    }
    Default {
        Write-Host "Unknown command: $command" -ForegroundColor Red
        Show-Usage
    }
}
