#!/bin/bash
set -e

echo "Stopping Go Web Server..."
if lsof -ti:8080 > /dev/null; then
    lsof -ti:8080 | xargs kill -9
    echo "Web server stopped."
else
    echo "Web server not running on port 8080."
fi

echo "Stopping MySQL container..."
cd "$(dirname "$0")/../deploy/docker"
docker compose stop

echo "Development environment stopped."
