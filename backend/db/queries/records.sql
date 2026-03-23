-- name: CreateTravelRecord :one
INSERT INTO travel_records (
  id, spd_number, employee_id, creator_id, start_date, end_date, location, province, type, purpose, stakeholder, agenda, status, is_viewed, report_status, payment_status, total_cost, surat_tugas_path, surat_tugas_number, surat_tugas_date
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
) RETURNING *;

-- name: GetTravelRecords :many
SELECT * FROM travel_records
WHERE deleted_at IS NULL
  AND (NULLIF($1::text, '') IS NULL OR status = $1)
ORDER BY created_at DESC
LIMIT 1000;

-- name: GetTravelRecordByID :one
SELECT * FROM travel_records WHERE id = $1 AND deleted_at IS NULL LIMIT 1;

-- name: GetLatestSpdNumber :one
SELECT spd_number FROM travel_records
WHERE spd_number LIKE 'ID-SPD-%'
ORDER BY CAST(SUBSTRING(spd_number FROM 8) AS INTEGER) DESC
LIMIT 1;

-- name: GetOverlappingRecords :many
SELECT * FROM travel_records 
WHERE employee_id = $1 
  AND start_date <= $2 
  AND end_date >= $3 
  AND status != 'Rejected'
  AND deleted_at IS NULL;

-- name: UpdateTravelRecord :one
UPDATE travel_records SET
  spd_number = $2,
  employee_id = $3,
  creator_id = $4,
  start_date = $5,
  end_date = $6,
  location = $7,
  province = $8,
  type = $9,
  purpose = $10,
  stakeholder = $11,
  agenda = $12,
  status = $13,
  is_viewed = $14,
  report_status = $15,
  payment_status = $16,
  total_cost = $17,
  surat_tugas_path = $18,
  surat_tugas_number = $19,
  surat_tugas_date = $20,
  updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTravelRecord :exec
UPDATE travel_records SET deleted_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: DeleteTravelRecordBySpd :exec
UPDATE travel_records SET deleted_at = CURRENT_TIMESTAMP WHERE spd_number = $1;

-- name: DeleteTravelLocationsBySpd :exec
UPDATE travel_locations SET deleted_at = CURRENT_TIMESTAMP 
WHERE travel_record_id IN (SELECT id FROM travel_records WHERE spd_number = $1);

-- name: CreateTravelCost :one
INSERT INTO travel_costs (
  travel_record_id, ticket_go, ticket_back, daily_allowance_days, daily_allowance_rate, hotel_days, hotel_rate, local_transport, regional_transport, transport_mode, transport_amount, other_cost, other_cost_desc, receipt_files, ticket_go_file, ticket_back_file, boarding_pass_file, hotel_file, transport_file, additional_costs, details
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21
) RETURNING *;

-- name: GetTravelCostByRecordID :one
SELECT * FROM travel_costs WHERE travel_record_id = $1 LIMIT 1;

-- name: UpdateTravelCost :one
UPDATE travel_costs SET
  ticket_go = $2, ticket_back = $3, daily_allowance_days = $4, daily_allowance_rate = $5, hotel_days = $6, hotel_rate = $7, local_transport = $8, regional_transport = $9, transport_mode = $10, transport_amount = $11, other_cost = $12, other_cost_desc = $13, receipt_files = $14, ticket_go_file = $15, ticket_back_file = $16, boarding_pass_file = $17, hotel_file = $18, transport_file = $19, additional_costs = $20, details = $21
WHERE travel_record_id = $1
RETURNING *;

-- name: DeleteTravelCost :exec
DELETE FROM travel_costs WHERE travel_record_id = $1;

-- name: CreateTravelReport :one
INSERT INTO travel_reports (
  travel_record_id, text, submitted_at, files, sppd_file, surat_tugas_file
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTravelReportByRecordID :one
SELECT * FROM travel_reports WHERE travel_record_id = $1 LIMIT 1;

-- name: UpdateTravelReport :one
UPDATE travel_reports SET
  text = $2, submitted_at = $3, files = $4, sppd_file = $5, surat_tugas_file = $6
WHERE travel_record_id = $1
RETURNING *;

-- name: DeleteTravelReport :exec
DELETE FROM travel_reports WHERE travel_record_id = $1;

-- name: CreateTravelLocation :one
INSERT INTO travel_locations (
  id, travel_record_id, location, province, start_date, end_date
) VALUES (
  $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetTravelLocationsByRecordID :many
SELECT * FROM travel_locations WHERE travel_record_id = $1 AND deleted_at IS NULL ORDER BY start_date ASC, created_at ASC;

-- name: DeleteTravelLocationsByRecordID :exec
UPDATE travel_locations SET deleted_at = CURRENT_TIMESTAMP WHERE travel_record_id = $1;
