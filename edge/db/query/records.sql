-- name: CreateRecord :one
INSERT INTO records (id, device_id, type, amount, is_uploaded, uploaded_at, ts)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: SetRecordIsUploaded :exec
UPDATE records
SET is_uploaded = $2, uploaded_at=$3
WHERE id = $1;

-- name: GetUnuploadedRecords :many
SELECT *
FROM records
WHERE is_uploaded = false
LIMIT $1;