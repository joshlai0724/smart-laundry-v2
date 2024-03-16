package edge

import (
	"context"
	"errors"
	"sync"
	"time"
)

type WsRpcRepo interface {
	Rpc(ctx context.Context, corrID string, bytes []byte) (MessageType2[WsResponse], error)
	SetTimeout(t time.Duration)
	GetTimeout() time.Duration
}

var _ (WsRpcRepo) = (*wsRpcRepo)(nil)

type wsRpcRepo struct {
	corrTable map[string]chan MessageType2[WsResponse]
	m1        sync.Mutex

	timeout time.Duration
	m2      sync.RWMutex

	toClientChan chan []byte
}

func newWsRpcRepo(toClientChan chan []byte) *wsRpcRepo {
	r := wsRpcRepo{
		corrTable:    map[string]chan MessageType2[WsResponse]{},
		timeout:      5 * time.Second,
		toClientChan: toClientChan,
	}
	return &r
}

func (r *wsRpcRepo) Rpc(ctx context.Context, corrID string, bytes []byte) (MessageType2[WsResponse], error) {
	ch := make(chan MessageType2[WsResponse], 1)
	r.m1.Lock()
	r.corrTable[corrID] = ch
	r.m1.Unlock()

	r.toClientChan <- bytes

	select {
	case result := <-ch:
		return result, nil
	case <-time.After(r.GetTimeout()):
	case <-ctx.Done():
	}

	r.m1.Lock()
	delete(r.corrTable, corrID)
	r.m1.Unlock()
	return MessageType2[WsResponse]{}, ErrRPCRequestTimeout
}

// SetTimeout is used to set the timeout of RPC.
func (r *wsRpcRepo) SetTimeout(t time.Duration) {
	r.m2.Lock()
	defer r.m2.Unlock()
	r.timeout = t
}

// GetTimeout is used to get the timeout of RPC.
func (r *wsRpcRepo) GetTimeout() time.Duration {
	r.m2.RLock()
	defer r.m2.RUnlock()
	return r.timeout
}

func (r *wsRpcRepo) handleResponse(bytes []byte, m2 MessageType2[WsResponse]) {
	r.m1.Lock()
	defer r.m1.Unlock()
	if ch, exist := r.corrTable[m2.CorrID]; exist {
		ch <- m2
	}
	delete(r.corrTable, m2.CorrID)
}

var ErrRPCRequestTimeout = errors.New("rpc request timeout")
