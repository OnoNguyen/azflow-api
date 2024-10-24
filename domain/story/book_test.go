package story

import (
	"azflow-api/ffmpeg"
	"azflow-api/openai"
	"encoding/json"
	"fmt"
	_ "github.com/flashlabs/rootpath"
	"github.com/joho/godotenv"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	m.Run()
}

func TestRecreateAudiosFromWorkDir(t *testing.T) {
	workDir := "/Users/hungnguyen/src/azflow-api/video/20241016-180220-viet"

	openai.Init()
	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var summaryStruct SummaryStruct
	if err := json.Unmarshal([]byte(meta), &summaryStruct); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create audio for introduction and conclusion
	if err := openai.TextToSpeech(summaryStruct.Introduction, "", filepath.Join(workDir, "0-intro.mp3")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	concFileName := fmt.Sprintf("%d-conc.mp3", len(summaryStruct.MainSummaries)+1)
	if err := openai.TextToSpeech(summaryStruct.Conclusion, "", filepath.Join(workDir, concFileName)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create audios
	for i, mainSummary := range summaryStruct.MainSummaries {
		err := openai.TextToSpeech(mainSummary, "", filepath.Join(workDir, fmt.Sprintf("%d.mp3", i+1)))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	t.Logf("workDir: %v", workDir)

}

func TestCreateBookSummaryAndImageIdeas(t *testing.T) {
	openai.Init()

	sumStruct, err := CreateBookSummaryAndImageIdeas("Gambler by Billy Walter")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create meta folder
	timestamp := time.Now().Format("20060102-150405")
	workingFolder := filepath.Join("video", timestamp)
	if err := os.MkdirAll(workingFolder, os.ModePerm); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create and save meta
	outFile2, err2 := os.Create(fmt.Sprintf("%s/meta.json", workingFolder))
	if err2 != nil {
		t.Fatalf("expected no error, got %v", err2)
	}
	defer outFile2.Close()

	t.Logf("result: %v", sumStruct)
}

func TestCreateBookSummaryVideo(t *testing.T) {
	openai.Init()

	result, err := CreateBookSummaryVideo("Zero to one")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", result)
}

func TestFfemgExecutor(t *testing.T) {
	workDir := "/Users/hungnguyen/src/azflow-api/video/20241023-181934"
	// create video
	err := ffmpeg.ExecuteScript(workDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
