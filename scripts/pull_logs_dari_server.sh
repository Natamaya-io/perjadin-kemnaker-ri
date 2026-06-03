#!/usr/bin/env bash
set -e

# Konfigurasi
SERVER="gatsu51@100.115.101.14"
REMOTE_DIR="~/perjadin_scripts"
LOCAL_LOG_DIR="$HOME/Proyek/perjadin_kemnaker_ri/log"

echo "========================================================"
echo "🔄 MEMULAI PROSES AUTOMATIC DIAGNOSTIC END-TO-END"
echo "========================================================"

echo ""
echo "🚀 [1/4] Terhubung ke Server untuk pembersihan dan eksekusi skrip..."
# Menjalankan perintah remote via SSH. (Akan meminta password server Anda)
ssh "$SERVER" << EOF
  cd $REMOTE_DIR
  
  echo "🧹 (Remote) Membersihkan folder perjadin_scripts lama (kecuali log_puller.sh)..."
  # Menghapus semua file dan folder kecuali log_puller.sh
  find . -mindepth 1 -maxdepth 1 ! -name 'log_puller.sh' -exec rm -rf {} +
  
  echo "🔥 (Remote) Mengeksekusi log_puller.sh..."
  ./log_puller.sh
EOF

echo ""
echo "🧹 [2/4] Membersihkan direktori log lokal lama..."
mkdir -p "$LOCAL_LOG_DIR"
rm -rf "${LOCAL_LOG_DIR:?}"/*

echo ""
echo "📥 [3/4] Menyedot file diagnostik terbaru (.tar.gz) dari server..."
# Mengambil file tar.gz dari server
scp "$SERVER:$REMOTE_DIR/diagnostics_dump_*.tar.gz" "$LOCAL_LOG_DIR/"

echo ""
echo "📦 [4/4] Mengekstrak file arsip lokal agar siap dibaca..."
# Mengekstrak file tar.gz ke dalam folder log/
cd "$LOCAL_LOG_DIR"
tar -xzf diagnostics_dump_*.tar.gz
rm -f diagnostics_dump_*.tar.gz

echo ""
echo "========================================================"
echo "✅ SUKSES MUTLAK! Proses remote dan sinkronisasi lokal selesai."
echo "📂 Data Anda yang paling mutakhir dan terekstrak rapi ada di:"
echo "   $LOCAL_LOG_DIR"
echo "========================================================"
