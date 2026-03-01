-- Migration Script untuk Sistem Perjalanan Dinas (Perjadin) Protokol Kemnaker RI
-- Script ini aman dijalankan pada database yang sudah ada (existing).
-- Dialect: PostgreSQL

-- 1. Enable Extension untuk UUID
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 2. Create Core Tables (Organization)
CREATE TABLE IF NOT EXISTS departments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE,
    parent_id UUID REFERENCES departments(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ranks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    golongan VARCHAR(10) NOT NULL,
    level INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS positions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    echelon VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 3. Update Existing 'users' Table (Add missing columns)
-- Kita menggunakan DO block untuk mengecek kolom sebelum menambahkannya agar tidak error jika dijalankan ulang.
DO $$
BEGIN
    -- Add NIP if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='nip') THEN
        ALTER TABLE users ADD COLUMN nip VARCHAR(50) UNIQUE;
    END IF;

    -- Add NIK if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='nik') THEN
        ALTER TABLE users ADD COLUMN nik VARCHAR(50);
    END IF;

    -- Add Phone Number if not exists
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='phone_number') THEN
        ALTER TABLE users ADD COLUMN phone_number VARCHAR(20);
    END IF;

    -- Add Relations (Foreign Keys)
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='department_id') THEN
        ALTER TABLE users ADD COLUMN department_id UUID REFERENCES departments(id);
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='rank_id') THEN
        ALTER TABLE users ADD COLUMN rank_id UUID REFERENCES ranks(id);
    END IF;
    
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='position_id') THEN
        ALTER TABLE users ADD COLUMN position_id UUID REFERENCES positions(id);
    END IF;

    -- Add Status
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='users' AND column_name='status') THEN
        ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active';
    END IF;
END $$;

-- 4. Master Data Tables
CREATE TABLE IF NOT EXISTS ref_provinces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(10) UNIQUE
);

CREATE TABLE IF NOT EXISTS ref_cities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID REFERENCES ref_provinces(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50)
);

CREATE TABLE IF NOT EXISTS sbm_uh_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID REFERENCES ref_provinces(id),
    city_id UUID REFERENCES ref_cities(id),
    fullboard_rate DECIMAL(15, 2) NOT NULL,
    fullname_rate DECIMAL(15, 2) NOT NULL,
    outside_city_rate DECIMAL(15, 2) NOT NULL,
    inside_city_rate DECIMAL(15, 2) NOT NULL,
    diklat_rate DECIMAL(15, 2) NOT NULL,
    year INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS sbm_hotel_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    province_id UUID REFERENCES ref_provinces(id),
    echelon_1_rate DECIMAL(15, 2),
    echelon_2_rate DECIMAL(15, 2),
    echelon_3_rate DECIMAL(15, 2),
    echelon_4_rate DECIMAL(15, 2),
    staff_rate DECIMAL(15, 2),
    year INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS sbm_transport_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    origin_city_id UUID REFERENCES ref_cities(id),
    destination_city_id UUID REFERENCES ref_cities(id),
    transport_mode VARCHAR(50),
    business_class_max DECIMAL(15, 2),
    economy_class_max DECIMAL(15, 2),
    year INTEGER NOT NULL
);

-- 5. Finance & Budgeting
CREATE TABLE IF NOT EXISTS budget_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    fiscal_year INTEGER NOT NULL,
    total_budget DECIMAL(15, 2) NOT NULL,
    current_balance DECIMAL(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 6. Travel Management
CREATE TABLE IF NOT EXISTS travel_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_number VARCHAR(100) UNIQUE NOT NULL,
    creator_id UUID REFERENCES users(id),
    approver_id UUID REFERENCES users(id),
    budget_account_id UUID REFERENCES budget_accounts(id),
    purpose TEXT NOT NULL,
    transport_type VARCHAR(50),
    departure_date DATE NOT NULL,
    return_date DATE NOT NULL,
    origin_city_id UUID REFERENCES ref_cities(id),
    destination_city_id UUID REFERENCES ref_cities(id),
    status VARCHAR(50) DEFAULT 'DRAFT',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS travel_spds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    travel_order_id UUID REFERENCES travel_orders(id) ON DELETE CASCADE,
    employee_id UUID REFERENCES users(id),
    spd_number VARCHAR(100) UNIQUE,
    departure_date DATE,
    return_date DATE,
    travel_type VARCHAR(50),
    status VARCHAR(50) DEFAULT 'ISSUED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS travel_expenses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    spd_id UUID REFERENCES travel_spds(id) ON DELETE CASCADE,
    expense_category VARCHAR(50) NOT NULL,
    description VARCHAR(255),
    amount_proposed DECIMAL(15, 2) NOT NULL,
    amount_approved DECIMAL(15, 2),
    proof_file_url TEXT,
    is_lumpsum BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 7. Reports
-- Handle existing travel_reports table which might lack 'id'
DO $$
DECLARE
    r RECORD;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'travel_reports') THEN
        CREATE TABLE travel_reports (
            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            spd_id UUID REFERENCES travel_spds(id) UNIQUE,
            activity_summary TEXT NOT NULL,
            issues_identified TEXT,
            recommendations TEXT,
            submitted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            status VARCHAR(50) DEFAULT 'SUBMITTED'
        );
    ELSE
        -- Table exists, check for 'id' column
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='travel_reports' AND column_name='id') THEN
             ALTER TABLE travel_reports ADD COLUMN id UUID DEFAULT gen_random_uuid();
             
             -- Find and drop existing PK
             FOR r IN SELECT constraint_name FROM information_schema.table_constraints 
                      WHERE table_name = 'travel_reports' AND constraint_type = 'PRIMARY KEY'
             LOOP
                 EXECUTE 'ALTER TABLE travel_reports DROP CONSTRAINT "' || r.constraint_name || '"';
             END LOOP;

             -- Add new PK
             ALTER TABLE travel_reports ADD PRIMARY KEY (id);
        END IF;

        -- Check for 'spd_id'
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='travel_reports' AND column_name='spd_id') THEN
             ALTER TABLE travel_reports ADD COLUMN spd_id UUID REFERENCES travel_spds(id);
             ALTER TABLE travel_reports ADD CONSTRAINT travel_reports_spd_id_unique UNIQUE (spd_id);
        END IF;
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS report_attachments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id UUID REFERENCES travel_reports(id) ON DELETE CASCADE,
    file_url TEXT NOT NULL,
    file_type VARCHAR(50),
    description VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS approval_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_id UUID NOT NULL,
    reference_type VARCHAR(50) NOT NULL,
    approver_id UUID REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    comments TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_travel_orders_dates ON travel_orders(departure_date, return_date);
CREATE INDEX IF NOT EXISTS idx_travel_spds_employee ON travel_spds(employee_id);
CREATE INDEX IF NOT EXISTS idx_expenses_spd ON travel_expenses(spd_id);
