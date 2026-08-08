CREATE TABLE dalkot_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_name VARCHAR(255) NOT NULL UNIQUE,
    rate_amount NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial rates
INSERT INTO dalkot_rates (category_name, rate_amount) VALUES
    ('Jam Kerja', 150000),
    ('Lembur', 150000),
    ('Hari Libur', 250000)
ON CONFLICT (category_name) DO UPDATE 
SET rate_amount = EXCLUDED.rate_amount,
    updated_at = CURRENT_TIMESTAMP;
