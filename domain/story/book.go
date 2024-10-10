package story

import (
	"azflow-api/ffmeg"
	"azflow-api/openai"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func CreateBookSummary(title string) (string, error) {
	messages := []map[string]string{
		{"role": "system", "content": "summarize this book with less than 4000 words, in the language of user's content"},
		{"role": "user", "content": title},
	}
	summary, err := openai.Chat(messages)
	if err != nil {
		return "", err
	}
	return summary, nil
}

type SummaryStruct struct {
	BookSummary string `json:"book_summary"`
	ImageIdeas  []struct {
		IdeaDescription string `json:"idea_description"`
	}
}

func CreateBookSummaryAndImageIdeas(title string) (*SummaryStruct, error) {

	return openai.CreateStructuredChatCompletion[SummaryStruct](context.Background(),
		"Summarize this book title with less than 4000 words,"+
			"then create a list of image ideas for the first main points of the summary"+
			"then create a list of images for the corresponding ideas", title)
}

func CreateBookSummaryVideo(title string) (string, error) {
	sumStruct, err := CreateBookSummaryAndImageIdeas(title)
	if err != nil {
		return "", err
	}

	// create video folder
	timestamp := time.Now().Format("20060102-150405")
	videoFolder := filepath.Join("video", timestamp)
	if err := os.MkdirAll(videoFolder, os.ModePerm); err != nil {
		return "", err
	}

	// create images
	for i, idea := range sumStruct.ImageIdeas {
		_, err := openai.CreateImage(idea.IdeaDescription, fmt.Sprintf("%s/%d.png", videoFolder, i+1))
		if err != nil {
			return "", err
		}
	}

	// create and save audio
	outFile, err := os.Create(fmt.Sprintf("%s/audio.mp3", videoFolder))
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	openai.Tts(sumStruct.BookSummary, "", outFile)

	// create and save meta
	outFile2, err2 := os.Create(fmt.Sprintf("%s/meta.json", videoFolder))
	if err2 != nil {
		return "", err2
	}
	defer outFile2.Close()
	// convert sumStruct to json str and write to file
	json, _ := json.Marshal(sumStruct)
	str := fmt.Sprintf("%s", json)
	if _, err := outFile2.WriteString(str); err != nil {
		return "", err
	}

	// create video
	abs, err3 := filepath.Abs(videoFolder)
	if err3 != nil {
		return "", err3
	}
	if err := ffmeg.ExecuteScript(abs); err != nil {
		return "", err
	}

	return videoFolder, nil
}
