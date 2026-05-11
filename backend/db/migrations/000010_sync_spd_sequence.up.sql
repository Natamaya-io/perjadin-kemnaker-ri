-- Sync the sequence again to account for records created while the old logic was active.
-- This ensures the sequence catches up to the true maximum ID-SPJ number currently in the table.
SELECT setval(
    'spd_number_seq',
    GREATEST(
        COALESCE(
            (SELECT MAX(CAST(NULLIF(regexp_replace(spd_number, '\D', '', 'g'), '') AS BIGINT))
             FROM travel_records
             WHERE spd_number ILIKE 'ID-SPJ-%'),
            1
        ),
        1
    )
);
