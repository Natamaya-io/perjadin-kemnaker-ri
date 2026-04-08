#!/bin/bash
set -eo pipefail

# ==========================================================
# 🚀 aaPanel Webhook Deployment Script - OPTIMIZED
# ==========================================================

# --- 1. Konfigurasi Environment Dasar ---
export HOME="/root"
export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"

PROJECT_DIR="${PROJECT_DIR:-/www/wwwroot/staging.lium.site}"
cd "$PROJECT_DIR"

echo "=========================================================="
echo "📅 Date: $(date)"
echo "🚀 Deployment Staging: $PROJECT_DIR"
echo "=========================================================="

# --- 2. Aggressive Cleanup ---
echo "[+] Membersihkan sisa container dan file lama..."
# Hapus file compose lama agar tidak ada konflik versi
rm -f compose.yml docker-compose.yml

# Matikan layanan dan hapus container (tanpa menghapus volume data utama)
docker compose down --remove-orphans 2>/dev/null || true
docker rm -f perjadin_redis perjadin_db_container perjadin_backend perjadin_frontend perjadin_gotenberg 2>/dev/null || true

# --- 3. Load & Ekstrak Kredensial (Doppler) ---
SECRETS_FILE="/root/.perjadin-secrets"
if [ ! -f "$SECRETS_FILE" ]; then
    echo "❌ [ERROR] File $SECRETS_FILE tidak ditemukan!"
    exit 1
fi

source "$SECRETS_FILE"
export DOPPLER_TOKEN="$DOPPLER_TOKEN"

echo "[+] Mengambil kredensial terbaru dari Doppler..."
# Menggunakan satu kali panggil doppler secrets download untuk efisiensi jika perlu, 
# tapi tetap mempertahankan gaya 'get' agar aman dengan karakter khusus.
DB_PWD=$(doppler secrets get DB_PASSWORD --plain | tr -d '\r\n ')
DB_USR=$(doppler secrets get DB_USER --plain | tr -d '\r\n ')
DB_NAM=$(doppler secrets get DB_NAME --plain | tr -d '\r\n ')
RD_PWD=$(doppler secrets get REDIS_PASSWORD --plain | tr -d '\r\n ')
JWT_SECRET=$(doppler secrets get JWT_SECRET --plain | tr -d '\r\n ')
FONNTE_API=$(doppler secrets get FONNTE_API --plain | tr -d '\r\n ')
PUBLIC_API_URL=$(doppler secrets get PUBLIC_API_URL --plain | tr -d '\r\n ')

# --- 4. Generate Fresh docker-compose.yml ---
APP_REPO="${REPO_NAME:-raxbyte-org/perjadin-kemnaker-ri}"
APP_TAG="${IMAGE_TAG:-staging}"

echo "[+] Membuat file docker-compose.yml baru..."
cat << EOF > docker-compose.yml
services:
  db:
    image: postgres:16-alpine
    container_name: perjadin_db_container
    environment:
      POSTGRES_USER: "${DB_USR}"
      POSTGRES_PASSWORD: "${DB_PWD}"
      POSTGRES_DB: "${DB_NAM}"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    container_name: perjadin_redis
    ports:
      - "6380:6379"
    volumes:
      - redis_data:/data
    command: redis-server --requirepass "${RD_PWD}"
    restart: unless-stopped

  backend:
    image: ghcr.io/${APP_REPO}-backend:${APP_TAG}
    container_name: perjadin_backend
    environment:
      DB_HOST: "db"
      DB_PORT: "5432"
      DB_USER: "${DB_USR}"
      DB_PASSWORD: "${DB_PWD}"
      DB_NAME: "${DB_NAM}"
      DB_SSLMODE: "disable"
      JWT_SECRET: "${JWT_SECRET}"
      REDIS_PASSWORD: "${RD_PWD}"
      REDIS_HOST: "redis"
      REDIS_PORT: "6379"
      FONNTE_API: "${FONNTE_API}"
      SEED_DB: "true"
      PPK_NAME: "Arief Hafidiyanto"
      PPK_NIP: "19720827 200312 1 002"
      APP_PORT: "8081"
      GOTENBERG_URL: "http://gotenberg:3000"
    ports:
      - "8081:8081"
    depends_on:
      - db
      - redis
    volumes:
      - backend_uploads:/app/uploads
    restart: unless-stopped

  frontend:
    image: ghcr.io/${APP_REPO}-frontend:${APP_TAG}
    container_name: perjadin_frontend
    environment:
      INTERNAL_API_URL: "http://backend:8081"
      VITE_API_URL: "${PUBLIC_API_URL:-https://staging.lium.site}/api/v1"
      VITE_USE_REAL_API: "true"
    ports:
      - "3005:3000"
    depends_on:
      - backend
    restart: unless-stopped

  gotenberg:
    image: gotenberg/gotenberg:8
    container_name: perjadin_gotenberg
    shm_size: "2g"
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
  backend_uploads:
EOF

# --- 5. Deploy ---
echo "[+] Menarik image terbaru dari GHCR..."
docker compose pull

echo "[+] Menjalankan layanan (Force Recreate)..."
docker compose up -d --force-recreate

# --- 6. Robust Healthcheck & Database Cleanup ---
echo "[+] Menunggu Database siap..."
MAX_RETRIES=30
RETRY_COUNT=0
DB_READY=false

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
   if docker exec perjadin_db_container pg_isready -U "${DB_USR}" -d "${DB_NAM}" >/dev/null 2>&1; then
      DB_READY=true
      break
   fi
   echo "    ⏳ Database sedang inisialisasi... ($((RETRY_COUNT+1))/$MAX_RETRIES)"
   sleep 2
   RETRY_COUNT=$((RETRY_COUNT+1))
done

if [ "$DB_READY" = true ]; then
   echo "[+] Database OK."
   
   # FIX: Hapus sisa akun demo lama jika SEED_DB diaktifkan agar tidak menumpuk
   echo "[+] Membersihkan tabel user untuk sinkronisasi data demo..."
   docker exec perjadin_db_container psql -U "${DB_USR}" -d "${DB_NAM}" -c "TRUNCATE TABLE users CASCADE;" || echo "⚠️ Gagal membersihkan tabel users (mungkin tabel belum ada)."
   
   # Reset dirty flag migration jika terjadi kegagalan sebelumnya
   echo "[+] Membersihkan dirty flag pada schema_migrations..."
   docker exec perjadin_db_container psql -U "${DB_USR}" -d "${DB_NAM}" -c "UPDATE schema_migrations SET dirty = false;" || echo "⚠️ Skip dirty flag."
   
   # Restart backend agar proses Seeding berjalan di database yang sudah bersih
   echo "[+] Merestart backend untuk memicu seeding ulang..."
   docker restart perjadin_backend
else
   echo "❌ [ERROR] Database gagal inisialisasi dalam waktu yang ditentukan."
   docker logs --tail 20 perjadin_db_container
   exit 1
fi

# --- 7. Final Cleanup ---
echo "[+] Membersihkan resource Docker yang tidak terpakai..."
docker image prune -af --filter "until=12h"
docker container prune -f
docker network prune -f

echo "✅ Deployment Selesai!"
