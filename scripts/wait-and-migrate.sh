#!/bin/sh

echo "⏳ Waiting for MySQL on $DB_HOST:$DB_PORT..."

# 等待資料庫連得通
while ! nc -z "$DB_HOST" "$DB_PORT"; do
  sleep 2
  echo "⏳ Still waiting for MySQL..."
done

echo "✅ MySQL is up, running migrations..."
migrate -path /migrations \
  -database "mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}" \
  -verbose \
  -lock-timeout 5 \
  up