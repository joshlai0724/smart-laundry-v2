package iot

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func Rpc[T any](ctx context.Context, rpcRepo RpcRepo, _type string, request T) (MessageType2[WsResponse], error) {
	m1 := MessageType1[T]{
		Type:    _type,
		CorrID:  uuid.NewString(),
		Request: request,
		Ts1:     time.Now().UnixMilli(),
	}
	m1Json, _ := json.Marshal(m1)
	m2, err := rpcRepo.Rpc(ctx, m1.CorrID, m1Json)
	if err != nil {
		return MessageType2[WsResponse]{}, err
	}

	if err := m2.Err(); err != nil {
		return MessageType2[WsResponse]{}, err
	}
	return m2, nil
}
