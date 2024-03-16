-- name: CreateStoreUserHistory :one
INSERT INTO store_users_history (changed_at, changed_type, changed_by, changed_user_agent, changed_client_ip, store_id, user_id, balance, points, balance_earmark, points_earmark, role_id, state, created_at)
SELECT $3, $4, $5, $6, $7, store_id, user_id, balance, points, balance_earmark, points_earmark, role_id, state, created_at
FROM store_users AS su
WHERE su.store_id = $1 AND su.user_id = $2
RETURNING *;
