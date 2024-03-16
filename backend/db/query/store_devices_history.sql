-- name: CreateStoreDeviceHistory :one
INSERT INTO store_devices_history (changed_at, changed_type, changed_by, changed_user_agent, changed_client_ip, store_id, device_id, name, real_type, display_type, state, created_at)
SELECT $3, $4, $5, $6, $7, store_id, device_id, name, real_type, display_type, state, created_at
FROM store_devices AS sd
WHERE sd.store_id = $1 AND sd.device_id = $2
RETURNING *;
