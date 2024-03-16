// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: store_users_history.sql

package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createStoreUserHistory = `-- name: CreateStoreUserHistory :one
INSERT INTO store_users_history (changed_at, changed_type, changed_by, changed_user_agent, changed_client_ip, store_id, user_id, balance, points, balance_earmark, points_earmark, role_id, state, created_at)
SELECT $3, $4, $5, $6, $7, store_id, user_id, balance, points, balance_earmark, points_earmark, role_id, state, created_at
FROM store_users AS su
WHERE su.store_id = $1 AND su.user_id = $2
RETURNING changed_at, changed_type, changed_by, changed_user_agent, changed_client_ip, store_id, user_id, balance, points, balance_earmark, points_earmark, role_id, state, created_at, history_created_at
`

type CreateStoreUserHistoryParams struct {
	StoreID          uuid.UUID
	UserID           uuid.UUID
	ChangedAt        int64
	ChangedType      string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
}

func (q *Queries) CreateStoreUserHistory(ctx context.Context, arg CreateStoreUserHistoryParams) (StoreUsersHistory, error) {
	row := q.db.QueryRowContext(ctx, createStoreUserHistory,
		arg.StoreID,
		arg.UserID,
		arg.ChangedAt,
		arg.ChangedType,
		arg.ChangedBy,
		arg.ChangedUserAgent,
		arg.ChangedClientIp,
	)
	var i StoreUsersHistory
	err := row.Scan(
		&i.ChangedAt,
		&i.ChangedType,
		&i.ChangedBy,
		&i.ChangedUserAgent,
		&i.ChangedClientIp,
		&i.StoreID,
		&i.UserID,
		&i.Balance,
		&i.Points,
		&i.BalanceEarmark,
		&i.PointsEarmark,
		&i.RoleID,
		&i.State,
		&i.CreatedAt,
		&i.HistoryCreatedAt,
	)
	return i, err
}
