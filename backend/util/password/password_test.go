package passwordutil

import (
	randomutil "backend/util/random"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPassword(t *testing.T) {
	password := randomutil.RandomAlphaNumString(16)

	hashedPassword1, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)

	err = CheckPassword(password, hashedPassword1)
	require.NoError(t, err)

	wrongPassword := randomutil.RandomAlphaNumString(16)
	err = CheckPassword(wrongPassword, hashedPassword1)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedPassword2, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword1)
	require.NotEqual(t, hashedPassword1, hashedPassword2)
}

func TestDoesPasswordMeetRule(t *testing.T) {
	err := DoesPasswordMeetRule("12345")
	require.EqualError(t, err, ErrInvalidPasswordLength.Error())

	err = DoesPasswordMeetRule("thisisgoodp@ssw0rd")
	require.NoError(t, err)
}
