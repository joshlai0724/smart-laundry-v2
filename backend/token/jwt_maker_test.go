package token

import (
	randomutil "backend/util/random"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestJMTMaker(t *testing.T) {
	maker, err := NewJWTMaker(randomutil.RandomAlphaNumString(32))
	require.NoError(t, err)

	userID := uuid.New()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(userID, duration)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload.ID)
	require.Equal(t, userID, payload.Subject)
	require.InDelta(t, issuedAt.UnixMilli(), payload.IssuedAt, 1000)   // 1000 ms
	require.InDelta(t, expiredAt.UnixMilli(), payload.ExpiredAt, 1000) // 1000 ms
}

func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(randomutil.RandomAlphaNumString(32))
	require.NoError(t, err)

	token, _, err := maker.CreateToken(uuid.New(), -time.Minute)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	payload, err := maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrExpiredToken.Error())
	require.Nil(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := NewPayload(uuid.New(), time.Minute)
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := NewJWTMaker(randomutil.RandomAlphaNumString(32))
	require.NoError(t, err)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.EqualError(t, err, ErrInvalidToken.Error())
	require.Nil(t, payload)
}
