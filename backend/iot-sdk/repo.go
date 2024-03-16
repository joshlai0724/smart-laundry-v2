package iotsdk

import (
	logutil "backend/util/log"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RpcRepo interface {
	Rpc(ctx context.Context, corrID string, msg []byte) ([]byte, error)
	SetTimeout(t time.Duration)
	GetTimeout() time.Duration
	Close()
}

type rpcRepo struct {
	exchange   string
	requestKey string

	conn    *amqp.Connection
	channel *amqp.Channel

	corrTable map[string]chan []byte
	m1        sync.Mutex

	cancel context.CancelFunc

	timeout time.Duration
	m2      sync.RWMutex
}

func NewRpcRepo(url, appName, exchange, requestKey, responseKey string, timeout time.Duration) (*rpcRepo, error) {
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

	queue, err := channel.QueueDeclare(
		fmt.Sprintf("%s-sdk-rpc-%s", appName, randomString(15)), // name
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

	err = channel.QueueBind(
		queue.Name,  // queue name
		responseKey, // routing key
		exchange,    // exchange
		false,
		nil,
	)
	if err != nil {
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

	r := &rpcRepo{
		exchange:   exchange,
		requestKey: requestKey,
		conn:       conn,
		channel:    channel,
		corrTable:  make(map[string]chan []byte),
		cancel:     cancel,
		timeout:    timeout,
	}

	go func(r *rpcRepo) {
		for {
			select {
			case d, ok := <-deliveries:
				if !ok {
					logutil.GetLogger().Warn("rabbitmq rpc channel closed")
					return
				}
				keys := strings.Split(d.RoutingKey, ".")
				if len(keys) != 4 {
					continue
				}
				corrID := keys[3]
				r.m1.Lock()
				if ch, exist := r.corrTable[corrID]; exist {
					ch <- d.Body
				}
				delete(r.corrTable, corrID)
				r.m1.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}(r)

	return r, nil
}

func (r *rpcRepo) Rpc(ctx context.Context, corrID string, msg []byte) ([]byte, error) {
	ch := make(chan []byte, 1)
	r.m1.Lock()
	r.corrTable[corrID] = ch
	r.m1.Unlock()

	if err := r.channel.PublishWithContext(
		ctx,
		r.exchange,   // exchange
		r.requestKey, // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			Body: msg,
		},
	); err != nil {
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
	return nil, ErrRPCRequestTimeout
}

// SetTimeout is used to set the timeout of RPC.
func (r *rpcRepo) SetTimeout(t time.Duration) {
	r.m2.Lock()
	defer r.m2.Unlock()
	r.timeout = t
}

// GetTimeout is used to get the timeout of RPC.
func (r *rpcRepo) GetTimeout() time.Duration {
	r.m2.RLock()
	defer r.m2.RUnlock()
	return r.timeout
}

func (r *rpcRepo) Close() {
	r.channel.Close()
	r.conn.Close()
	r.cancel()
}
