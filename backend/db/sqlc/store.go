package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
)

type IStore interface {
	Querier

	InitDB(ctx context.Context) error

	CreateUserWithLog(ctx context.Context, arg CreateUserWithLogParams) (User, error)
	SetUserPasswordAndStateWithLog(ctx context.Context, arg SetUserPasswordAndStateWithLogParams) error
	SetUserNameWithLog(ctx context.Context, arg SetUserNameWithLogParams) error

	CreateStoreWithLog(ctx context.Context, arg CreateStoreWithLogParams) (Store, error)
	SetStoreStateWithLog(ctx context.Context, arg SetStoreStateWithLogParams) error
	SetStoreNameAndAddressWithLog(ctx context.Context, arg SetStoreNameAndAddressWithLogParams) error
	SetStorePasswordWithLog(ctx context.Context, arg SetStorePasswordWithLogParams) error

	CreateStoreUserWithLog(ctx context.Context, arg CreateStoreUserWithLogParams) (StoreUser, error)
	SetStoreUserStateWithLog(ctx context.Context, arg SetStoreUserStateWithLogParams) error
	SetStoreUserRoleIDWithLog(ctx context.Context, arg SetStoreUserRoleIDWithLogParams) error
	SetStoreUserBalanceWithLog(ctx context.Context, arg SetStoreUserBalanceWithLogParams) error

	CreateStoreDeviceWithLog(ctx context.Context, arg CreateStoreDeviceWithLogParams) (StoreDevice, error)
	SetStoreDeviceNameAndDisplayTypeWithLog(ctx context.Context, arg SetStoreDeviceNameAndDisplayTypeWithLogParams) error
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

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

var initDB = `
INSERT INTO stores (id, name, address, state, password, created_at)
VALUES ('1bb62adb-8741-4dd7-8d13-d205bbb8a9a0', '1234567890', '12345678901234567890123456789012345678901234567890', 'active', '$2a$10$nZEyYXPJXxDhjKqtqlVOTurhiLBLXuFWRHX0D1erDdX15Jswl5X.u', 1699145166404);

INSERT INTO users (id, phone_number, name, password, password_error_count, password_changed_at, role_id, state, created_at)
VALUES ('c3d4daba-01b8-45fb-9926-333ab1cb115e', 'admin', 'admin測試帳號', '$2a$10$qbUm.KeiZRggI/.x/LkdYOVe5hnJyLT7nycuhPar/46lELq1GlmvW', 0, NULL, 1, 'active', 1699145130406);
INSERT INTO users (id, phone_number, name, password, password_error_count, password_changed_at, role_id, state, created_at)
VALUES ('b61f42fd-234f-4ba4-8a1a-df7c349f26e8', 'hq', 'hq測試帳號', '$2a$10$qbUm.KeiZRggI/.x/LkdYOVe5hnJyLT7nycuhPar/46lELq1GlmvW', 0, NULL, 2, 'active', 1699145130406);
INSERT INTO users (id, phone_number, name, password, password_error_count, password_changed_at, role_id, state, created_at)
VALUES ('ec2dfa2c-f4cd-445f-a58c-a95d949d2dc3', 'member', 'member測試帳號', '$2a$10$qbUm.KeiZRggI/.x/LkdYOVe5hnJyLT7nycuhPar/46lELq1GlmvW', 0, NULL, 3, 'active', 1699145130406);
`

func (store *SQLStore) InitDB(ctx context.Context) error {
	_, err := store.db.ExecContext(ctx, initDB)
	return err
}

type CreateUserWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	ID               uuid.UUID
	PhoneNumber      string
	Name             string
	Password         string
	RoleID           int16
	State            string
}

