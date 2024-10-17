package story

import (
	"azflow-api/ffmpeg"
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
	Introduction      string   `json:"introduction"`
	MainSummaries     []string `json:"main_summaries"`
	Conclusion        string   `json:"conclusion"`
	IntroductionImage string   `json:"intro_image"`
	ConclusionImage   string   `json:"conclusion_image"`
	MainSummaryImages []string `json:"image_ideas"`
}

func CreateBookSummaryAndImageIdeas(title string) (*SummaryStruct, error) {
	return openai.CreateStructuredChatCompletion[SummaryStruct](context.Background(),
		"From the book title create an introduction, a list of different paragraphs of elaborations from the key points of the book, and then a conclusion."+
			" Each paragraph should be less than 1000 words. "+
			" Then create an image idea for the introduction, conclusion, and a list of image ideas for each of the paragraphs in the main summary list, for the purpose of image generation, and the number of images has to match the number of paragraphs.", title)
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

	// create and save meta
	outFile2, err2 := os.Create(fmt.Sprintf("%s/meta.json", videoFolder))
	if err2 != nil {
		return "", err2
	}
	defer outFile2.Close()
	// convert sumStruct to json str and write to file
	formattedJSON, err := json.MarshalIndent(sumStruct, "", "  ") // Indent with two spaces
	if _, err := outFile2.Write(formattedJSON); err != nil {
		return "", err
	}

	// create images and audios for the introduction and conclusion
	_, err = openai.CreateImage(sumStruct.Introduction, fmt.Sprintf("%s/0-intro.png", videoFolder))
	if err != nil {
		return "", err
	}
	_, err = openai.CreateImage(sumStruct.Conclusion, fmt.Sprintf("%s/%d-conc.png", videoFolder, len(sumStruct.MainSummaries)+1))
	if err != nil {
		return "", err
	}

	outFile, err := os.Create(fmt.Sprintf("%s/0-intro.mp3", videoFolder))
	if err != nil {
		return "", err
	}
	openai.Tts(sumStruct.Introduction, "", outFile)
	outFile.Close()
	outFile, err = os.Create(fmt.Sprintf("%s/%d-conc.mp3", videoFolder, len(sumStruct.MainSummaries)+1))
	if err != nil {
		return "", err
	}
	openai.Tts(sumStruct.Conclusion, "", outFile)
	outFile.Close()

	// create images and audios for the summaries
	for i, idea := range sumStruct.MainSummaryImages {
		_, err := openai.CreateImage(idea, fmt.Sprintf("%s/%d.png", videoFolder, i+1))
		if err != nil {
			return "", err
		}

		outFile, err := os.Create(fmt.Sprintf("%s/%d.mp3", videoFolder, i+1))
		if err != nil {
			return "", err
		}
		openai.Tts(sumStruct.MainSummaries[i], "", outFile)
		outFile.Close()
	}

	// create video
	abs, err3 := filepath.Abs(videoFolder)
	if err3 != nil {
		return "", err3
	}
	if err := ffmpeg.ExecuteScript(abs); err != nil {
		return "", err
	}

	return videoFolder, nil
}
