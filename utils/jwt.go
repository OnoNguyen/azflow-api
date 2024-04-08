package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"log"
	"os"
)

func GenerateJWT() string {
	var (
		key []byte
		t   *jwt.Token
		s   string
	)

	key = []byte(os.Getenv("JWT_SECRET"))
	t = jwt.New(jwt.SigningMethodHS256)
	s, err = t.SignedString(key)

	if err != nil {
		log.Println(err)
	}

	return s
}
