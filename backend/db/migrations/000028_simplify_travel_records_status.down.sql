ALTER TABLE travel_records ADD COLUMN report_status VARCHAR(50) DEFAULT 'Pending';
ALTER TABLE travel_records ADD COLUMN payment_status VARCHAR(50) DEFAULT 'Unpaid';
