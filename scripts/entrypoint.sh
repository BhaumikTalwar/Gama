#!/bin/bash
set -e

echo "Waiting for database..."
until pg_isready -h "$APP_POSTGRES_HOST" -p "$APP_POSTGRES_PORT" -U "$APP_POSTGRES_USER"; do
  echo "Database is unavailable - sleeping"
  sleep 1
done

echo "Database is up - running migrations"
goose -dir migrations postgres "$APP_POSTGRES_URI" up

echo "Setting up MinIO buckets..."
mc alias set local "$APP_S3_ENDPOINT" "$APP_S3_ACCESS_KEY" "$APP_S3_SECRET_KEY" 2>/dev/null || true
mc mb local/"$APP_S3_PUBLIC_BUCKET" --ignore-existing 2>/dev/null || true
mc mb local/"$APP_S3_PRIVATE_BUCKET" --ignore-existing 2>/dev/null || true
mc anonymous set download local/"$APP_S3_PUBLIC_BUCKET" 2>/dev/null || true

echo "Migrations completed - starting application"
exec /app/gama server
