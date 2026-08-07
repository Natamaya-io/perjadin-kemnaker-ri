-- Add new Account Code
INSERT INTO account_codes (id, code, mak, description) VALUES
(gen_random_uuid(), 'S.521111', '2158.01.WA.2158.EBA.994.002.S.521111', 'Belanja Keperluan Perkantoran (S)')
ON CONFLICT (code) DO NOTHING;

-- Add new Procurement Types
INSERT INTO procurement_types (id, account_code_id, name, is_active) VALUES
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.521111'), 'Pembelian ATK', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.524111'), 'Tiket Pesawat', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.521119'), 'VIP Halim Perdanakusuma', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.521119'), 'VIP Soekarno Hatta', true),
(gen_random_uuid(), (SELECT id FROM account_codes WHERE code = 'S.521111'), 'Sewa Kendaraan', true)
ON CONFLICT (name) DO NOTHING;
