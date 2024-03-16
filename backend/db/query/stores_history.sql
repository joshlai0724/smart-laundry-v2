-- name: CreateStoreHistory :one
INSERT INTO stores_history (changed_at, changed_type, changed_by, changed_user_agent, changed_client_ip, store_id, name, address, state, password, created_at)
SELECT $2, $3, $4, $5, $6, id, name, address, state, password, created_at
FROM stores AS s
WHERE s.id = $1
RETURNING *;
