#!/bin/bash
set -eu

log() {
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

cd "$HOME/discord-bot"

GIT_OUTPUT=$(git pull --ff-only)

if [ "$GIT_OUTPUT" = "Already up to date." ]; then
  log "No changes; nothing to do"
  exit 0
fi

log "Updating bot"

go mod download
/usr/local/go/bin/go build ./cmd/jarvbot/

sudo cp /usr/local/jarvbot /usr/local/jarvbot.bak || true
sudo mv ./jarvbot /usr/local/jarvbot
sudo systemctl restart discordbot

log "Update complete"
