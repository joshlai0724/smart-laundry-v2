package iotsdk

import (
	"encoding/json"

	"github.com/google/uuid"
)

type event struct {
	StoreID  uuid.UUID `json:"string"`
	DeviceID string    `json:"device_id"`
	Points   int32     `json:"points"`
	State    string    `json:"state"`
	Ts       int64     `json:"ts"`
}

type eventCtrl struct {
	bs       *broadcastService
	handlers map[string]func(messageType3[event])
}

func newEventCtrl(bs *broadcastService) *eventCtrl {
	c := &eventCtrl{bs: bs}
	c.handlers = map[string]func(messageType3[event]){
		"coin-acceptor-status-changed": c.handleCoinAcceptorStatusChanged,
	}

	return c
}

func (c *eventCtrl) handler(bytes []byte) {
	var m3 messageType3[event]
	err := json.Unmarshal(bytes, &m3)
	if err != nil {
		return
	}

	handler, ok := c.handlers[m3.Type]
	if !ok {
		return
	}

	handler(m3)
}

func (c *eventCtrl) handleCoinAcceptorStatusChanged(m3 messageType3[event]) {
	c.bs.pubCoinAcceptorStatusChangedEvent(CoinAcceptorStatusChangedEvent{
		StoreID:  m3.Event.StoreID,
		DeviceID: m3.Event.DeviceID,
		Points:   m3.Event.Points,
		State:    m3.Event.State,
		Ts:       m3.Event.Ts,
	})
}
