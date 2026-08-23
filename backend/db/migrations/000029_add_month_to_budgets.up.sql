ALTER TABLE budgets ADD COLUMN IF NOT EXISTS month_number SMALLINT DEFAULT 0 NOT NULL;
ALTER TABLE budgets DROP CONSTRAINT IF EXISTS budgets_procurement_type_id_year_key;
ALTER TABLE budgets ADD CONSTRAINT budgets_procurement_type_id_year_month_key UNIQUE (procurement_type_id, year, month_number);
