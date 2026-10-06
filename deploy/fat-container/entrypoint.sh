#!/bin/sh
set -e

# Setup PostgreSQL Data Directory
if [ -z "$(ls -A /var/lib/postgresql/data)" ]; then
    echo "Initializing PostgreSQL Database..."
    chown -R postgres:postgres /var/lib/postgresql/data
    su -s /bin/sh postgres -c "initdb -D /var/lib/postgresql/data"
    
    echo "host all all 127.0.0.1/32 trust" >> /var/lib/postgresql/data/pg_hba.conf
    echo "listen_addresses='*'" >> /var/lib/postgresql/data/postgresql.conf
fi

chown -R postgres:postgres /var/lib/postgresql/data
chown -R postgres:postgres /run/postgresql

exec "$@"
