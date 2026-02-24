#!/bin/bash
set -e
cd "$(dirname "$0")/../deploy/docker"
docker compose up -d
echo "Ensuring MySQL is ready..."
sleep 10
cd ../../
export APP_CONFIG=./config/config.example.json

# Wait until MySQL is genuinely ready
until docker exec pierakladnia-mysql mysqladmin ping -uapp -papp_password --silent; do
    echo "Waiting for MySQL..."
    sleep 2
done

# Run migrations
bash scripts/migrate_up.sh

echo "Starting app..."
go run ./cmd/app
