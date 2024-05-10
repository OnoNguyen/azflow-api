package handlers

import (
	"golang.org/x/crypto/bcrypt"
	"log"
)

func HashPassord(password string) string {
	pass := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)

	if err != nil {
		log.Print(err)
	}

	return string(hash)
}

func CheckPasswordHash(password string, hash string) bool {
	pass := []byte(password)
	has := []byte(hash)

	err := bcrypt.CompareHashAndPassword(has, pass)
	return err == nil
}
