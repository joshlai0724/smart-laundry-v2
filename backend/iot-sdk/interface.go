package iotsdk

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type IoT interface {
	SetDefaultTimeout(t time.Duration)
	GetDefaultTimeout() time.Duration
	Close()

	AddPointsToCoinAcceptor(ctx context.Context, storeID uuid.UUID, deviceID string, amount int32) error
	GetCoinAcceptorStatus(ctx context.Context, storeID uuid.UUID, deviceID string) (*CoinAcceptorStatus, error)
	BlinkCoinAcceptor(ctx context.Context, storeID uuid.UUID, deviceID string) error

	SubCoinAcceptorStatusChangedEvent() (ch <-chan CoinAcceptorStatusChangedEvent, cancel func())
}
