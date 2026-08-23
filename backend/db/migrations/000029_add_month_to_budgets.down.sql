ALTER TABLE budgets DROP CONSTRAINT budgets_procurement_type_id_year_month_key;
ALTER TABLE budgets ADD CONSTRAINT budgets_procurement_type_id_year_key UNIQUE (procurement_type_id, year);
ALTER TABLE budgets DROP COLUMN month_number;
