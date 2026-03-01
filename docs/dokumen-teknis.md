Modul Master Data
Berfungsi untuk menyimpan data dasar yang digunakan seluruh sistem.
Sub-modul:
•	Master Pegawai (Nama, nomor telepon, NIP)
•	Master User (Protokol, Super Admin, Kasubbag)
•	Master Tempat (Provinsi, Zona Dalam/Luar Kota)
•	Standar Biaya Masukan (SBM)
•	Data Transportasi (Pesawat, Kereta, dll.)
•	Role & Hak Akses User

Modul Mendelegasikan Perjalanan Dinas
Digunakan Super Admin/Kassubag untuk mendelegasikan para Protokol 
Fitur:
•	Form pengajuan (Lokasi dinas, Tanggal Mulai/Selesai, dalam rangka, nama yang bertugas)
•	Upload dokumen pendukung (Surat tugas)
•	Estimasi biaya otomatis (berdasarkan SBM)
•	Draft & Submit
•	Tracking status pengajuan
Output:
•	Notifikasi kepada Protokol yang di tugaskan
•	Estimasi RAB Perjadin










HAK AKSES APLIKASI PERJADIN PROTOKOL
User/Protokol
•	Dashboard 
-	Notifikasi tugas terbaru  (berbentuk button card)
Menampilkan Surat Tugas, Id SPD, Nama Petugas, Tujuan, Tanggal mulai/selesai
-	Table tugas (Id SPD, Nama, tujuan, tanggal mulai/selesai) dengan status Assigned, In Progress, Completed.
Note :
Assigned  = Status ketika diberikan penugasan oleh Superadmin/Kasubbag.
In Progress  = Status berubah menjadi In Progress ketika sudah selesai Penugasan (dari tanggal perjadin yang telah ditetapkan)
Completed  = Status berubah menjadi Completed ketika Perjadin sudah di bayarkan (Status di ubah oleh Superadmin/Kasubbag)

-	Kalender notifikasi tugas (berbentuk kalender 1 bulan, dengan tanda merah untuk tugas yang diberikan pada tanggal tersebut)
•	Laporan Perjadin (Setelah diajukan oleh Kasubag atau Super Admin)
-	Menu Table tugas dan status seperti pada Dashboard table (sertakan button Review dan Hapus).
Note :
*Tampilan Review : Berisi Form Tambah Perjadin yang sudah terisi oleh User/Protokol. (Tambahkan aksi/button Edit dan Print)

-	Button Tambah Perjadin :
o	Input Surat Tugas *pdf
o	Input SPPD (lembar 12) *pdf
o	Laporan perjadin (template surat yang diberikan oleh Mba Nurin, dengan dokumentasi di dalamnya)
o	Input Boarding Pass *pdf,  *jpeg
o	Input Tiket Pergi (sertakan kolom nominal) dan input kwitansi *pdf, *jpeg
o	Input Tiket Pulang (sertakan kolom nominal) dan input kwitansi *pdf, *jpeg
o	Input Bill Hotel (sertakan kolom nominal) dan input kwitansi *pdf, *jpeg
Note: Tambahkan kolom kali kan hari.
o	Input add cost (sertakan kolom nominal) dan input kwitansi *pdf, *jpeg
Note: Tambahkan aksi/button untuk menambahkan kolom add cost 

Output :
- Tampil di Menu table tugas dengan status dikerjakan.
- Tampil di Menu SPJ role Superadmin (dengan nominal sudah terinput)


•	Menu Dokumen
-	Template surat-suratan (isi nya menyesuaikan dan didiskusikan terlebih dahulu)



Super Admin/Pembendaharaan
•	Dashboard
-	Card Tugas pending, card tugas Selesai, card total dinas luar kota, card total dinas dalam kota
-	Pie Chart
-	Kalender Bulanan


•	Menu Pengajuan Perjadin
-	Tabel Pengajuan Perjadin ( Id SPD, tujuan, lokasi, SBM Provinsi, status )
Button Review dan Hapus
-	Button Tambah Pengajuan Perjadin :
o	Form Pengajuan
- Nama Kegiatan *Textfield
- Tujuan *Textfield
- Tanggal mulai/Selesai *datepicker
- Nama yang bertugas *dropdown nama Protokol
*Button tambah Nama Petugas, apabila petugas lebih dari satu
- Stakeholders *dropdown (Menaker, Wamenaker, Pendamping Menaker, Pendamping Wamenaker)
- Detail Agenda *textbox
o	Input dokumen pendukung (Surat Tugas) *pdf
o	Input Kota dan Provinsi (untuk menentukan nilai SBM) *dropdown
Note :
Logic total hari (datepicker) sudah otomatis terkalkulasi dengan SBM per Provinsi.

Output :
- Tampil di menu Tabel pengajuan Perjadin (dengan button Review dan Hapus).
Note :
Review, menampilkan Modal Form Pengajuan yang sudah terinput. (tambahkan No. SPD kepada masing-masing petugas protokol).
- Tampil Notifikasi Whatsapp ke petugas protokol yang di tugaskan.
- Tampil di Menu Dashboard Aktor User/Protokol.

•	Menu SPJ (Surat Pertanggungjawaban)
-	Tabel Pengisian SPJ menampilkan laporan perjadin yang telah di input oleh User/Protokol.
-	Tabel SPJ (menampilkan Id SPD, No SPD, Nama Petugas, Lokasi dinas) button: Review, hapus, dan cetak kwitansi. 
Note : Review menampilkan halaman Laporan perjadin dari User Protokol, dengan fitur button edit dan cetak. 

