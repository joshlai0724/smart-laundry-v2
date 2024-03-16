package token

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Subject   uuid.UUID `json:"subject"`
	IssuedAt  int64     `json:"issued_at"`
	ExpiredAt int64     `json:"expired_at"`
}

func NewPayload(userID uuid.UUID, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	payload := &Payload{
		ID:        tokenID,
		Subject:   userID,
		IssuedAt:  now.UnixMilli(),
		ExpiredAt: now.Add(duration).UnixMilli(),
	}
	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().After(time.UnixMilli(payload.ExpiredAt)) {
		return ErrExpiredToken
	}
	return nil
}
