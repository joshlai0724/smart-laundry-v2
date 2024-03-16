package iotsdk

func (i *iot) SubCoinAcceptorStatusChangedEvent() (ch <-chan CoinAcceptorStatusChangedEvent, cancel func()) {
	return i.bs.subCoinAcceptorStatusChangedEvent()
}
