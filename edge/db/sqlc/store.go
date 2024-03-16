package db

import (
	"database/sql"
)

type IStore interface {
	Querier
}

type SQLStore struct {
	*Queries
	db *sql.DB
}

var _ IStore = (*SQLStore)(nil)

func NewStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

const (
	RecordTypeCoinAcceptorCoinInserted string = "coin_acceptor_coin_inserted"
)
