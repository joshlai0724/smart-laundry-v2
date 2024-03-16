package simulator

import (
	"encoding/json"
	"time"
)

type CoinAcceptorInfo struct {
	FirmwareVersion string
}

type CoinAcceptorStatus struct {
	Points int32
	State  string
}

type CoinAcceptor struct {
	deviceID  string
	info      CoinAcceptorInfo
	status    CoinAcceptorStatus
	eventRepo *CoinAcceptorRepo
}

func NewCoinAcceptor(deviceID string, info CoinAcceptorInfo, status CoinAcceptorStatus) *CoinAcceptor {
	return &CoinAcceptor{deviceID: deviceID, info: info, status: status}
}

func (c *CoinAcceptor) AddPoints(amount int32) {
	c.status.Points += amount

	if c.eventRepo != nil {
		m3 := MessageType3{
			Type: "device-status-changed",
			Event: struct {
				DeviceID string `json:"device_id"`
				Points   int32  `json:"points"`
				State    string `json:"state"`
				Ts       int64  `json:"ts"`
			}{DeviceID: c.deviceID, Points: c.status.Points, State: c.status.State, Ts: time.Now().UnixMilli()},
			Ts3: time.Now().UnixMilli(),
		}
		j, _ := json.Marshal(m3)
		c.eventRepo.Publish(c.deviceID, j)
	}
}

func (c *CoinAcceptor) GetDeviceStatus() (CoinAcceptorStatus, int64) {
	return c.status, time.Now().UnixMilli()
}

func (c *CoinAcceptor) GetDeviceInfo() CoinAcceptorInfo {
	return c.info
}

func (c *CoinAcceptor) SetEventRepo(r *CoinAcceptorRepo) {
	c.eventRepo = r
}
