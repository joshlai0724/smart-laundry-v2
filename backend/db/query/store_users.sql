-- name: CreateStoreUser :one
INSERT INTO store_users (store_id, user_id, role_id, state)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetStoreUser :one
SELECT * FROM store_users
WHERE store_id = $1 AND user_id = $2;

-- name: SetStoreUserState :exec
UPDATE store_users
SET state = $3
WHERE store_id = $1 AND user_id = $2;

-- name: SetStoreUserRoleID :exec
UPDATE store_users
SET role_id = $3
WHERE store_id = $1 AND user_id = $2;

-- name: GetStoreUsersByStoreID :many
SELECT u.id, u.phone_number, u.name, su.state, su.role_id
FROM store_users su, stores s, users u
WHERE su.store_id = s.id AND su.user_id = u.id AND su.store_id = $1;

-- name: SetStoreUserBalance :exec
UPDATE store_users
SET balance = $3, points = $4, balance_earmark = $5, points_earmark = $6
WHERE store_id = $1 AND user_id = $2;