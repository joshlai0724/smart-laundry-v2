package iot

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type edgeEventCtrl struct {
	storeEventRepo *RbmqRepo
	storeID        uuid.UUID
	handlers       map[string]func(MessageType3[WsEvent])
}

func newEdgeEventCtrl(r *RbmqRepo, storeID uuid.UUID) *edgeEventCtrl {
	c := edgeEventCtrl{
		storeEventRepo: r,
		storeID:        storeID,
	}
	c.handlers = map[string]func(MessageType3[WsEvent]){
		"coin-acceptor-status-changed": c.handleCoinAcceptorStatusChangedEvent,
	}
	return &c
}

func (c *edgeEventCtrl) handleEvent(bytes []byte, m3 MessageType3[WsEvent]) {
	handler, ok := c.handlers[m3.Type]
	if !ok {
		return
	}

	handler(m3)
}

func (c *edgeEventCtrl) handleCoinAcceptorStatusChangedEvent(m3 MessageType3[WsEvent]) {
	rbmqM3 := MessageType3[any]{
		Type: "coin-acceptor-status-changed",
		Event: struct {
			StoreID  uuid.UUID `json:"store_id"`
			DeviceID string    `json:"device_id"`
			Points   int32     `json:"points"`
			State    string    `json:"state"`
			Ts       int64     `json:"ts"`
		}{StoreID: c.storeID, DeviceID: m3.Event.DeviceID, Points: m3.Event.Points, State: m3.Event.State, Ts: m3.Event.Ts},
		Ts3: time.Now().UnixMilli(),
	}
	j, _ := json.Marshal(rbmqM3)
	c.storeEventRepo.Publish(c.storeID.String(), j)
}
