// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package db

import (
	"context"

	"github.com/google/uuid"
)

type Querier interface {
	BlockVerCodes(ctx context.Context, id uuid.UUID) error
	CreateRecord(ctx context.Context, arg CreateRecordParams) (Record, error)
	CreateStore(ctx context.Context, arg CreateStoreParams) (Store, error)
	CreateStoreDevice(ctx context.Context, arg CreateStoreDeviceParams) (StoreDevice, error)
	CreateStoreDeviceHistory(ctx context.Context, arg CreateStoreDeviceHistoryParams) (StoreDevicesHistory, error)
	CreateStoreHistory(ctx context.Context, arg CreateStoreHistoryParams) (StoresHistory, error)
	CreateStoreUser(ctx context.Context, arg CreateStoreUserParams) (StoreUser, error)
	CreateStoreUserHistory(ctx context.Context, arg CreateStoreUserHistoryParams) (StoreUsersHistory, error)
	CreateToken(ctx context.Context, arg CreateTokenParams) (Token, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	CreateUserHistory(ctx context.Context, arg CreateUserHistoryParams) (UsersHistory, error)
	CreateVerCode(ctx context.Context, arg CreateVerCodeParams) (VerCode, error)
	GetStore(ctx context.Context, id uuid.UUID) (Store, error)
	GetStoreDevice(ctx context.Context, arg GetStoreDeviceParams) (StoreDevice, error)
	GetStoreDeviceRecords(ctx context.Context, arg GetStoreDeviceRecordsParams) ([]GetStoreDeviceRecordsRow, error)
	GetStoreDevices(ctx context.Context, storeID uuid.UUID) ([]StoreDevice, error)
	GetStoreUser(ctx context.Context, arg GetStoreUserParams) (StoreUser, error)
	GetStoreUserRecords(ctx context.Context, arg GetStoreUserRecordsParams) ([]GetStoreUserRecordsRow, error)
	GetStoreUsersByStoreID(ctx context.Context, storeID uuid.UUID) ([]GetStoreUsersByStoreIDRow, error)
	GetStores(ctx context.Context) ([]Store, error)
	GetToken(ctx context.Context, id uuid.UUID) (Token, error)
	GetUser(ctx context.Context, id uuid.UUID) (User, error)
	GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (User, error)
	GetUserStores(ctx context.Context, userID uuid.UUID) ([]Store, error)
	GetVerCodesByTypeAndCode(ctx context.Context, arg GetVerCodesByTypeAndCodeParams) ([]VerCode, error)
	GetVerCodesByTypeAndPhoneNumber(ctx context.Context, arg GetVerCodesByTypeAndPhoneNumberParams) ([]VerCode, error)
	GetVerCodesByTypeAndPhoneNumberAndCode(ctx context.Context, arg GetVerCodesByTypeAndPhoneNumberAndCodeParams) ([]VerCode, error)
	SetStoreDeviceNameAndDisplayType(ctx context.Context, arg SetStoreDeviceNameAndDisplayTypeParams) error
	SetStoreNameAndAddress(ctx context.Context, arg SetStoreNameAndAddressParams) error
	SetStorePassword(ctx context.Context, arg SetStorePasswordParams) error
	SetStoreState(ctx context.Context, arg SetStoreStateParams) error
	SetStoreUserBalance(ctx context.Context, arg SetStoreUserBalanceParams) error
	SetStoreUserRoleID(ctx context.Context, arg SetStoreUserRoleIDParams) error
	SetStoreUserState(ctx context.Context, arg SetStoreUserStateParams) error
	SetUserName(ctx context.Context, arg SetUserNameParams) error
	SetUserPasswordAndState(ctx context.Context, arg SetUserPasswordAndStateParams) error
}

var _ Querier = (*Queries)(nil)