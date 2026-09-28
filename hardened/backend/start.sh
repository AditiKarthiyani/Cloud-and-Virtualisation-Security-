#!/bin/sh
set -e

# PostgreSQL now runs in a seperate container ( see docker-compose.yml)

# Wait for PostgreSQL to become available before starting the Go backend
until pg_isready -h "$DB_HOST" -p "$DB_PORT"  >/dev/null 2>&1; do
    echo "Waitng for database "
    sleep 1
done

exec  /server
