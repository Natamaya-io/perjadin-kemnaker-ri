CREATE SEQUENCE IF NOT EXISTS dalkot_spd_number_seq START 1;

DO $$
DECLARE
    max_id BIGINT;
BEGIN
    SELECT MAX(CAST(NULLIF(regexp_replace(spd_number, '\D', '', 'g'), '') AS BIGINT)) 
    INTO max_id 
    FROM dalkot_records 
    WHERE spd_number ILIKE 'DLK-%' AND deleted_at IS NULL;

    IF max_id IS NOT NULL THEN
        EXECUTE 'ALTER SEQUENCE dalkot_spd_number_seq RESTART WITH ' || (max_id + 1);
    END IF;
END $$;
