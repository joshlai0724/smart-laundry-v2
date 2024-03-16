package edge

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttRpcRepo interface {
	Rpc(ctx context.Context, corrID string, bytes []byte) ([]byte, error)
	SetTimeout(t time.Duration)
	GetTimeout() time.Duration
	Close()
}

var _ (MqttRpcRepo) = (*mqttRpcRepo)(nil)

type mqttRpcRepo struct {
	corrTable map[string]chan []byte
	m1        sync.Mutex

	timeout time.Duration
	m2      sync.RWMutex

	client      mqtt.Client
	requestKey  string
	responseKey string
}

func newMqttRpcRepo(client mqtt.Client, requestKey, responseKey string) *mqttRpcRepo {
	r := mqttRpcRepo{
		corrTable:   map[string]chan []byte{},
		timeout:     5 * time.Second,
		client:      client,
		requestKey:  requestKey,
		responseKey: responseKey,
	}
	r.client.Subscribe(r.responseKey, 2, r.handleResponse)
	return &r
}

func (r *mqttRpcRepo) Rpc(ctx context.Context, corrID string, bytes []byte) ([]byte, error) {
	ch := make(chan []byte, 1)
	r.m1.Lock()
	r.corrTable[corrID] = ch
	r.m1.Unlock()

	token := r.client.Publish(r.requestKey, 2, false, bytes)
	if ok := token.WaitTimeout(100 * time.Millisecond); !ok {
		return nil, ErrMqttPublishTokenWaitFailed
	}

	if err := token.Error(); err != nil {
		return nil, err
	}

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(r.GetTimeout()):
	case <-ctx.Done():
	}

	r.m1.Lock()
	delete(r.corrTable, corrID)
	r.m1.Unlock()
	return []byte{}, ErrRPCRequestTimeout
}

// SetTimeout is used to set the timeout of RPC.
func (r *mqttRpcRepo) SetTimeout(t time.Duration) {
	r.m2.Lock()
	defer r.m2.Unlock()
	r.timeout = t
}

// GetTimeout is used to get the timeout of RPC.
func (r *mqttRpcRepo) GetTimeout() time.Duration {
	r.m2.RLock()
	defer r.m2.RUnlock()
	return r.timeout
}

func (r *mqttRpcRepo) handleResponse(client mqtt.Client, msg mqtt.Message) {
	keys := strings.Split(msg.Topic(), "/")
	if len(keys) != 6 {
		return
	}
	corrID := keys[5]
	r.m1.Lock()
	defer r.m1.Unlock()
	if ch, exist := r.corrTable[corrID]; exist {
		ch <- msg.Payload()
	}
	delete(r.corrTable, corrID)
}

func (r *mqttRpcRepo) Close() {
	r.client.Unsubscribe(r.responseKey)
}

var ErrMqttPublishTokenWaitFailed = errors.New("mqtt publish token wait failed")
