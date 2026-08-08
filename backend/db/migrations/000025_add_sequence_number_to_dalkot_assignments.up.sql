CREATE SEQUENCE IF NOT EXISTS dalkot_assignments_seq_no START WITH 1;
ALTER TABLE dalkot_assignments ADD COLUMN IF NOT EXISTS sequence_number INTEGER DEFAULT nextval('dalkot_assignments_seq_no');

WITH numbered AS (
  SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC) as rn
  FROM dalkot_assignments
)
UPDATE dalkot_assignments da
SET sequence_number = n.rn
FROM numbered n
WHERE da.id = n.id AND (da.sequence_number = 0 OR da.sequence_number IS NULL);
