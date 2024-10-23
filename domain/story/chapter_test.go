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

func TestCreateChapterAudiosForOpenAndClose(t *testing.T) {
	openai.Init()
	workDir := "video/20241022-184118/"

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
	//if _, err = openai.CreateImage(fmt.Sprintf("%s. Embed the line: %s into the center of the image taking up roughly half of the image total space.", sumStruct.IntroductionImage, sumStruct.Title), fmt.Sprintf("%s/0-intro.png", workDir)); err != nil {
	//	t.Fatalf("expected no error, got %v", err)
	//}
	if _, err = openai.CreateImage(fmt.Sprintf("%s. Print the following exact quote: '%s', make it stand out and easy to read, make it take up roughly half of the image's total space, in the center.", sumStruct.ConclusionImage, "True progress is achieved through innovation (0 to 1), not merely by copying or scaling existing ideas (1 to n)."), fmt.Sprintf("%s/%d-conc.png", workDir, len(sumStruct.MainSummaries)+1)); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	t.Logf("result: %v", workDir)
}
