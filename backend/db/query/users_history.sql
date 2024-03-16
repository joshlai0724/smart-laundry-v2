-- name: CreateUserHistory :one
INSERT INTO users_history (changed_at, changed_type, changed_by, changed_user_agent, changed_client_ip, user_id, phone_number, name, password, password_error_count, password_changed_at, role_id, state, created_at)
SELECT $2, $3, $4, $5, $6, id, phone_number, name, password, password_error_count, password_changed_at, role_id, state, created_at
FROM users AS u
WHERE u.id = $1
RETURNING *;
