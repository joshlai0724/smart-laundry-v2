// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: stores.sql

package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createStore = `-- name: CreateStore :one
INSERT INTO stores (id, name, address, state)
VALUES ($1, $2, $3, $4)
RETURNING id, name, address, state, password, created_at
`

type CreateStoreParams struct {
	ID      uuid.UUID
	Name    string
	Address string
	State   string
}

func (q *Queries) CreateStore(ctx context.Context, arg CreateStoreParams) (Store, error) {
	row := q.db.QueryRowContext(ctx, createStore,
		arg.ID,
		arg.Name,
		arg.Address,
		arg.State,
	)
	var i Store
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Address,
		&i.State,
		&i.Password,
		&i.CreatedAt,
	)
	return i, err
}

const getStore = `-- name: GetStore :one
SELECT id, name, address, state, password, created_at FROM stores
WHERE id = $1
`

func (q *Queries) GetStore(ctx context.Context, id uuid.UUID) (Store, error) {
	row := q.db.QueryRowContext(ctx, getStore, id)
	var i Store
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Address,
		&i.State,
		&i.Password,
		&i.CreatedAt,
	)
	return i, err
}

const getStores = `-- name: GetStores :many
SELECT id, name, address, state, password, created_at FROM stores
`

func (q *Queries) GetStores(ctx context.Context) ([]Store, error) {
	rows, err := q.db.QueryContext(ctx, getStores)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Store{}
	for rows.Next() {
		var i Store
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Address,
			&i.State,
			&i.Password,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getUserStores = `-- name: GetUserStores :many
SELECT s.id, s.name, s.address, s.state, s.password, s.created_at
FROM stores s INNER JOIN store_users su
ON s.id = su.store_id
WHERE su.user_id = $1
`

func (q *Queries) GetUserStores(ctx context.Context, userID uuid.UUID) ([]Store, error) {
	rows, err := q.db.QueryContext(ctx, getUserStores, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Store{}
	for rows.Next() {
		var i Store
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Address,
			&i.State,
			&i.Password,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const setStoreNameAndAddress = `-- name: SetStoreNameAndAddress :exec
UPDATE stores
SET name = $2, address = $3
WHERE id = $1
`

type SetStoreNameAndAddressParams struct {
	ID      uuid.UUID
	Name    string
	Address string
}

func (q *Queries) SetStoreNameAndAddress(ctx context.Context, arg SetStoreNameAndAddressParams) error {
	_, err := q.db.ExecContext(ctx, setStoreNameAndAddress, arg.ID, arg.Name, arg.Address)
	return err
}

const setStorePassword = `-- name: SetStorePassword :exec
UPDATE stores
SET password = $2
WHERE id = $1
`

type SetStorePasswordParams struct {
	ID       uuid.UUID
	Password sql.NullString
}

func (q *Queries) SetStorePassword(ctx context.Context, arg SetStorePasswordParams) error {
	_, err := q.db.ExecContext(ctx, setStorePassword, arg.ID, arg.Password)
	return err
}

const setStoreState = `-- name: SetStoreState :exec
UPDATE stores
SET state = $2
WHERE id = $1
`

type SetStoreStateParams struct {
	ID    uuid.UUID
	State string
}

func (q *Queries) SetStoreState(ctx context.Context, arg SetStoreStateParams) error {
	_, err := q.db.ExecContext(ctx, setStoreState, arg.ID, arg.State)
	return err
}
