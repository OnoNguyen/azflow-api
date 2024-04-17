package handlers

import (
	"golang.org/x/crypto/bcrypt"
)

func HashPassord(password string) (string, error) {
	pass := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)

	return string(hash), err
}

func CheckPasswordHash(password string, hash string) bool {
	pass := []byte(password)
	has := []byte(hash)

	err := bcrypt.CompareHashAndPassword(has, pass)
	return err == nil
}
