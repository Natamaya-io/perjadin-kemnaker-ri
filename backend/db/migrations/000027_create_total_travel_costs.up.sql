CREATE TABLE total_travel_costs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    year SMALLINT NOT NULL,
    month SMALLINT NOT NULL CHECK (month BETWEEN 1 AND 12),
    dalkot_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    luar_kota_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    luar_negeri_total NUMERIC(18,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (year, month)
);
