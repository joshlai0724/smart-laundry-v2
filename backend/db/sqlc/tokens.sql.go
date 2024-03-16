// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: tokens.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createToken = `-- name: CreateToken :one
INSERT INTO tokens (id, type, user_agent, client_ip, user_id, expired_at, issued_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, type, user_agent, client_ip, is_blocked, user_id, expired_at, issued_at, create_at
`

type CreateTokenParams struct {
	ID        uuid.UUID
	Type      string
	UserAgent string
	ClientIp  string
	UserID    uuid.UUID
	ExpiredAt int64
	IssuedAt  int64
}

func (q *Queries) CreateToken(ctx context.Context, arg CreateTokenParams) (Token, error) {
	row := q.db.QueryRowContext(ctx, createToken,
		arg.ID,
		arg.Type,
		arg.UserAgent,
		arg.ClientIp,
		arg.UserID,
		arg.ExpiredAt,
		arg.IssuedAt,
	)
	var i Token
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.UserID,
		&i.ExpiredAt,
		&i.IssuedAt,
		&i.CreateAt,
	)
	return i, err
}

const getToken = `-- name: GetToken :one
SELECT id, type, user_agent, client_ip, is_blocked, user_id, expired_at, issued_at, create_at FROM tokens
WHERE id = $1
`

func (q *Queries) GetToken(ctx context.Context, id uuid.UUID) (Token, error) {
	row := q.db.QueryRowContext(ctx, getToken, id)
	var i Token
	err := row.Scan(
		&i.ID,
		&i.Type,
		&i.UserAgent,
		&i.ClientIp,
		&i.IsBlocked,
		&i.UserID,
		&i.ExpiredAt,
		&i.IssuedAt,
		&i.CreateAt,
	)
	return i, err
}
