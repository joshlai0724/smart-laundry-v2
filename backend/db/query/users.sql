-- name: CreateUser :one
INSERT INTO users (id, phone_number, name, password, role_id, state)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByPhoneNumber :one
SELECT * FROM users
WHERE phone_number = $1;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: SetUserPasswordAndState :exec
UPDATE users
SET password = $2, password_error_count = $3, password_changed_at = $4, state = $5
WHERE id = $1;

-- name: SetUserName :exec
UPDATE users
SET name = $2
WHERE id = $1;