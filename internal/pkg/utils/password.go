package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func IsValidPassword(password string, hashPassword string) bool {
	byteHash := []byte(hashPassword)

	err := bcrypt.CompareHashAndPassword(byteHash, []byte(password))

	return err == nil
}

func GeneratePasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
