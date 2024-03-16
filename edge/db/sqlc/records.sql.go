// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0
// source: records.sql

package db

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const createRecord = `-- name: CreateRecord :one
INSERT INTO records (id, device_id, type, amount, is_uploaded, uploaded_at, ts)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, device_id, type, amount, is_uploaded, uploaded_at, ts, created_at
`

type CreateRecordParams struct {
	ID         uuid.UUID
	DeviceID   string
	Type       string
	Amount     int32
	IsUploaded bool
	UploadedAt sql.NullInt64
	Ts         int64
}

func (q *Queries) CreateRecord(ctx context.Context, arg CreateRecordParams) (Record, error) {
	row := q.db.QueryRowContext(ctx, createRecord,
		arg.ID,
		arg.DeviceID,
		arg.Type,
		arg.Amount,
		arg.IsUploaded,
		arg.UploadedAt,
		arg.Ts,
	)
	var i Record
	err := row.Scan(
		&i.ID,
		&i.DeviceID,
		&i.Type,
		&i.Amount,
		&i.IsUploaded,
		&i.UploadedAt,
		&i.Ts,
		&i.CreatedAt,
	)
	return i, err
}

const getUnuploadedRecords = `-- name: GetUnuploadedRecords :many
SELECT id, device_id, type, amount, is_uploaded, uploaded_at, ts, created_at
FROM records
WHERE is_uploaded = false
LIMIT $1
`

func (q *Queries) GetUnuploadedRecords(ctx context.Context, limit int32) ([]Record, error) {
	rows, err := q.db.QueryContext(ctx, getUnuploadedRecords, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Record{}
	for rows.Next() {
		var i Record
		if err := rows.Scan(
			&i.ID,
			&i.DeviceID,
			&i.Type,
			&i.Amount,
			&i.IsUploaded,
			&i.UploadedAt,
			&i.Ts,
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

const setRecordIsUploaded = `-- name: SetRecordIsUploaded :exec
UPDATE records
SET is_uploaded = $2, uploaded_at=$3
WHERE id = $1
`

type SetRecordIsUploadedParams struct {
	ID         uuid.UUID
	IsUploaded bool
	UploadedAt sql.NullInt64
}

func (q *Queries) SetRecordIsUploaded(ctx context.Context, arg SetRecordIsUploadedParams) error {
	_, err := q.db.ExecContext(ctx, setRecordIsUploaded, arg.ID, arg.IsUploaded, arg.UploadedAt)
	return err
}
