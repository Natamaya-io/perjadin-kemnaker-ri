#!/usr/bin/env bash
set -e

SERVER="gatsu51@100.115.101.14"

echo "========================================================"
echo "🛡️ MEMASANG PERBAIKAN KEAMANAN GOTENBERG (PRIORITAS 2)"
echo "========================================================"
echo "Menghubungkan ke server $SERVER..."

ssh -t "$SERVER" << 'EOF'
  echo "🔍 Menganalisis konfigurasi Nginx..."
  
  # Masalahnya adalah pwa-protokolwamen.gatsu51.com (domain usang)
  # yang mem-proxy port 3001 (sekarang dipakai oleh Gotenberg).
  # Kita akan menghapus symlink domain-domain usang tersebut dari sites-enabled
  # agar Nginx berhenti merespons trafik publik untuk Gotenberg.
  
  echo "🗑️ Menonaktifkan konfigurasi domain lama yang bocor..."
  sudo rm -f /etc/nginx/sites-enabled/pwa-protokolwamen.gatsu51.com
  sudo rm -f /etc/nginx/sites-enabled/api-protokolwamen.gatsu51.com
  sudo rm -f /etc/nginx/sites-enabled/protokolwamen.gatsu51.com
  sudo rm -f /etc/nginx/sites-enabled/pwa-protokol.gatsu51.com
  sudo rm -f /etc/nginx/sites-enabled/api-protokol.gatsu51.com
  sudo rm -f /etc/nginx/sites-enabled/protokol.gatsu51.com
  
  echo "🔄 Me-reload layanan Nginx..."
  sudo systemctl reload nginx
  
  # Opsi tambahan keamanan: Kita pastikan Nginx berjalan normal
  if systemctl is-active --quiet nginx; then
      echo "✅ Nginx berhasil di-reload. Celah Gotenberg telah ditutup!"
  else
      echo "❌ Gagal me-reload Nginx. Silakan periksa status Nginx di server."
  fi
EOF

echo "========================================================"
echo "Proses pengamanan selesai."
echo "========================================================"
