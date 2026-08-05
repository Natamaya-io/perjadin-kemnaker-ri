-- Rollback: hapus kolom timestamps yang ditambahkan (hanya jika tidak ada sebelumnya)
-- CATATAN: Pastikan tidak ada data penting sebelum rollback

ALTER TABLE account_codes
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE procurement_types
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE funding_sources
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE budgets
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;

ALTER TABLE gup_transactions
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;
