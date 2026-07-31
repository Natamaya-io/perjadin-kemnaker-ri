CREATE TABLE dalkot_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    spd_number VARCHAR(100) UNIQUE,
    execution_date DATE NOT NULL,
    category VARCHAR(50) NOT NULL, -- e.g., 'Jam Kerja', 'Overtime', 'Hari Libur'
    official VARCHAR(255) NOT NULL, -- Pejabat yang didampingi
    dalkot_type VARCHAR(100) NOT NULL, -- 'SPJ RILL' or 'Kebijakan Protokol'
    activity_name VARCHAR(255) NOT NULL,
    location VARCHAR(255) NOT NULL,
    
    surat_tugas_number VARCHAR(100),
    surat_tugas_date DATE,
    report_content TEXT,
    
    total_spj_cost DOUBLE PRECISION DEFAULT 0,
    total_actual_cost DOUBLE PRECISION DEFAULT 0,
    
    status VARCHAR(50) DEFAULT 'Draft',
    
    documentation_file JSONB -- Menyimpan metadata file upload (menggantikan dalkot_documents)
);

CREATE INDEX idx_dalkot_records_deleted_at ON dalkot_records(deleted_at);
CREATE INDEX idx_dalkot_records_spd_number ON dalkot_records(spd_number);
CREATE INDEX idx_dalkot_records_execution_date ON dalkot_records(execution_date);

CREATE TABLE dalkot_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    dalkot_record_id UUID NOT NULL REFERENCES dalkot_records(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    
    assignment_type VARCHAR(50) NOT NULL, -- 'SPJ' or 'RIIL'
    spj_cost DOUBLE PRECISION DEFAULT 0,
    actual_cost DOUBLE PRECISION DEFAULT 0,
    status VARCHAR(50) DEFAULT 'Pending'
);

CREATE INDEX idx_dalkot_assignments_deleted_at ON dalkot_assignments(deleted_at);
CREATE INDEX idx_dalkot_assignments_record_id ON dalkot_assignments(dalkot_record_id);
CREATE INDEX idx_dalkot_assignments_user_id ON dalkot_assignments(user_id);

-- Sequence untuk menghasilkan Nomor SPD Dalkot secara atomik dan anti-bentrok
CREATE SEQUENCE IF NOT EXISTS dalkot_number_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;
