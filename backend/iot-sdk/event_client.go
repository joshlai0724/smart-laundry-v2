package iotsdk

import (
	logutil "backend/util/log"
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type eventClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel

	cancel context.CancelFunc
}

func newEventClient(
	callback func(body []byte),
	url, appName, exchange, routingKey string,
	prefetchCount int,
) (*eventClient, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	if err := channel.Qos(
		prefetchCount, // prefetch count
		0,             // prefetch size
		false,         // global
	); err != nil {
		channel.Close()
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

	queue, err := channel.QueueDeclare(
		fmt.Sprintf("%s-sdk-event-%s", appName, randomString(15)),
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	if err := channel.QueueBind(
		queue.Name, // queue name
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil,
	); err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	deliveries, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		for {
			select {
			case d, ok := <-deliveries:
				if !ok {
					logutil.GetLogger().Warn("rabbitmq event channel closed")
					return
				}
				callback(d.Body)
			case <-ctx.Done():
				return
			}
		}
	}()

	return &eventClient{
		conn:    conn,
		channel: channel,
		cancel:  cancel,
	}, nil
}

func (c *eventClient) Close() {
	c.channel.Close()
	c.conn.Close()
	c.cancel()
}
