#!/usr/bin/env bash
set -euo pipefail

APP_DIR="/home/user/Projects/prozessor_gostack"
BIN="/usr/local/bin/prozessor"
SERVICE="prozessor"

cd "$APP_DIR"

echo "Pulling latest changes..."
git pull

echo "Building CSS..."
npm run build

echo "Generating templates..."
templ generate

echo "Building binary..."
go build -C ./cmd -o "$BIN" .

echo "Restarting service..."
sudo systemctl restart "$SERVICE"

echo "Done. Status:"
systemctl status "$SERVICE" --no-pager
