#!/bin/sh
set -e

# Setup PostgreSQL Data Directory
if [ -z "$(ls -A /var/lib/postgresql/data)" ]; then
    echo "Initializing PostgreSQL Database..."
    chown -R postgres:postgres /var/lib/postgresql/data
    su-exec postgres initdb -D /var/lib/postgresql/data
    
    # Start postgres temporarily to create user/db if needed, but since our app uses "postgres" user with no password locally, it's fine.
    # By default, local connections via unix socket or 127.0.0.1 map to the postgres user if trust is enabled (default in Alpine initdb).
    echo "host all all 127.0.0.1/32 trust" >> /var/lib/postgresql/data/pg_hba.conf
    echo "listen_addresses='*'" >> /var/lib/postgresql/data/postgresql.conf
fi

chown -R postgres:postgres /var/lib/postgresql/data
chown -R postgres:postgres /run/postgresql

exec "$@"
