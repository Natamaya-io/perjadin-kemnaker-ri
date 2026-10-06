#!/bin/sh
# Wrapper to start PostgreSQL with correct Debian-versioned binary path
PG_VERSION=$(ls /usr/lib/postgresql/ 2>/dev/null | sort -V | tail -1)
exec gosu postgres /usr/lib/postgresql/${PG_VERSION}/bin/postgres -D /var/lib/postgresql/data
