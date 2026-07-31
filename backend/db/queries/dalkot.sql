-- name: CreateDalkotRecord :one
INSERT INTO dalkot_records (
    id, spd_number, execution_date, category, official, dalkot_type, activity_name, location,
    surat_tugas_number, surat_tugas_date, report_content, status, documentation_file
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: UpdateDalkotRecord :one
UPDATE dalkot_records
SET
    spd_number = $2,
    execution_date = $3,
    category = $4,
    official = $5,
    dalkot_type = $6,
    activity_name = $7,
    location = $8,
    surat_tugas_number = $9,
    surat_tugas_date = $10,
    report_content = $11,
    status = $12,
    documentation_file = $13,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteDalkotRecord :exec
UPDATE dalkot_records
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetDalkotRecordByID :one
SELECT * FROM dalkot_records
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetDalkotRecords :many
SELECT * FROM dalkot_records
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: CreateDalkotAssignment :one
INSERT INTO dalkot_assignments (
    id, dalkot_record_id, user_id, assignment_type, spj_cost, actual_cost, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: UpdateDalkotAssignment :one
UPDATE dalkot_assignments
SET
    assignment_type = $2,
    spj_cost = $3,
    actual_cost = $4,
    status = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteDalkotAssignment :exec
UPDATE dalkot_assignments
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetDalkotAssignmentsByRecordID :many
SELECT * FROM dalkot_assignments
WHERE dalkot_record_id = $1 AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: GetDalkotLocations :many
SELECT * FROM dalkot_locations
WHERE is_active = true
ORDER BY id ASC;
