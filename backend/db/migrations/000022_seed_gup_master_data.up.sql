-- Insert Account Codes
INSERT INTO account_codes (id, code, mak, description) VALUES
(gen_random_uuid(), 'S.521119', '2158.01.WA.2158.EBA.994.002.S.521119', 'Belanja Barang Operasional Lainnya'),
(gen_random_uuid(), 'S.523121', '2158.01.WA.2158.EBA.994.002.S.523121', 'Belanja Pemeliharaan Peralatan dan Mesin'),
(gen_random_uuid(), 'S.522141', '2158.01.WA.2158.EBA.994.002.S.522141', 'Belanja Sewa'),
(gen_random_uuid(), 'S.524111', '2158.01.WA.2158.EBA.994.002.S.524111', 'Belanja Perjalanan Dinas Biasa'),
(gen_random_uuid(), 'S.524113', '2158.01.WA.2158.EBA.994.002.S.524113', 'Belanja Perjalanan Dinas Dalam Kota'),
(gen_random_uuid(), 'S.524211', '2158.01.WA.2158.EBA.994.002.S.524211', 'Belanja Perjalanan Dinas Luar Negeri Biasa')
ON CONFLICT (code) DO NOTHING;

-- Insert Procurement Types
INSERT INTO procurement_types (id, account_code_id, name, is_active) VALUES
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.521119'), 'VIP Halim', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.521119'), 'Pass Bandara', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.521119'), 'Langganan AI', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.523121'), 'Pemeliharaan PC', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.523121'), 'Pemeliharaan Printer', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.523121'), 'Pemeliharaan Notebook', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.523121'), 'Pemeliharaan Mobil Operasional Protokol', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.522141'), 'Sewa Mobil Menaker', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.522141'), 'Sewa Kendaraan Protokol', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.524111'), 'Perjalanan Dinas Dalam Negeri', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.524113'), 'Perjalanan Dinas Dalam Kota', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.524211'), 'Luar Negeri', true)
ON CONFLICT (name) DO NOTHING;

-- Insert initial funding sources (Months)
INSERT INTO funding_sources (year, month_number, month_name, gup_label) VALUES 
(2026, 1, 'Januari', 'GUP 01'),
(2026, 2, 'Februari', 'GUP 02'),
(2026, 3, 'Maret', 'GUP 03'),
(2026, 4, 'April', 'GUP 04'),
(2026, 5, 'Mei', 'GUP 05'),
(2026, 6, 'Juni', 'GUP 06'),
(2026, 7, 'Juli', 'GUP 07'),
(2026, 8, 'Agustus', 'GUP 08'),
(2026, 9, 'September', 'GUP 09'),
(2026, 10, 'Oktober', 'GUP 10'),
(2026, 11, 'November', 'GUP 11'),
(2026, 12, 'Desember', 'GUP 12')
ON CONFLICT (year, month_number) DO NOTHING;
