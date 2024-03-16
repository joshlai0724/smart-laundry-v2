package edge

import (
	logutil "edge/util/log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ServiceDiscoveryCtrl struct {
	deviceMapService *DeviceMapService
}

func NewServiceDiscoveryCtrl(s *DeviceMapService) *ServiceDiscoveryCtrl {
	return &ServiceDiscoveryCtrl{deviceMapService: s}
}

func (s *ServiceDiscoveryCtrl) HandleCoinAcceptorBeacon(client mqtt.Client, deviceID string) {
	if !s.deviceMapService.AddCoinAcceptor(deviceID) {
		return
	}

	coinAcceptor := NewCoinAcceptor(client, deviceID)
	s.deviceMapService.SetCoinAcceptor(deviceID, coinAcceptor)
	logutil.GetLogger().Infof("coin acceptor connected, device_id=%s", deviceID)
}
