#!/bin/sh
set -e

# Discover Debian PostgreSQL version and set paths
PG_VERSION=$(ls /usr/lib/postgresql/ 2>/dev/null | sort -V | tail -1)
PG_BIN="/usr/lib/postgresql/${PG_VERSION}/bin"
export PATH="${PG_BIN}:${PATH}"

PGDATA="/var/lib/postgresql/data"

# Setup PostgreSQL Data Directory
if [ -z "$(ls -A ${PGDATA} 2>/dev/null)" ]; then
    echo "Initializing PostgreSQL Database..."
    mkdir -p ${PGDATA}
    chown -R postgres:postgres ${PGDATA}
    gosu postgres ${PG_BIN}/initdb -D ${PGDATA}
    
    # Allow local connections
    echo "host all all 127.0.0.1/32 trust" >> ${PGDATA}/pg_hba.conf
    echo "listen_addresses='*'" >> ${PGDATA}/postgresql.conf
fi

chown -R postgres:postgres ${PGDATA}
mkdir -p /run/postgresql
chown -R postgres:postgres /run/postgresql

exec "$@"
