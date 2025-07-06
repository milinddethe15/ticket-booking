#!/bin/bash
set -e

# Run migrations in order
echo "Running database migrations..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" < /docker-entrypoint-initdb.d/migrations/001_initial_schema.up.sql
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" < /docker-entrypoint-initdb.d/migrations/002_add_locked_status.up.sql

# Load sample data
echo "Loading sample data..."
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" < /docker-entrypoint-initdb.d/scripts/sample_data.sql

echo "Database initialization complete!" 