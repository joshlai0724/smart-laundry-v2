package iotsdk

import (
	"context"

	"github.com/google/uuid"
)

type addPointsToCoinAcceptorRequest struct {
	StoreID  uuid.UUID `json:"store_id"`
	DeviceID string    `json:"device_id"`
	Amount   int32     `json:"amount"`
}

func (i *iot) AddPointsToCoinAcceptor(ctx context.Context, storeID uuid.UUID, deviceID string, amount int32) error {
	_, err := rpc[
		addPointsToCoinAcceptorRequest,
		struct{},
	](
		ctx,
		i.rpcRepo,
		"add-points-to-coin-acceptor",
		addPointsToCoinAcceptorRequest{
			StoreID:  storeID,
			DeviceID: deviceID,
			Amount:   amount,
		},
	)

	return err
}

type getCoinAcceptorStatusRequest struct {
	StoreID  uuid.UUID `json:"store_id"`
	DeviceID string    `json:"device_id"`
}

type getCoinAcceptorStatusResponse struct {
	Points int32  `json:"points"`
	State  string `json:"state"`
	Ts     int64  `json:"ts"`
}

func (m *getCoinAcceptorStatusResponse) convert() *CoinAcceptorStatus {
	return &CoinAcceptorStatus{
		Points: m.Points,
		State:  m.State,
		Ts:     m.Ts,
	}
}

func (i *iot) GetCoinAcceptorStatus(ctx context.Context, storeID uuid.UUID, deviceID string) (*CoinAcceptorStatus, error) {
	m2, err := rpc[
		getCoinAcceptorStatusRequest,
		getCoinAcceptorStatusResponse,
	](
		ctx,
		i.rpcRepo,
		"get-coin-acceptor-status",
		getCoinAcceptorStatusRequest{
			StoreID:  storeID,
			DeviceID: deviceID,
		},
	)

	if err != nil {
		return nil, err
	}

	return m2.Response.convert(), nil
}

type blinkCoinAcceptorRequest struct {
	StoreID  uuid.UUID `json:"store_id"`
	DeviceID string    `json:"device_id"`
}

func (i *iot) BlinkCoinAcceptor(ctx context.Context, storeID uuid.UUID, deviceID string) error {
	_, err := rpc[
		blinkCoinAcceptorRequest,
		struct{},
	](
		ctx,
		i.rpcRepo,
		"blink-coin-acceptor",
		blinkCoinAcceptorRequest{
			StoreID:  storeID,
			DeviceID: deviceID,
		},
	)

	return err
}
