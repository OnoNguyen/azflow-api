package handlers

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"log"
)

func HashPassord(password string) {
	pass := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(pass, bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
	}

	fmt.Println(string(hash))
}

func CheckPassword(password string, hash string) {
	pass := []byte(password)
	has := []byte(hash)

	err := bcrypt.CompareHashAndPassword(has, pass)
	if err != nil {
		log.Println(err)
	}
}
