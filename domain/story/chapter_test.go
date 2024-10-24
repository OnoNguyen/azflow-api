package story

import (
	"azflow-api/openai"
	"encoding/json"
	"fmt"
	_ "github.com/flashlabs/rootpath"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateChapterSummaryAndImageIdeas(t *testing.T) {
	openai.Init()

	sumStruct, err := CreateChapterSummaryAndImageIdeas("Zero To One by Peter Thiel", 1)
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
	metaPath := fmt.Sprintf("%s/meta.json", workingFolder)
	metaFile, err2 := os.Create(metaPath)
	if err2 != nil {
		t.Fatalf("expected no error, got %v", err2)
	}
	defer metaFile.Close()

	// convert sumStruct to json str and write to file
	formattedJSON, err := json.MarshalIndent(sumStruct, "", "  ") // Indent with two spaces
	if _, err := metaFile.Write(formattedJSON); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", metaPath)
}

func TestCreateChapterImagesForOpenAndClose(t *testing.T) {
	openai.Init()
	workDir := "video/20241023-181934/"

	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var sumStruct SummaryStruct
	if err := json.Unmarshal(meta, &sumStruct); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// create images for the introduction and conclusion
	if _, err = openai.CreateImage(sumStruct.IntroductionImage, fmt.Sprintf("%s/0-intro.png", workDir)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if _, err = openai.CreateImage(sumStruct.ConclusionImage, fmt.Sprintf("%s/%d-conc.png", workDir, len(sumStruct.MainSummaries)+1)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	//

	t.Logf("result: %v", workDir)
}

func TestCreateChapterAudiosForOpenAndClose(t *testing.T) {
	openai.Init()
	workDir := "video/20241023-181934/"

	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var sumStruct SummaryStruct
	if err := json.Unmarshal(meta, &sumStruct); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create audio for introduction and conclusion
	if err := openai.TextToSpeech(sumStruct.Introduction, "", filepath.Join(workDir, "0-intro.mp3")); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	concFileName := fmt.Sprintf("%d-conc.mp3", len(sumStruct.MainSummaries)+1)
	if err := openai.TextToSpeech(sumStruct.Conclusion, "", filepath.Join(workDir, concFileName)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", workDir)
}

func TestCreateChapterVideosForOpenAndClose(t *testing.T) {
	openai.Init()
	workDir := "video/20241023-181934/"

	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var sumStruct SummaryStruct
	if err := json.Unmarshal(meta, &sumStruct); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", workDir)
}
