CREATE TABLE account_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(30) NOT NULL UNIQUE,
    mak VARCHAR(100) NOT NULL UNIQUE,
    description VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE procurement_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_code_id UUID NOT NULL REFERENCES account_codes(id) ON DELETE RESTRICT,
    name VARCHAR(180) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE funding_sources (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year SMALLINT NOT NULL,
    month_number SMALLINT NOT NULL CHECK (month_number BETWEEN 1 AND 12),
    month_name VARCHAR(20) NOT NULL,
    gup_label VARCHAR(20) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (year, month_number),
    UNIQUE (year, gup_label)
);

CREATE TABLE budgets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year SMALLINT NOT NULL,
    procurement_type_id UUID NOT NULL REFERENCES procurement_types(id) ON DELETE RESTRICT,
    amount NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (year, procurement_type_id)
);

CREATE TABLE monthly_ls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    funding_source_id UUID NOT NULL UNIQUE REFERENCES funding_sources(id) ON DELETE CASCADE,
    amount NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE gup_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    business_id VARCHAR(50) NOT NULL UNIQUE,
    payment_description VARCHAR(255) NOT NULL,
    procurement_type_id UUID NOT NULL REFERENCES procurement_types(id) ON DELETE RESTRICT,
    funding_source_id UUID REFERENCES funding_sources(id) ON DELETE SET NULL,
    value_amount NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (value_amount >= 0),
    paid_amount NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (paid_amount >= 0),
    tax_amount NUMERIC(18,2) NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    receipt_date DATE,
    recipient VARCHAR(200),
    pum VARCHAR(150),
    document_file JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
