package simulator

import (
	"errors"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type CoinAcceptorRepo struct {
	client mqtt.Client

	keyFmt string
}

func NewCoinAcceptorRepo(client mqtt.Client, keyFmt string) *CoinAcceptorRepo {
	return &CoinAcceptorRepo{client: client, keyFmt: keyFmt}
}

func (r *CoinAcceptorRepo) Publish(id string, msg []byte) error {
	token := r.client.Publish(
		fmt.Sprintf(r.keyFmt, id),
		2,
		false,
		msg,
	)
	if ok := token.WaitTimeout(100 * time.Millisecond); !ok {
		return ErrMqttPublishTokenWaitFailed
	}
	return token.Error()
}

var ErrMqttPublishTokenWaitFailed = errors.New("mqtt publish token wait failed")
