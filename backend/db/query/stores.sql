-- name: CreateStore :one
INSERT INTO stores (id, name, address, state)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetStore :one
SELECT * FROM stores
WHERE id = $1;

-- name: GetStores :many
SELECT * FROM stores;

-- name: SetStoreState :exec
UPDATE stores
SET state = $2
WHERE id = $1;

-- name: SetStoreNameAndAddress :exec
UPDATE stores
SET name = $2, address = $3
WHERE id = $1;

-- name: SetStorePassword :exec
UPDATE stores
SET password = $2
WHERE id = $1;

-- name: GetUserStores :many
SELECT s.*
FROM stores s INNER JOIN store_users su
ON s.id = su.store_id
WHERE su.user_id = $1;