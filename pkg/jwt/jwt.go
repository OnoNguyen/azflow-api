package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"log"
	"os"
	"time"
)

var (
	SecretKey = []byte(os.Getenv("JWT_SECRET"))
)

func GenerateToken(username string) (string, error) {
	var (
		token    *jwt.Token
		tokenStr string
		err      error
	)

	token = jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = username
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	tokenStr, err = token.SignedString(SecretKey)

	if err != nil {
		log.Fatal("Error in generating token")
		return "", err
	}

	return tokenStr, nil
}

func ParseToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		return username, nil
	}
	return "", err
}
