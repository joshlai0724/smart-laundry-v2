-- name: CreateToken :one
INSERT INTO tokens (id, type, user_agent, client_ip, user_id, expired_at, issued_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetToken :one
SELECT * FROM tokens
WHERE id = $1;