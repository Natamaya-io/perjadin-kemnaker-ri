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

echo "Running Sanitization (Anonymization)..."
# Use a valid bcrypt hash for "12345678"
podman exec -e PGPASSWORD="$STG_DB_PASS" perjadin-stg-db psql -U "$STG_DB_USER" -d "$STG_DB_NAME" -c "
UPDATE users 
SET password = '\$2a\$10\$w09ZlOqC8a6pXlHkG8K/Q.27Jt.O/2Lz2.6gZq0qM/G6I0N/X0UeC', 
    email = id || '@demo.local', 
    name = 'Demo User ' || substr(id::text, 1, 8), 
    nip = substr(md5(random()::text), 1, 18)
WHERE role != 'super_admin';
"

echo "Restarting Staging Backend & Redis..."
systemctl --user restart perjadin-stg-backend.service perjadin-stg-redis.service

echo "Cleaning up temp files..."
podman exec perjadin-db rm /tmp/prod_dump.backup
podman exec perjadin-stg-db rm /tmp/prod_dump.backup
rm /tmp/prod_dump.backup

echo "Done! Staging is now safely synced with Production."
