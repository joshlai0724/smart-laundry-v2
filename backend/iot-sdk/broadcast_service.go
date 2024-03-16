package iotsdk

type broadcastService struct {
	coinAcceptorStatusChangedEventBroadcaster *broadcaster[CoinAcceptorStatusChangedEvent]
}

func newBroadcastService() *broadcastService {
	return &broadcastService{
		coinAcceptorStatusChangedEventBroadcaster: newBroadcaster[CoinAcceptorStatusChangedEvent](100),
	}
}

func (s *broadcastService) subCoinAcceptorStatusChangedEvent() (ch <-chan CoinAcceptorStatusChangedEvent, cancel func()) {
	return s.coinAcceptorStatusChangedEventBroadcaster.Sub()
}

func (s *broadcastService) pubCoinAcceptorStatusChangedEvent(event CoinAcceptorStatusChangedEvent) {
	s.coinAcceptorStatusChangedEventBroadcaster.Pub(event)
}
