package story

import (
	"azflow-api/db"
	"github.com/joho/godotenv"
	"log"
	"testing"
)

func TestInsertAudio(t *testing.T) {
	// init db connection
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db.Init()

	// Call the function
	_, err = insertAudio("test@example.com", "ext-id-member-123", "ext-id-audio-124", "testfileName1", "test title")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
