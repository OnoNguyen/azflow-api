package story

import (
	"azflow-api/ffmeg"
	"azflow-api/openai"
	_ "github.com/flashlabs/rootpath"
	"github.com/joho/godotenv"
	"log"
	"testing"
)

func TestMain(m *testing.M) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	m.Run()
}

func TestCreateBookSummaryAndImageIdeas(t *testing.T) {
	openai.Init()

	result, err := CreateBookSummaryAndImageIdeas("Gambler by Billy Walter")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", result)
}

func TestCreateBookSummaryVideo(t *testing.T) {
	openai.Init()

	result, err := CreateBookSummaryVideo("Zero to on")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", result)
}

func TestFfemgExecutor(t *testing.T) {
	// create video
	err := ffmeg.ExecuteScript("/Users/hungnguyen/src/azflow-api/video/20241016-180220")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
