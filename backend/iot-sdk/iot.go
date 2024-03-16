package iotsdk

import (
	"time"
)

type iot struct {
	rpcRepo     RpcRepo
	bs          *broadcastService
	eventClient *eventClient
}

var _ IoT = (*iot)(nil)

func newIot(url, appName string) (*iot, error) {
	bs := newBroadcastService()

	eventCtrl := newEventCtrl(bs)

	eventClient, err := newEventClient(eventCtrl.handler, url, appName, exchange, eventKey, 1000)
	if err != nil {
		return nil, err
	}

	rpcRepo, err := NewRpcRepo(url, appName, exchange, requestKey, responseKey, 30*time.Second)
	if err != nil {
		eventClient.Close()
		return nil, err
	}

	return &iot{
		rpcRepo:     rpcRepo,
		bs:          bs,
		eventClient: eventClient,
	}, err
}

func (i *iot) SetDefaultTimeout(t time.Duration) {
	i.rpcRepo.SetTimeout(t)
}

func (i *iot) GetDefaultTimeout() time.Duration {
	return i.rpcRepo.GetTimeout()
}

func (i *iot) Close() {
	i.rpcRepo.Close()
	i.eventClient.Close()
}
