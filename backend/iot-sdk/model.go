package iotsdk

import "github.com/google/uuid"

type CoinAcceptorStatus struct {
	Points int32  `json:"points"`
	State  string `json:"state"`
	Ts     int64  `json:"ts"`
}

type CoinAcceptorStatusChangedEvent struct {
	StoreID  uuid.UUID `json:"store_id"`
	DeviceID string    `json:"device_id"`
	Points   int32     `json:"points"`
	State    string    `json:"state"`
	Ts       int64     `json:"ts"`
}
