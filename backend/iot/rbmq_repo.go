package iot

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RbmqRepo struct {
	exchange string
	keyFmt   string

	conn    *amqp.Connection
	channel *amqp.Channel
}

func NewRbmqRepo(url, exchange, keyFmt string) (*RbmqRepo, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	err = channel.ExchangeDeclare(
		exchange, // name
		"topic",  // type
		false,    // durable
		false,    // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	return &RbmqRepo{exchange: exchange, keyFmt: keyFmt, conn: conn, channel: channel}, nil
}

func (r *RbmqRepo) Publish(id string, msg []byte) error {
	return r.channel.PublishWithContext(
		context.Background(),
		r.exchange,                // exchange
		fmt.Sprintf(r.keyFmt, id), // routing key
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			Body: msg,
		},
	)
}

func (r *RbmqRepo) Close() {
	r.channel.Close()
	r.conn.Close()
}
