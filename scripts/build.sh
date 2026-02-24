#!/bin/bash
set -e

DIST="dist/pierakladnia"

echo "==> Cleaning previous build..."
rm -rf "$DIST"
mkdir -p "$DIST/bin"

echo "==> Building binaries (linux/arm64)..."
GOOS=linux GOARCH=arm64 go build -o "$DIST/bin/pierakladnia" ./cmd/app
GOOS=linux GOARCH=arm64 go build -o "$DIST/bin/pierakladnia-admin" ./cmd/admin

echo "==> Copying templates, static, migrations, config..."
cp -r web/ "$DIST/web/"
cp -r migrations/ "$DIST/migrations/"
cp -r config/ "$DIST/config/"
cp -r scripts/migrate_up.sh "$DIST/"

echo ""
echo "Build complete! Output in $DIST/"
echo ""
echo "Contents:"
find "$DIST" -type f | sort
