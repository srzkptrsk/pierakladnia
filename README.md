# Pierakladnia — Translator Collaboration Service

A web app for collaborative game translation built with Go and MySQL.

## Screenshots

| | |
| --- | --- |
| ![Pierakladnia screenshot 1](https://ghuxuf7gwh.s3.amazonaws.com/github/pierakladnia_screenshot_01.png) | ![Pierakladnia screenshot 2](https://ghuxuf7gwh.s3.amazonaws.com/github/pierakladnia_screenshot_02.png) |
| ![Pierakladnia screenshot 3](https://ghuxuf7gwh.s3.amazonaws.com/github/pierakladnia_screenshot_03.png) | ![Pierakladnia screenshot 4](https://ghuxuf7gwh.s3.amazonaws.com/github/pierakladnia_screenshot_04.png) |

## Features

- **Strings & Translations** — manage source strings and their translations with status tracking (todo, draft, needs_review, done)
- **Search & Filtering** — filter strings by source text, translated text, status, and ID expressions such as ranges or lists, with filters remembered between visits
- **Comments** — discuss translations and glossary terms
- **Revision History** — review translation changes with user avatars in history views
- **Glossary** — maintain project glossary (characters, places, terms)
- **Project Statistics** — visual overview of translation progress with charts
- **Multi-project** — work on multiple translation projects
- **Auth** — register/login with email verification via AWS SES
- **Access control** — admin and user roles, per-user translation permissions
- **Import/Export** — JSON and GNU gettext `.po` format import/export
- **Project Export Settings** — configure per-project export filenames and `.po` header metadata

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

## Project Admin

Admins can manage projects from `/admin/projects`:

- create and edit projects
- assign users to projects
- configure export filename base for downloads
- configure `.po` metadata such as `Project-Id-Version`, `Language`, `Language-Team`, `Last-Translator`, and `Report-Msgid-Bugs-To`

## Import/Export Data

### Export

- `GET /admin/strings/export` downloads the active project as JSON using the configured filename base or project name
- `GET /admin/strings/export/po` downloads the active project as a GNU gettext `.po` file with optional per-project header fields

### Import (Admin only)

- `POST /admin/strings/import` uploads and imports a JSON file
- `POST /admin/strings/import/po` uploads and imports a GNU gettext `.po` file

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
