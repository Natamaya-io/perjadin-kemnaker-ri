#!/usr/bin/env bash
set -e

SERVER="gatsu51@100.115.101.14"

echo "========================================================"
echo "⚡ MEMASANG PERBAIKAN INDEX PERFORMA (PRIORITAS 1)"
echo "========================================================"
echo "Menghubungkan ke server $SERVER..."

ssh -t "$SERVER" << 'EOF'
  # Mengambil username dan database name asli dari env backend
  REAL_DB_USER=$(podman exec perjadin-backend env | grep 'DB_USER' | cut -d '=' -f 2 | tr -d '\r')
  REAL_DB_NAME=$(podman exec perjadin-backend env | grep 'DB_NAME' | cut -d '=' -f 2 | tr -d '\r')
  
  # Fallback jika env kosong
  if [ -z "$REAL_DB_USER" ]; then REAL_DB_USER="perjadin"; fi
  if [ -z "$REAL_DB_NAME" ]; then REAL_DB_NAME="perjadin"; fi

  echo "🛠️ Menerapkan Index pada Database: $REAL_DB_NAME dengan user: $REAL_DB_USER..."
  
  # Eksekusi raw SQL untuk membuat Expression Index secara permanen
  podman exec -i perjadin-db psql -U "$REAL_DB_USER" -d "$REAL_DB_NAME" -c "
    CREATE INDEX IF NOT EXISTS idx_travel_records_spd_numeric 
    ON travel_records ( (CAST(SUBSTRING(spd_number FROM '[0-9]+') AS INTEGER)) );
  "
  
  if [ $? -eq 0 ]; then
    echo "✅ SUKSES MUTLAK! Index telah terpasang di database production."
  else
    echo "❌ Gagal memasang index."
  fi
EOF

echo "========================================================"
echo "Proses selesai."
echo "========================================================"
