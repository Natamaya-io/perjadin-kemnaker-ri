-- Create a sequence for atomic, race-free SPD number generation.
-- Seed it from the current maximum to prevent collisions with existing records.
-- The COALESCE default of 0 handles an empty table gracefully (next value = 1).
CREATE SEQUENCE IF NOT EXISTS spd_number_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

-- Advance the sequence past the highest number already in use so the first
-- nextval() call returns current_max + 1. If the table is empty, GREATEST
-- ensures we seed with 1 (the minimum valid value for the sequence).
SELECT setval(
    'spd_number_seq',
    GREATEST(
        COALESCE(
            MAX(CAST(SUBSTRING(spd_number FROM 8) AS BIGINT)),
            1
        ),
        1
    )
)
FROM travel_records
WHERE spd_number ~ '^ID-SPJ-[0-9]+$';
