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

	content, _ := Chat(messages)

	t.Log(content)
}

func TestCreateImage(t *testing.T) {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Init()

	path, _ := CreateImage("A man finding out solution and throwing table due to overjoy.")

	t.Log(path)
}
