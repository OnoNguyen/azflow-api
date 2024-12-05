package story

import (
	"azflow-api/openai"
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/flashlabs/rootpath"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateChatCompletion(t *testing.T) {
	openai.Init()

	if c, e := openai.CreateChatCompletion(context.Background(), "", "summarise zero to one chapter 2"); e != nil {
		t.Fatalf("expected no error, got %v", e)
	} else {
		t.Logf("result: %v", c)
	}
}

func TestMakeAudios(t *testing.T) {
	openai.Init()

	workDir := "video/zero-to-one-c4/"

	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var cs ChapterSummaryStruct
	if err := json.Unmarshal(meta, &cs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create audios
	for i := 0; i < len(cs.Sentences); i++ {
		if err := openai.TextToSpeech(cs.Sentences[i], "", filepath.Join(workDir, fmt.Sprintf("%d.mp3", i))); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

}

func TestCreateAssFile(t *testing.T) {
	workDir := "video/zero-to-one-c4/"

	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var cs ChapterSummaryStruct
	if err := json.Unmarshal(meta, &cs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for i := 0; i < len(cs.Sentences); i++ {
		CreateAssFile(workDir, cs.Sentences[i], i)
	}
}

func TestCreateChapterMeta(t *testing.T) {

	// create meta folder
	workingFolder := filepath.Join("video", "zero-to-one-c4")
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

	// create meta
	openai.Init()
	sumStruct, err := CreateChapterSummary("Zero To One by Peter Thiel", 4)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// convert sumStruct to json str and write to file
	formattedJSON, err := json.MarshalIndent(sumStruct, "", "  ") // Indent with two spaces
	if _, err := metaFile.Write(formattedJSON); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", metaPath)
}

func TestCreateChapterImagesForIntroAndConc(t *testing.T) {
	openai.Init()
	workDir := "video/20241023-181934/"

	// Read meta1.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta1.json"))
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

	t.Logf("result: %v", workDir)
}

func TestCreateChapterAudiosForIntroAndConc(t *testing.T) {
	openai.Init()
	workDir := "video/zero-to-one-c2/"

	// Read meta1.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta1.json"))
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

func TestCreateChapterAudiosForMiddleSummaries(t *testing.T) {
	openai.Init()
	workDir := "video/20241023-181934/"

	// Read meta1.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta1.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var sumStruct SummaryStruct
	if err := json.Unmarshal(meta, &sumStruct); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create audios
	for i, mainSummary := range sumStruct.MainSummaries {
		err := openai.TextToSpeech(mainSummary, "", filepath.Join(workDir, fmt.Sprintf("%d.mp3", i+1)))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	t.Logf("result: %v", workDir)
}
