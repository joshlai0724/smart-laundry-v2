package edge

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

func WsRpc[T any](ctx context.Context, rpcRepo WsRpcRepo, _type string, request T) (MessageType2[WsResponse], error) {
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

func MqttRpc[T any, U any](ctx context.Context, mqttRpcRepo MqttRpcRepo, _type string, request T) (MessageType2[U], error) {
	m1 := MessageType1[T]{
		Type:    _type,
		CorrID:  uuid.NewString(),
		Request: request,
		Ts1:     time.Now().UnixMilli(),
	}
	m1Json, _ := json.Marshal(m1)
	m2Json, err := mqttRpcRepo.Rpc(ctx, m1.CorrID, m1Json)
	if err != nil {
		return MessageType2[U]{}, err
	}

	var m2 MessageType2[U]
	if err := json.Unmarshal(m2Json, &m2); err != nil {
		return MessageType2[U]{}, err
	}
	if err := m2.Err(); err != nil {
		return MessageType2[U]{}, err
	}
	return m2, nil
}
