-- Skema Database Komprehensif untuk Sistem Perjalanan Dinas (Perjadin) Protokol Kemnaker RI
-- Dialect: PostgreSQL
-- Dibuat oleh: Gemini CLI
-- Tanggal: 1 Maret 2026

-- =================================================================================================
-- 1. CORE & ORGANIZATIONAL STRUCTURE
-- =================================================================================================

-- Unit Kerja / Departemen (Hierarkis)
CREATE TABLE departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE, -- Kode Unit Kerja
    parent_id UUID REFERENCES departments(id), -- Untuk sub-unit
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Pangkat & Golongan (e.g., IV/a, III/d)
CREATE TABLE ranks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL, -- e.g., "Pembina Utama"
    golongan VARCHAR(10) NOT NULL, -- e.g., "IV/e"
    level INTEGER NOT NULL, -- Tingkatan untuk sorting/logic (e.g., 1=Highest)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Jabatan (e.g., Menteri, Eselon I, Staf)
CREATE TABLE positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    echelon VARCHAR(50), -- Eselon I, II, III, IV, Non-Eselon
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Users / Employees (Menggabungkan konsep User login dan Data Pegawai)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    nip VARCHAR(50) UNIQUE, -- Nomor Induk Pegawai (Optional for non-PNS)
    nik VARCHAR(50), -- Nomor Induk Kependudukan
    phone_number VARCHAR(20),
    
    -- Relasi Kepegawaian
    department_id UUID REFERENCES departments(id),
    rank_id UUID REFERENCES ranks(id),
    position_id UUID REFERENCES positions(id),
    
    -- Role & Permissions
    role VARCHAR(50) DEFAULT 'protokol', -- admin, keuangan, ppk, bendahara, protokol, user
    status VARCHAR(20) DEFAULT 'active', -- active, inactive, suspended
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =================================================================================================
-- 2. MASTER DATA & REFERENCE (SBM - Standar Biaya Masukan)
-- =================================================================================================

-- Provinsi
CREATE TABLE ref_provinces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(10) UNIQUE
);

-- Kota / Kabupaten
CREATE TABLE ref_cities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID REFERENCES ref_provinces(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) -- KOTA, KABUPATEN
);

-- Standar Biaya Uang Harian (Perpres SBM)
CREATE TABLE sbm_uh_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID REFERENCES ref_provinces(id),
    city_id UUID REFERENCES ref_cities(id), -- Optional, jika spesifik kota
    
    fullboard_rate DECIMAL(15, 2) NOT NULL, -- Paket Fullboard
    fullname_rate DECIMAL(15, 2) NOT NULL, -- Paket Fullhalfd
    outside_city_rate DECIMAL(15, 2) NOT NULL, -- Luar Kota Biasa
    inside_city_rate DECIMAL(15, 2) NOT NULL, -- Dalam Kota > 8 Jam
    diklat_rate DECIMAL(15, 2) NOT NULL, -- Uang Saku Diklat
    
    year INTEGER NOT NULL, -- Tahun Anggaran SBM
    is_active BOOLEAN DEFAULT TRUE
);

-- Standar Biaya Penginapan (Hotel)
CREATE TABLE sbm_hotel_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID REFERENCES ref_provinces(id),
    
    echelon_1_rate DECIMAL(15, 2), -- Menteri / Pejabat Negara / Eselon I
    echelon_2_rate DECIMAL(15, 2), -- Eselon II
    echelon_3_rate DECIMAL(15, 2), -- Eselon III / Gol IV
    echelon_4_rate DECIMAL(15, 2), -- Eselon IV / Gol III
    staff_rate DECIMAL(15, 2),     -- Gol II / I
    
    year INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

-- Standar Biaya Transport (Tiket Pesawat/Kereta)
CREATE TABLE sbm_transport_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    origin_city_id UUID REFERENCES ref_cities(id),
    destination_city_id UUID REFERENCES ref_cities(id),
    
    transport_mode VARCHAR(50), -- PESAWAT, KERETA, BUS
    business_class_max DECIMAL(15, 2),
    economy_class_max DECIMAL(15, 2),
    
    year INTEGER NOT NULL
);

-- =================================================================================================
-- 3. BUDGETING & FINANCE
-- =================================================================================================

-- Mata Anggaran (DIPA / RKAKL)
CREATE TABLE budget_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL, -- e.g., 524111 (Belana Perjalanan Dinas Biasa)
    name VARCHAR(255) NOT NULL,
    description TEXT,
    fiscal_year INTEGER NOT NULL,
    
    total_budget DECIMAL(15, 2) NOT NULL,
    current_balance DECIMAL(15, 2) NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =================================================================================================
