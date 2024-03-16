package passwordutil

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrFailedToHashPassword = errors.New("failed to hash password")

	ErrInvalidPasswordLength = errors.New("invalid password length")
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", ErrFailedToHashPassword
	}
	return string(hashedPassword), err
}

func CheckPassword(password string, hashedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func DoesPasswordMeetRule(password string) error {
	if len(password) < 6 {
		return ErrInvalidPasswordLength
	}
	return nil
}
