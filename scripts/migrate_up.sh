#!/bin/bash
set -e

# Assuming golang-migrate is installed
MIGRATE="migrate"
if ! command -v $MIGRATE &> /dev/null
then
    echo "golang-migrate could not be found, please install it: https://github.com/golang-migrate/migrate"
    exit 1
fi

MYSQL_DSN="mysql://app:app_password@tcp(127.0.0.1:3306)/gob_trans"

echo "Running migrations..."
$MIGRATE -path migrations -database "$MYSQL_DSN" up
echo "Migrations applied."