-- 4. TRAVEL MANAGEMENT (PERJADIN)
-- =================================================================================================

-- Surat Perintah Tugas (SPT) - Dokumen Induk
-- Satu SPT bisa untuk banyak pegawai (rombongan)
CREATE TABLE travel_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(100) UNIQUE NOT NULL, -- Nomor Surat Tugas
    
    creator_id UUID REFERENCES users(id), -- Pembuat SPT
    approver_id UUID REFERENCES users(id), -- Penanda Tangan SPT (Pejabat Berwenang)
    
    budget_account_id UUID REFERENCES budget_accounts(id), -- Beban Anggaran
    
    purpose TEXT NOT NULL, -- Maksud Perjalanan Dinas
    transport_type VARCHAR(50), -- Kendaraan Dinas, Umum, Pribadi
    
    departure_date DATE NOT NULL,
    return_date DATE NOT NULL,
    
    origin_city_id UUID REFERENCES ref_cities(id),
    destination_city_id UUID REFERENCES ref_cities(id), -- Tujuan Utama
    
    status VARCHAR(50) DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, APPROVED, REJECTED, CANCELED
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Surat Perjalanan Dinas (SPD) - Dokumen Perorangan
-- Setiap pegawai dalam SPT memiliki satu SPD unik
CREATE TABLE travel_spds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    travel_order_id UUID REFERENCES travel_orders(id) ON DELETE CASCADE,
    
    employee_id UUID REFERENCES users(id), -- Pegawai yang ditugaskan
    spd_number VARCHAR(100) UNIQUE, -- Nomor SPD (Generated saat Approval)
    
    -- Override tanggal jika berbeda dengan SPT (misal pulang lebih awal)
    departure_date DATE,
    return_date DATE,
    
    travel_type VARCHAR(50), -- DALAM_KOTA, LUAR_KOTA, LUAR_NEGERI
    
    -- Status Perorangan
    status VARCHAR(50) DEFAULT 'ISSUED', -- ISSUED, COMPLETED, REPORTED
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Rincian Biaya Riil (Expenditures)
-- Detail pengeluaran untuk setiap SPD
CREATE TABLE travel_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spd_id UUID REFERENCES travel_spds(id) ON DELETE CASCADE,
    
    expense_category VARCHAR(50) NOT NULL, -- UANG_HARIAN, HOTEL, TIKET, TAKSI, REPRESENTASI
    
    description VARCHAR(255), -- Keterangan (e.g., "Garuda Indonesia GA-123")
    
    amount_proposed DECIMAL(15, 2) NOT NULL, -- Biaya yang diajukan
    amount_approved DECIMAL(15, 2), -- Biaya yang disetujui (oleh PPK/Keuangan)
    
    proof_file_url TEXT, -- Link ke bukti (Boarding Pass, Bill Hotel)
    
    is_lumpsum BOOLEAN DEFAULT FALSE, -- Jika TRUE, tidak butuh bukti detail (seperti Uang Harian)
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- =================================================================================================
-- 5. REPORTING & OUTPUT
-- =================================================================================================

-- Laporan Hasil Perjalanan Dinas
CREATE TABLE travel_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spd_id UUID REFERENCES travel_spds(id) UNIQUE, -- Satu SPD satu Laporan
    
    activity_summary TEXT NOT NULL, -- Ringkasan Kegiatan
    issues_identified TEXT, -- Masalah yang ditemukan
    recommendations TEXT, -- Saran/Tindak Lanjut
    
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'SUBMITTED' -- SUBMITTED, REVIEWED, ACCEPTED
);

-- Lampiran Laporan (Foto Kegiatan, dll)
CREATE TABLE report_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES travel_reports(id) ON DELETE CASCADE,
    
    file_url TEXT NOT NULL,
    file_type VARCHAR(50),
    description VARCHAR(255)
);

-- Approval History (Audit Trail)
CREATE TABLE approval_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_id UUID NOT NULL, -- ID of SPT or SPD or Report
    reference_type VARCHAR(50) NOT NULL, -- 'SPT', 'SPD', 'REPORT'
    
    approver_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL, -- APPROVE, REJECT, REQUEST_REVISION
    
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for Performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_travel_orders_dates ON travel_orders(departure_date, return_date);
CREATE INDEX idx_travel_spds_employee ON travel_spds(employee_id);
CREATE INDEX idx_expenses_spd ON travel_expenses(spd_id);
