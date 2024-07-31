package openai

import (
	"github.com/joho/godotenv"
	"log"
	"testing"
)

func TestChat(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Init()

	messages := []map[string]string{
		{"role": "user", "content": "Hello!"},
	}

	content := Chat(messages)

	t.Log(content)
}
