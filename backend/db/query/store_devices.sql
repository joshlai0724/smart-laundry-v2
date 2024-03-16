-- name: CreateStoreDevice :one
INSERT INTO store_devices (store_id, device_id, name, real_type, display_type, state)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (store_id, device_id) DO NOTHING
RETURNING *;

-- name: GetStoreDevices :many
SELECT *
FROM store_devices
WHERE store_id = $1;

-- name: GetStoreDevice :one
SELECT *
FROM store_devices
WHERE store_id = $1 AND device_id = $2;

-- name: SetStoreDeviceNameAndDisplayType :exec
UPDATE store_devices
SET name = $3, display_type=$4
WHERE store_id = $1 AND device_id = $2;