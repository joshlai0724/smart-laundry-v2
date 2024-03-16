-- name: CreateVerCode :one
INSERT INTO ver_codes (id, phone_number, code, type, request_id, expired_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetVerCodesByTypeAndPhoneNumber :many
SELECT * FROM ver_codes
WHERE type = $1 AND phone_number = $2 AND create_at >= sqlc.arg(from_ts);

-- name: GetVerCodesByTypeAndPhoneNumberAndCode :many
SELECT * FROM ver_codes
WHERE type = $1 AND phone_number = $2 AND code = $3;

-- name: GetVerCodesByTypeAndCode :many
SELECT * FROM ver_codes
WHERE type = $1 AND code = $2;

-- name: BlockVerCodes :exec
UPDATE ver_codes
SET is_blocked = TRUE
WHERE id = $1;