-- Update existing statuses to match Dalkot
UPDATE travel_records SET status = 'Completed' WHERE status = 'Approved' AND payment_status = 'Paid';
UPDATE travel_records SET status = 'Pending' WHERE status = 'Rejected';

ALTER TABLE travel_records DROP COLUMN IF EXISTS report_status;
ALTER TABLE travel_records DROP COLUMN IF EXISTS payment_status;
