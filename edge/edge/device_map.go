package edge

import "sync"

type DeviceMapService struct {
	coinAcceptors map[string]*CoinAcceptor
	m1            sync.RWMutex
}

func NewDeviceMapService() *DeviceMapService {
	return &DeviceMapService{coinAcceptors: map[string]*CoinAcceptor{}}
}

func (s *DeviceMapService) AddCoinAcceptor(deviceID string) bool {
	s.m1.Lock()
	defer s.m1.Unlock()
	if _, ok := s.coinAcceptors[deviceID]; ok {
		return false
	}
	s.coinAcceptors[deviceID] = nil
	return true
}

func (s *DeviceMapService) SetCoinAcceptor(deviceID string, coinAcceptor *CoinAcceptor) {
	s.m1.Lock()
	defer s.m1.Unlock()
	s.coinAcceptors[deviceID] = coinAcceptor
}

func (s *DeviceMapService) DeleteCoinAcceptor(deviceID string) {
	s.m1.Lock()
	defer s.m1.Unlock()
	ca := s.coinAcceptors[deviceID]
	if ca != nil {
		ca.Close()
	}
	delete(s.coinAcceptors, deviceID)
}

func (s *DeviceMapService) GetCoinAcceptor(deviceID string) *CoinAcceptor {
	s.m1.Lock()
	defer s.m1.Unlock()
	return s.coinAcceptors[deviceID]
}

func (s *DeviceMapService) GetCoinAcceptorList() []*CoinAcceptor {
	coinAcceptors := []*CoinAcceptor{}
	s.m1.RLock()
	defer s.m1.RUnlock()
	for _, v := range s.coinAcceptors {
		if v != nil {
			coinAcceptors = append(coinAcceptors, v)
		}
	}
	return coinAcceptors
}
