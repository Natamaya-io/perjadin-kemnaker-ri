-- This migration cannot be completely rolled back because we deleted data, but we can re-insert the Jabodetabek locations.
INSERT INTO dalkot_locations (name) VALUES
    ('Kepulauan Seribu'),
    ('Kota Bogor'),
    ('Kabupaten Bogor'),
    ('Kota Depok'),
    ('Kota Tangerang'),
    ('Kota Tangerang Selatan'),
    ('Kabupaten Tangerang'),
    ('Kota Bekasi'),
    ('Kabupaten Bekasi')
ON CONFLICT (name) DO NOTHING;
