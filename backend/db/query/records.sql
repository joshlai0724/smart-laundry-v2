-- name: CreateRecord :one
INSERT INTO records (created_by, created_user_agent, created_client_ip, type, store_id, record_id, user_id, device_id, from_online_payment, amount, point_amount, ts)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (store_id, record_id) DO NOTHING
RETURNING *;

-- name: GetStoreDeviceRecords :many
SELECT r.type, r.user_id, u.name AS user_name, r.amount, r.point_amount, r.ts
FROM records AS r LEFT JOIN users AS u ON r.user_id = u.id
WHERE r.store_id = $1 AND r.device_id = sqlc.arg(device_id)::TEXT
ORDER BY r.ts DESC;

-- name: GetStoreUserRecords :many
SELECT
  r.type,
  r.created_by AS created_by_user_id,
  u1.name AS created_by_user_name,
  r.user_id AS user_id,
  u2.name AS user_name,
  r.device_id,
  sd.name AS device_name,
  sd.real_type AS device_real_type,
  sd.display_type AS device_display_type,
  r.from_online_payment,
  r.amount,
  r.point_amount,
  r.ts
FROM records AS r
LEFT JOIN store_devices AS sd ON r.device_id = sd.device_id AND r.store_id = sd.store_id
LEFT JOIN users AS u1 ON r.created_by = u1.id
LEFT JOIN users AS u2 ON r.user_id = u2.id
WHERE (
  r.user_id = sqlc.arg(user_id)::UUID OR
  r.created_by = sqlc.arg(user_id)::UUID
) AND r.type = ANY(sqlc.arg(types)::TEXT[]) AND r.store_id = $1
ORDER BY r.ts DESC;