func (store *SQLStore) CreateUserWithLog(ctx context.Context, arg CreateUserWithLogParams) (User, error) {
	result := User{}

	oerr := store.execTx(ctx, func(q *Queries) error {
		var err error

		result, err = q.CreateUser(ctx, CreateUserParams{
			ID:          arg.ID,
			PhoneNumber: arg.PhoneNumber,
			Name:        arg.Name,
			Password:    arg.Password,
			RoleID:      arg.RoleID,
			State:       arg.State,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateUserHistory(ctx, CreateUserHistoryParams{
			ID:               arg.ID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return result, oerr
}

type SetUserPasswordAndStateWithLogParams struct {
	ChangedAt          int64
	ChangeType         string
	ChangedBy          uuid.NullUUID
	ChangedUserAgent   sql.NullString
	ChangedClientIp    sql.NullString
	ID                 uuid.UUID
	Password           string
	PasswordErrorCount int16
	PasswordChangedAt  sql.NullInt64
	State              string
}

func (store *SQLStore) SetUserPasswordAndStateWithLog(ctx context.Context, arg SetUserPasswordAndStateWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetUserPasswordAndState(ctx, SetUserPasswordAndStateParams{
			ID:                 arg.ID,
			Password:           arg.Password,
			PasswordErrorCount: arg.PasswordErrorCount,
			PasswordChangedAt:  arg.PasswordChangedAt,
			State:              arg.State,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateUserHistory(ctx, CreateUserHistoryParams{
			ID:               arg.ID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type SetUserNameWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	ID               uuid.UUID
	Name             string
}

func (store *SQLStore) SetUserNameWithLog(ctx context.Context, arg SetUserNameWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetUserName(ctx, SetUserNameParams{
			ID:   arg.ID,
			Name: arg.Name,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateUserHistory(ctx, CreateUserHistoryParams{
			ID:               arg.ID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type CreateStoreWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	ID               uuid.UUID
	Name             string
	Address          string
	State            string
}

func (store *SQLStore) CreateStoreWithLog(ctx context.Context, arg CreateStoreWithLogParams) (Store, error) {
	result := Store{}

	oerr := store.execTx(ctx, func(q *Queries) error {
		var err error

		result, err = q.CreateStore(ctx, CreateStoreParams{
			ID:      arg.ID,
			Name:    arg.Name,
			Address: arg.Address,
			State:   arg.State,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreHistory(ctx, CreateStoreHistoryParams{
			ID:               arg.ID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return result, oerr
}

type SetStoreStateWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	ID               uuid.UUID
	State            string
}

func (store *SQLStore) SetStoreStateWithLog(ctx context.Context, arg SetStoreStateWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetStoreState(ctx, SetStoreStateParams{
			ID:    arg.ID,
			State: arg.State,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreHistory(ctx, CreateStoreHistoryParams{
			ID:               arg.ID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type SetStoreNameAndAddressWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	ID               uuid.UUID
	Name             string
	Address          string
}

func (store *SQLStore) SetStoreNameAndAddressWithLog(ctx context.Context, arg SetStoreNameAndAddressWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetStoreNameAndAddress(ctx, SetStoreNameAndAddressParams{
			ID:      arg.ID,
			Name:    arg.Name,
			Address: arg.Address,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreHistory(ctx, CreateStoreHistoryParams{
			ID:               arg.ID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type SetStorePasswordWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	ID               uuid.UUID
	Password         sql.NullString
}

func (store *SQLStore) SetStorePasswordWithLog(ctx context.Context, arg SetStorePasswordWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetStorePassword(ctx, SetStorePasswordParams{
			ID:       arg.ID,
			Password: arg.Password,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreHistory(ctx, CreateStoreHistoryParams{
			ID:               arg.ID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type CreateStoreUserWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	StoreID          uuid.UUID
	UserID           uuid.UUID
	RoleID           int16
	State            string
}

func (store *SQLStore) CreateStoreUserWithLog(ctx context.Context, arg CreateStoreUserWithLogParams) (StoreUser, error) {
	result := StoreUser{}

	oerr := store.execTx(ctx, func(q *Queries) error {
		var err error

		result, err = q.CreateStoreUser(ctx, CreateStoreUserParams{
			StoreID: arg.StoreID,
			UserID:  arg.UserID,
			RoleID:  arg.RoleID,
			State:   arg.State,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreUserHistory(ctx, CreateStoreUserHistoryParams{
			StoreID:          arg.StoreID,
			UserID:           arg.UserID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return result, oerr
}

type SetStoreUserStateWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	StoreID          uuid.UUID
	UserID           uuid.UUID
	State            string
}

func (store *SQLStore) SetStoreUserStateWithLog(ctx context.Context, arg SetStoreUserStateWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetStoreUserState(ctx, SetStoreUserStateParams{
			StoreID: arg.StoreID,
			UserID:  arg.UserID,
			State:   arg.State,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreUserHistory(ctx, CreateStoreUserHistoryParams{
			StoreID:          arg.StoreID,
			UserID:           arg.UserID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type SetStoreUserRoleIDWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	StoreID          uuid.UUID
	UserID           uuid.UUID
	RoleID           int16
}

func (store *SQLStore) SetStoreUserRoleIDWithLog(ctx context.Context, arg SetStoreUserRoleIDWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetStoreUserRoleID(ctx, SetStoreUserRoleIDParams{
			StoreID: arg.StoreID,
			UserID:  arg.UserID,
			RoleID:  arg.RoleID,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreUserHistory(ctx, CreateStoreUserHistoryParams{
			StoreID:          arg.StoreID,
			UserID:           arg.UserID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type SetStoreUserBalanceWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	StoreID          uuid.UUID
	UserID           uuid.UUID
	Balance          int32
	Points           int32
	BalanceEarmark   int32
	PointsEarmark    int32
}

func (store *SQLStore) SetStoreUserBalanceWithLog(ctx context.Context, arg SetStoreUserBalanceWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetStoreUserBalance(ctx, SetStoreUserBalanceParams{
			StoreID:        arg.StoreID,
			UserID:         arg.UserID,
			Balance:        arg.Balance,
			Points:         arg.Points,
			BalanceEarmark: arg.BalanceEarmark,
			PointsEarmark:  arg.PointsEarmark,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreUserHistory(ctx, CreateStoreUserHistoryParams{
			StoreID:          arg.StoreID,
			UserID:           arg.UserID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}

type CreateStoreDeviceWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	StoreID          uuid.UUID
	DeviceID         string
	Name             string
	RealType         string
	DisplayType      string
	State            string
}

func (store *SQLStore) CreateStoreDeviceWithLog(ctx context.Context, arg CreateStoreDeviceWithLogParams) (StoreDevice, error) {
	result := StoreDevice{}

	oerr := store.execTx(ctx, func(q *Queries) error {
		var err error

		result, err = q.CreateStoreDevice(ctx, CreateStoreDeviceParams{
			StoreID:     arg.StoreID,
			DeviceID:    arg.DeviceID,
			Name:        arg.Name,
			RealType:    arg.RealType,
			DisplayType: arg.DisplayType,
			State:       arg.State,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreDeviceHistory(ctx, CreateStoreDeviceHistoryParams{
			StoreID:          arg.StoreID,
			DeviceID:         arg.DeviceID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return result, oerr
}

type SetStoreDeviceNameAndDisplayTypeWithLogParams struct {
	ChangedAt        int64
	ChangeType       string
	ChangedBy        uuid.NullUUID
	ChangedUserAgent sql.NullString
	ChangedClientIp  sql.NullString
	StoreID          uuid.UUID
	DeviceID         string
	Name             string
	DisplayType      string
}

func (store *SQLStore) SetStoreDeviceNameAndDisplayTypeWithLog(ctx context.Context, arg SetStoreDeviceNameAndDisplayTypeWithLogParams) error {
	oerr := store.execTx(ctx, func(q *Queries) error {
		err := q.SetStoreDeviceNameAndDisplayType(ctx, SetStoreDeviceNameAndDisplayTypeParams{
			StoreID:     arg.StoreID,
			DeviceID:    arg.DeviceID,
			Name:        arg.Name,
			DisplayType: arg.DisplayType,
		})
		if err != nil {
			return err
		}
		if _, err := q.CreateStoreDeviceHistory(ctx, CreateStoreDeviceHistoryParams{
			StoreID:          arg.StoreID,
			DeviceID:         arg.DeviceID,
			ChangedAt:        arg.ChangedAt,
			ChangedType:      arg.ChangeType,
			ChangedBy:        arg.ChangedBy,
			ChangedUserAgent: arg.ChangedUserAgent,
			ChangedClientIp:  arg.ChangedClientIp,
		}); err != nil {
			return err
		}
		return nil
	})

	return oerr
}
