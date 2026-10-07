# Sistem Informasi Perjalanan Dinas & GUP (Kemnaker RI)

![Status](https://img.shields.io/badge/Status-Production-success) ![Tech Stack](https://img.shields.io/badge/Stack-SvelteKit%20%7C%20Golang%20%7C%20PostgreSQL-blue) ![License](https://img.shields.io/badge/License-Restricted-red)

Sistem Informasi Perjalanan Dinas (Perjadin) dan Ganti Uang Persediaan (GUP) adalah platform *Enterprise Resource Planning* (ERP) yang dirancang khusus untuk memenuhi standar birokrasi dan administrasi keuangan di lingkungan **Kementerian Ketenagakerjaan Republik Indonesia (Kemnaker RI)**.

Aplikasi ini mengotomatiskan seluruh siklus birokrasi: mulai dari pengajuan Surat Tugas, kalkulasi Standar Biaya Masukan (SBM), pembuatan Surat Pertanggungjawaban (SPJ) secara otomatis (*pixel-perfect PDF*), hingga manajemen pencairan anggaran (GUP/LS) dan klasifikasi intensi (*Intent Classification*) menggunakan teknologi *Chatbot AI* mutakhir.

---

## 🏗️ Arsitektur & Tumpukan Teknologi

Aplikasi ini menggunakan pendekatan arsitektur **Fat Container Monolith** demi keandalan, determinisme, dan kemudahan deployment di lingkungan *on-premise* maupun VPS tunggal (Bare-Metal).

- **Frontend:** SvelteKit, SolidJS, TailwindCSS, Ark UI.
- **Backend:** Golang (Echo Framework, GORM).
- **Database & Cache:** PostgreSQL 16 & Redis 8.0.
- **Document Engine:** Gotenberg (berbasis Debian/LibreOffice 26.8) untuk *rendering* PDF yang *pixel-perfect*.

---

## 🔄 Alur Kerja Utama Aplikasi (Business Flow)

Aplikasi ini memodelkan alur birokrasi nyata di pemerintahan. Berikut adalah siklus hidup data dalam sistem ini:

1. **Inisiasi (Pengajuan Perjalanan):** Pegawai atau staf protokol membuat pengajuan perjalanan dinas, menentukan destinasi, tanggal berangkat/kembali, dan daftar peserta.
2. **Kalkulasi & SBM (Rekapitulasi):** Sistem secara otomatis menghitung *Per Diem* (uang harian), uang representasi, batas biaya hotel, dan tiket pesawat berdasarkan **Standar Biaya Masukan (SBM)** yang disesuaikan dengan Golongan dan Jabatan peserta.
3. **Penyusunan SPJ (Laporan Perjadin):** Setelah perjalanan selesai, sistem merakit data secara otomatis ke dalam templat `.docx` resmi dan mengonversinya menjadi dokumen cetak (PDF) seperti Kuitansi, Rincian Biaya, dan Laporan Hasil Perjalanan menggunakan Gotenberg.
4. **Pencairan Anggaran (GUP & LS):** Tagihan atau biaya yang sudah dikeluarkan direkapitulasi dan diajukan penggantiannya ke bendahara melalui mekanisme **GUP (Ganti Uang Persediaan)** atau **LS (Langsung)** untuk pembayaran pihak ketiga.
5. **Realokasi & Sinkronisasi (Integrasi MAK):** Semua transaksi keuangan ditautkan ke **Mata Anggaran Kegiatan (MAK)** untuk memastikan penyerapan anggaran seimbang dan tidak terjadi pagu minus.

---

## 📱 Panduan Komprehensif Fitur Sidebar

Penjabaran modul berdasarkan hak akses (RBAC):

### 1. Dashboard (`/dashboard`)
- **Fungsi:** Pusat kendali (*Command Center*) aplikasi.
- **Kegunaan:** Menampilkan analitik tingkat tinggi (*high-level analytics*), grafik penyerapan anggaran, jumlah perjalanan dinas yang sedang berlangsung, *timeline* persetujuan, dan *quick actions*. Pengguna dapat dengan cepat melihat tugas apa yang membutuhkan atensi mereka (misalnya SPJ yang belum ditandatangani).
- **Akses:** Super Admin, Kasubag, Protokol.

---

### 2. Perjalanan Dinas (Perjadin)

#### A. Pengajuan Perjalanan (`/dashboard/pengajuan`)
- **Fungsi:** Formulir entri utama untuk menjadwalkan tugas kedinasan.
- **Kegunaan:** 
  - Membuat Nomor Surat Tugas secara berurutan.
  - Memasukkan data tujuan (Dalam Kota / Luar Kota).
  - Mengelola daftar peserta (Nama, NIP, Golongan, Jabatan).
  - Mengatur jadwal *itinerary* (Keberangkatan & Kepulangan).
- **Akses:** Super Admin, Kasubag, Protokol.

#### B. Manajemen Perdin Admin (`/dashboard/admin/perdin`)
- **Fungsi:** Pusat rekapitulasi, verifikasi, dan mesin hitung otomatis berbasis *Standar Biaya Masukan* (SBM) Kementerian Keuangan.
- **Kegunaan:** 
  - Melakukan tinjauan menyeluruh (*override* admin) terhadap semua pengajuan dinas.
  - Sistem otomatis mengekstrak tarif uang harian berdasarkan provinsi tujuan.
  - Menghitung tarif penginapan maksimal sesuai eselon/golongan.
  - Membantu bendahara mencadangkan/memblokir pagu anggaran yang diperlukan sebelum uang benar-benar dicairkan.
- **Akses:** Super Admin, Kasubag.

#### C. Laporan Perjadin (`/dashboard/laporan`)
- **Fungsi:** Generator dokumen administratif birokrasi (Cetak PDF).
- **Kegunaan:**
  - Mengubah data JSON yang diinput menjadi formulir resmi pemerintahan (*pixel-perfect*).
  - **Mencetak Kuitansi Rampung:** Lengkap dengan terbilang (huruf) secara otomatis.
  - **Mencetak Rincian Biaya Perjalanan Dinas:** Tabel rincian pengeluaran per peserta.
  - **Mencetak Surat Pernyataan Tanggung Jawab Mutlak (SPTJM).**
  - Menggabungkan seluruh bukti *scan* struk/tiket pesawat (Evidence) menjadi satu PDF utuh untuk diserahkan ke pemeriksa (BPK/Inspektorat).
- **Akses:** Super Admin, Kasubag, Protokol.

---

### 3. Ganti Uang Persediaan (GUP)

#### A. Pengajuan GUP (`/dashboard/gup/pengajuan`)
- **Fungsi:** Pembuatan keranjang/batch tagihan yang akan diajukan penggantian dananya.
- **Kegunaan:** Ketika bendahara telah menalangi biaya Perjalanan Dinas menggunakan Uang Persediaan (UP), tagihan-tagihan SPJ tersebut dikelompokkan ke dalam satu "Pengajuan GUP". Di sini, pengguna mendaftarkan SPJ mana saja yang akan di-*reimburse* kepada KPPN (Kantor Pelayanan Perbendaharaan Negara).
- **Akses:** Super Admin, Kasubag.

#### B. Laporan GUP (`/dashboard/gup/laporan`)
- **Fungsi:** Laporan eksekutif terkait sisa Uang Persediaan.
- **Kegunaan:**
  - Melacak status setiap pengajuan GUP (Draft, Diajukan, Cair).
  - Mengekspor Buku Kas Umum (BKU) dan laporan pertanggungjawaban bendahara.
  - Memastikan *cash flow* kas negara di instansi tetap sehat dan akuntabel.
- **Akses:** Super Admin, Kasubag.

#### C. Pembayaran Langsung / LS (`/dashboard/gup/ls`)
- **Fungsi:** Manajemen Pembayaran Langsung.
- **Kegunaan:** Digunakan untuk transaksi perjalanan dinas atau konsinyering (Rapat di Luar Kantor) yang nominalnya melebihi limit Uang Persediaan, sehingga pembayaran ditransfer langsung dari Rekening Kas Negara ke rekening pihak ketiga (misalnya: Hotel, Maskapai, atau Event Organizer).
- **Akses:** Super Admin, Kasubag.

#### D. Integrasi MAK (`/dashboard/gup/integrasi`)
- **Fungsi:** Sinkronisasi pembebanan anggaran.
- **Kegunaan:**
  - Setiap SPJ wajib dibebankan pada satu MAK spesifik (contoh: *524111 - Belanja Perjalanan Dinas Biasa*).
  - Sistem mencatat pagu awal, pagu terpakai, dan sisa pagu.
  - Memblokir pembuatan SPJ jika MAK yang dituju sudah melewati batas (Pagu Minus).
- **Akses:** Super Admin, Kasubag.

---

### 4. AI (Kecerdasan Buatan)

#### A. Chatbot AI (`/dashboard/chatbot`)
- **Fungsi:** Asisten Virtual Birokrasi & Pencarian Cerdas.
- **Kegunaan:**
  - **Navigasi Sistem (Command Hub):** Alih-alih mencari menu secara manual, pengguna dapat berinteraksi dengan AI untuk memunculkan aksi spesifik.
  - **Tanya Jawab Regulasi SBM:** Pengguna dapat bertanya seputar regulasi SBM yang valid di dalam sistem.
- **Akses:** Super Admin, Kasubag.

---

### 5. Zona Admin

#### A. Manajemen Pengguna (`/dashboard/admin/users`)
- **Fungsi:** Manajemen Akun dan *Role-Based Access Control* (RBAC).
- **Kegunaan:**
  - Membuat, mengedit, atau menonaktifkan akun pegawai (Protokol, Bendahara, Kasubag, Admin).
  - Melakukan *reset password*.
  - Mengelola data profil, Nomor Induk Pegawai (NIP), dan jabatan struktural yang akan tercetak di lembar SPJ.
- **Akses:** Super Admin.

#### B. Pengaturan Sistem (`/dashboard/admin/settings`)
- **Fungsi:** Pusat konfigurasi sistem, parameter global, dan infrastruktur.
- **Kegunaan:**
  - Memperbarui tabel tarif SBM tahunan tanpa harus mengubah *source code*.
  - Menyetel variabel lingkungan dinamis (*environment variables*).
  - Manajemen *backup* dan koneksi sinkronisasi ke modul pihak ketiga.
- **Akses:** Super Admin.

---

## 🔒 Security & Deployment

Proyek ini telah menerapkan protokol pengamanan kelas produksi (*production-grade*):
1. **Zero-Trust Network:** Aplikasi hanya dapat dijangkau dari luar melalui Cloudflare Tunnels (Edge Proxy) yang difilter dan via Tailscale VPN untuk SSH server internal.
2. **Containerization Strict Boundaries:** Seluruh servis (PostgreSQL, Go, Redis, Svelte) dibungkus dalam **Satu Kontainer Raksasa (Fat Container)** menggunakan Debian 13 base (Gotenberg). Ini memecahkan fenomena *Render Engine Desync*—menjamin bahwa cetakan PDF di server akan identik 100% (*pixel-perfect*) dengan cetakan saat proses *development* lokal.
3. **Immutability:** Modifikasi variabel *environment* dan rahasia aplikasi diinjeksi sepenuhnya oleh **Doppler CLI**. Berkas `.env` dilarang hadir di dalam server produksi maupun lokal.

### ⚙️ Penggunaan Doppler (Service Token)
Untuk menjalankan proyek ini, Anda **wajib** menggunakan Doppler Service Token untuk menginjeksi rahasia dan konfigurasi lingkungan.

1. **Setup Service Token:**
   Pastikan Doppler CLI sudah terinstal, lalu masukkan *Service Token* dari proyek `perjadin-kemnaker-ri`:
   ```bash
   export DOPPLER_TOKEN="dp.st.prd.[YOUR_TOKEN]"
   ```

2. **Eksekusi Aplikasi:**
   Jalankan perintah aplikasi dengan membungkusnya menggunakan `doppler run`:
   ```bash
   # Backend (Go)
   doppler run -- go run ./cmd/api
   
   # Frontend (SvelteKit)
   doppler run -- npm run dev
   ```
   *Dengan cara ini, seluruh variabel rahasia akan diinjeksi langsung ke dalam memori RAM aplikasi tanpa pernah menyentuh disk/file system.*

*(Dokumen ini merupakan properti tertutup / restricted dan dirancang eksklusif untuk panduan pengembangan dan operasional internal kemnaker-ri).*
