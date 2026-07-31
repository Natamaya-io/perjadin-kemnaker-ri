#!/bin/bash
set -e

echo "Starting Sync from Production to Staging..."

# Extract env vars from containers
PROD_DB_USER=$(podman exec perjadin-db printenv POSTGRES_USER)
PROD_DB_NAME=$(podman exec perjadin-db printenv POSTGRES_DB)

STG_DB_USER=$(podman exec perjadin-stg-db printenv POSTGRES_USER)
STG_DB_NAME=$(podman exec perjadin-stg-db printenv POSTGRES_DB)
STG_DB_PASS=$(podman exec perjadin-stg-db printenv POSTGRES_PASSWORD)

echo "Dumping Production Database..."
podman exec perjadin-db pg_dump -U "$PROD_DB_USER" "$PROD_DB_NAME" -F c -f /tmp/prod_dump.backup

echo "Copying dump from Prod container to Host..."
podman cp perjadin-db:/tmp/prod_dump.backup /tmp/prod_dump.backup

echo "Copying dump from Host to Staging container..."
podman cp /tmp/prod_dump.backup perjadin-stg-db:/tmp/prod_dump.backup

echo "Resetting Staging Database..."
# Terminate existing connections
podman exec -e PGPASSWORD="$STG_DB_PASS" perjadin-stg-db psql -U "$STG_DB_USER" -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$STG_DB_NAME';" || true
podman exec -e PGPASSWORD="$STG_DB_PASS" perjadin-stg-db psql -U "$STG_DB_USER" -d postgres -c "DROP DATABASE IF EXISTS $STG_DB_NAME;"
podman exec -e PGPASSWORD="$STG_DB_PASS" perjadin-stg-db psql -U "$STG_DB_USER" -d postgres -c "CREATE DATABASE $STG_DB_NAME OWNER $STG_DB_USER;"

echo "Restoring Database to Staging..."
podman exec -e PGPASSWORD="$STG_DB_PASS" perjadin-stg-db pg_restore -U "$STG_DB_USER" -d "$STG_DB_NAME" --no-owner -1 /tmp/prod_dump.backup


echo "Restarting Staging Backend & Redis..."
systemctl --user restart perjadin-stg-backend.service perjadin-stg-redis.service

echo "Cleaning up temp files..."
podman exec perjadin-db rm /tmp/prod_dump.backup
podman exec perjadin-stg-db rm /tmp/prod_dump.backup
rm /tmp/prod_dump.backup

echo "Done! Staging is now safely synced with Production."
