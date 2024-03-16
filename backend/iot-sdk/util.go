package iotsdk

import (
	"context"
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func randomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[r.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func rpc[T any, U any](ctx context.Context, rpcRepo RpcRepo, _type string, request T) (messageType2[U], error) {
	m1 := messageType1[T]{
		Type:    _type,
		CorrID:  uuid.NewString(),
		Request: request,
		Ts1:     time.Now().UnixMilli(),
	}
	m1Json, _ := json.Marshal(m1)
	m2Json, err := rpcRepo.Rpc(ctx, m1.CorrID, m1Json)
	if err != nil {
		return messageType2[U]{}, err
	}

	var m2 messageType2[U]
	if err := json.Unmarshal(m2Json, &m2); err != nil {
		return messageType2[U]{}, err
	}
	if err := m2.Err(); err != nil {
		return messageType2[U]{}, err
	}
	return m2, nil
}
