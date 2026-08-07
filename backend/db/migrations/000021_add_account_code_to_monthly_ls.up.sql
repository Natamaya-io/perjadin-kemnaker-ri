ALTER TABLE monthly_ls ADD COLUMN account_code_id UUID REFERENCES account_codes(id) ON DELETE SET NULL;
