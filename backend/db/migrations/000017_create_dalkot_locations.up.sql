CREATE TABLE dalkot_locations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO dalkot_locations (name) VALUES
    ('Jakarta Pusat'),
    ('Jakarta Selatan'),
    ('Jakarta Barat'),
    ('Jakarta Utara'),
    ('Jakarta Timur');
