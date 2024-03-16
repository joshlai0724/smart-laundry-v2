package edge

import (
	"context"
	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type CoinAcceptor struct {
	deviceID    string
	mqttRpcRepo MqttRpcRepo
	registered  bool
}

func NewCoinAcceptor(client mqtt.Client, deviceID string) *CoinAcceptor {
	r := newMqttRpcRepo(client, fmt.Sprintf(CoinAcceptorRequestKeyFmt, deviceID), fmt.Sprintf(CoinAcceptorResponseKeyFmt, deviceID))
	return &CoinAcceptor{deviceID: deviceID, mqttRpcRepo: r, registered: false}
}

func (ca *CoinAcceptor) GetDeviceID() string {
	return ca.deviceID
}

type addPointsRequest struct {
	Amount int32 `json:"amount"`
}

func (ca *CoinAcceptor) AddPoints(ctx context.Context, amount int32) error {
	_, err := MqttRpc[
		addPointsRequest,
		struct{},
	](
		ctx,
		ca.mqttRpcRepo,
		"add-points",
		addPointsRequest{
			Amount: amount,
		},
	)

	return err
}

type getDeviceInfoResponse struct {
	FirmwareVersion string `json:"firmware_version"`
}

func (ca *CoinAcceptor) GetDeviceInfo(ctx context.Context) (*CoinAcceptorInfo, error) {
	m2, err := MqttRpc[
		struct{},
		getDeviceInfoResponse,
	](
		ctx,
		ca.mqttRpcRepo,
		"get-device-info",
		struct{}{},
	)

	if err != nil {
		return nil, err
	}

	return &CoinAcceptorInfo{
		FirmwareVersion: m2.Response.FirmwareVersion,
	}, nil
}

type getDeviceStatusResponse struct {
	Points int32  `json:"points"`
	State  string `json:"state"`
	Ts     int64  `json:"ts"`
}

func (ca *CoinAcceptor) GetDeviceStatus(ctx context.Context) (*CoinAcceptorStatus, error) {
	m2, err := MqttRpc[
		struct{},
		getDeviceStatusResponse,
	](
		ctx,
		ca.mqttRpcRepo,
		"get-device-status",
		struct{}{},
	)

	if err != nil {
		return nil, err
	}

	return &CoinAcceptorStatus{
		Points: m2.Response.Points,
		State:  m2.Response.State,
		Ts:     m2.Response.Ts,
	}, nil
}

func (ca *CoinAcceptor) CheckHealth(ctx context.Context) error {
	_, err := MqttRpc[
		struct{},
		struct{},
	](
		ctx,
		ca.mqttRpcRepo,
		"check-health",
		struct{}{},
	)

	return err
}

func (ca *CoinAcceptor) Blink(ctx context.Context) error {
	_, err := MqttRpc[
		struct{},
		struct{},
	](
		ctx,
		ca.mqttRpcRepo,
		"blink",
		struct{}{},
	)

	return err
}

func (ca *CoinAcceptor) SetRegistered(registered bool) {
	ca.registered = registered
}

func (ca *CoinAcceptor) Registered() bool {
	return ca.registered
}

func (ca *CoinAcceptor) Close() {
	ca.mqttRpcRepo.Close()
}
