package iotsdk

import (
	"sync"

	"github.com/google/uuid"
)

type broadcaster[T any] struct {
	subs map[string]chan T
	m    sync.RWMutex

	chSize int
}

func newBroadcaster[T any](chSize int) *broadcaster[T] {
	return &broadcaster[T]{
		subs:   make(map[string]chan T),
		chSize: chSize,
	}
}

func (b *broadcaster[T]) Sub() (<-chan T, func()) {
	b.m.Lock()
	defer b.m.Unlock()
	subID := uuid.NewString()
	ch := make(chan T, b.chSize)
	b.subs[subID] = ch
	return ch, func() { b.unsub(subID) }
}

func (b *broadcaster[T]) unsub(subID string) {
	b.m.Lock()
	defer b.m.Unlock()
	ch, ok := b.subs[subID]
	if !ok {
		return
	}
	close(ch)
	delete(b.subs, subID)
}

func (b *broadcaster[T]) Pub(data T) {
	b.m.RLock()
	defer b.m.RUnlock()
	for _, ch := range b.subs {
		ch <- data
	}
}
