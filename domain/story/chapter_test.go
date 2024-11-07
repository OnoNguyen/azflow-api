package story

import (
	"azflow-api/openai"
	"context"
	"encoding/json"
	"fmt"
	_ "github.com/flashlabs/rootpath"
	"os"
	"path/filepath"
	"strings"
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

type ChapterSummary struct {
	Paragraphs []string `json:"paragraphs"`
}

func TestBreakupChapterComponents(t *testing.T) {
	openai.Init()
	if p, e := openai.CreateStructuredChatCompletion[ChapterSummary](context.Background(), "Specify title at the beginning and give real life example at the end without mentioning it's a real life example.", "Summarise zero to one chapter 2. After each line break, put the content text on to a new paragraph in the Paragraphs array."); e != nil {
		t.Fatalf("expected no error, got %v", e)
	} else {
		// create meta folder
		workingFolder := filepath.Join("video", "zero-to-one-c2")
		if err := os.MkdirAll(workingFolder, os.ModePerm); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// create and save meta
		metaPath := fmt.Sprintf("%s/meta.json", workingFolder)
		metaFile, err2 := os.Create(metaPath)
		if err2 != nil {
			t.Fatalf("expected no error, got %v", err2)
		}
		defer func(metaFile *os.File) {
			err := metaFile.Close()
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		}(metaFile)

		if formattedJSON, err := json.MarshalIndent(p, "", "  "); err != nil {
			t.Fatalf("expected no error, got %v", err)
		} else {
			if _, err := metaFile.Write(formattedJSON); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			t.Logf("result: %v", metaPath)
		}
	}
}

type ChapterSummaryCaption struct {
	AssFileContent string `json:"ass_content"`
}

func TestMakeAudios(t *testing.T) {
	openai.Init()

	workDir := "video/zero-to-one-c3/"

	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var cs ChapterSummary
	if err := json.Unmarshal(meta, &cs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// create audios
	for i := 0; i < len(cs.Paragraphs); i++ {
		if err := openai.TextToSpeech(cs.Paragraphs[i], "", filepath.Join(workDir, fmt.Sprintf("%d.mp3", i))); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

}

func TestCreateAssFile(t *testing.T) {
	workDir := "video/zero-to-one-c3/"

	// Read meta.json from metaPath
	meta, err := os.ReadFile(filepath.Join(workDir, "meta.json"))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Unmarshal the JSON string into a struct
	var cs ChapterSummary
	if err := json.Unmarshal(meta, &cs); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for i := 0; i < len(cs.Paragraphs); i++ {
		CreateAssFile(workDir, cs.Paragraphs[i], i)
	}
}

func CreateAssFile(workDir string, text string, id int) error {
	// read 0.mp3 file from workDir
	if file, err := os.ReadFile(filepath.Join(workDir, fmt.Sprintf("%d.mp3", id))); err != nil {
		return fmt.Errorf("expected no error, got %v", err)
	} else {
		// get seconds in second of the mp3 content in file
		seconds := float64(len(file)) / 20_000

		wordCount := len(strings.Fields(text))
		secondsEachWord := seconds / float64(wordCount)

		// Create .ass file
		file, err := os.Create(fmt.Sprintf("%s/%d.ass", workDir, id))
		if err != nil {
			return err
		}
		defer file.Close()

		// Write header to .ass file
		header := fmt.Sprintf(`[Script Info]
Title: Zero to One - Chapter 3 - Paragraph %d
ScriptType: v4.00+
WrapStyle: 0
ScaledBorderAndShadow: yes
Collisions: Normal
PlayDepth: 0
Timer: 100.0000

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00444444,&H00000000,-1,0,0,0,100,100,0,0,1,1,0,5,10,10,10,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
`, id)
		file.WriteString(header)

		// Split text into chunks of up to 2 words
		words := strings.Fields(text)
		var startTime float64
		for i := 0; i < len(words); i += 2 {
			// Create a chunk of up to 2 words
			end := i + 2
			if end > len(words) {
				end = len(words)
			}
			chunkWords := words[i:end]

			// Calculate timing for each chunk
			chunkWordCount := len(chunkWords)
			duration := secondsEachWord * float64(chunkWordCount)
			endTime := startTime + duration

			// Convert times to `h:mm:ss.cs` format
			startTimeStr := formatTime(startTime)
			endTimeStr := formatTime(endTime)

			// Build karaoke effect with \k tag for each word
			karaokeText := ""
			for _, word := range chunkWords {
				wordDuration := int(secondsEachWord * 100) // convert to centiseconds
				karaokeText += fmt.Sprintf("{\\k%d}%s ", wordDuration, word)
			}

			// Write dialogue line with fade-in and karaoke effect
			dialogue := fmt.Sprintf("Dialogue: 0,%s,%s,Default,,0,0,0,,{\\fad(200,0)}%s\n", startTimeStr, endTimeStr, strings.TrimSpace(karaokeText))
			file.WriteString(dialogue)

			// Update start time for next line
			startTime = endTime
		}
	}
	return nil
}

// Helper function to format time as h:mm:ss.cs
func formatTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := int(seconds) % 3600 / 60
	secondsRemain := seconds - float64(hours*3600+minutes*60)
	return fmt.Sprintf("%d:%02d:%05.2f", hours, minutes, secondsRemain)
}

func TestCreateChapterMeta(t *testing.T) {

	// create meta folder
	workingFolder := filepath.Join("video", "zero-to-one-c3")
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
	sumStruct, err := CreateChapterSummaryAndImageIdeas("Zero To One by Peter Thiel", 3)
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
