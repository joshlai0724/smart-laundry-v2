package iotsdk

import (
	"backend/heartbeat"
	"context"
	"fmt"

	"github.com/google/uuid"
	goredislib "github.com/redis/go-redis/v9"
)

type iotDecorator struct {
	redisClient *goredislib.Client
	*iot
}

func New(url, appName string, rc *goredislib.Client) (*iotDecorator, error) {
	d := &iotDecorator{redisClient: rc}

	iot, err := newIot(url, appName)
	if err != nil {
		return nil, err
	}

	d.iot = iot

	return d, nil
}

func (d *iotDecorator) AddPointsToCoinAcceptor(ctx context.Context, storeID uuid.UUID, deviceID string, amount int32) error {
	exist, err := heartbeat.CheckHeartbeat(d.redisClient, heartbeat.GetStoreIDHeartbeatName(storeID.String()))
	if err != nil {
		return err
	}
	if !exist {
		return &StoreNotFoundError{S: fmt.Sprintf("store not found, store_id=%s", storeID)}
	}
	return d.iot.AddPointsToCoinAcceptor(ctx, storeID, deviceID, amount)
}

func (d *iotDecorator) GetCoinAcceptorStatus(ctx context.Context, storeID uuid.UUID, deviceID string) (*CoinAcceptorStatus, error) {
	exist, err := heartbeat.CheckHeartbeat(d.redisClient, heartbeat.GetStoreIDHeartbeatName(storeID.String()))
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, &StoreNotFoundError{S: fmt.Sprintf("store not found, store_id=%s", storeID)}
	}
	return d.iot.GetCoinAcceptorStatus(ctx, storeID, deviceID)
}
