package server

import (
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

// reference: https://github.com/rabbitmq/amqp091-go/blob/main/_examples/consumer/consumer.go

type RbmqServer struct {
	Url      string
	Exchange string
	Key      string
	Handler  func([]byte)

	conn    *amqp.Connection
	channel *amqp.Channel
	err     error
	done    chan struct{}
}

func (s *RbmqServer) ListenAndServe() error {
	conn, err := amqp.Dial(s.Url)
	if err != nil {
		return err
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	err = channel.ExchangeDeclare(
		s.Exchange, // name
		"topic",    // type
		false,      // durable
		false,      // auto-deleted
		false,      // internal
		false,      // noWait
		nil,        // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return err
	}

	queue, err := channel.QueueDeclare(
		"",    // name
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // noWait
		nil,   // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return err
	}

	err = channel.QueueBind(
		queue.Name, // queue name
		s.Key,      // routing key
		s.Exchange, // exchange
		false,
		nil,
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return err
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
		return err
	}

	s.conn, s.channel, s.err = conn, channel, errors.New("the connection was closed unexpectedly")
	s.done = make(chan struct{})

	for d := range deliveries {
		go s.Handler(d.Body)
	}

	close(s.done)
	return s.err
}

func (s *RbmqServer) Shutdown() {
	s.err = nil
	s.channel.Close()
	s.conn.Close()
	<-s.done
}
