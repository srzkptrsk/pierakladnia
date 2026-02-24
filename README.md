# Pierakladnia — Translator Collaboration Service

A web app for collaborative game translation built with Go and MySQL.

## Features

- **Strings & Translations** — manage source strings and their translations with status tracking (todo, draft, needs_review, done)
- **Comments** — discuss translations and glossary terms
- **Glossary** — maintain project glossary (characters, places, terms)
- **Multi-project** — work on multiple translation projects
- **Auth** — register/login with email verification via AWS SES
- **Access control** — admin and user roles, per-user translation permissions
- **Import/Export** — JSON-based string import and export

## Tech Stack

- **Go** (net/http, html/template)
- **MySQL 8** (utf8mb4)
- **AWS SES** for email verification
- **No frameworks** — lightweight, single binary

## Quick Start

### 1. Run MySQL via Docker

```bash
cd deploy/docker
docker compose up -d
```

### 2. Configure

```bash
cp config/config.example.json config/config.local.json
# Edit config.local.json with your DB credentials and SES keys
```

### 3. Run Migrations

```bash
./scripts/migrate_up.sh
```

### 4. Create Admin User

```bash
export APP_CONFIG=./config/config.local.json
go run ./cmd/admin create-admin --email admin@example.com --password yourpassword
```

### 5. Run the App

```bash
export APP_CONFIG=./config/config.local.json
go run ./cmd/app
```

## Admin CLI

```bash
# Create admin user
go run ./cmd/admin create-admin --email admin@example.com --password secret

# Delete a user
go run ./cmd/admin delete-user --email user@example.com
```

## Production Build (linux/arm64)

```bash
./scripts/build.sh
```

Output goes to `dist/pierakladnia/`.

## Deploy (EC2 / systemd)

```bash
# Copy dist/pierakladnia/ to your server, then:
sudo cp deploy/pierakladnia.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable pierakladnia
sudo systemctl start pierakladnia

# Logs
sudo journalctl -u pierakladnia -f
```
