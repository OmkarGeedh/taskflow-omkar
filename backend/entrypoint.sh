#!/bin/sh
set -e

echo "Waiting for postgres..."
until pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER"; do
  sleep 1
done

echo "Running migrations..."
psql "$DATABASE_URL" -f /migrations/20250414_user_table_up.sql
psql "$DATABASE_URL" -f /migrations/20250414_projects_table_up.sql
psql "$DATABASE_URL" -f /migrations/20250414_task_table_up.sql

echo "Seeding database..."
psql "$DATABASE_URL" -f /migrations/seed.sql

echo "Starting API..."
exec ./taskflow
