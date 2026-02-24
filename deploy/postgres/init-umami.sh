#!/bin/bash
set -e

# Create the umami database and user for analytics.
# This script runs only on first init (empty data volume).

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'umami') THEN
            CREATE ROLE umami WITH LOGIN PASSWORD '${UMAMI_DB_PASSWORD}';
        END IF;
    END
    \$\$;

    SELECT 'CREATE DATABASE umami OWNER umami'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'umami')\gexec

    GRANT ALL PRIVILEGES ON DATABASE umami TO umami;
EOSQL
