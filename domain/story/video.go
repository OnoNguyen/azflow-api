package story

import (
	"azflow-api/ffmpeg"
	"azflow-api/openai"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func CreateVideoPreview(images []string, contenTrunks []string) (string, error) {
	fmt.Printf("CreateVideoPreview input content trunks %v\n", contenTrunks)

	return "http://localhost:8080/video/output.mp4", nil
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
	metaFile, err2 := os.Create(fmt.Sprintf("%s/meta1.json", videoFolder))
	if err2 != nil {
		return "", err2
	}
	defer metaFile.Close()

	// convert sumStruct to json str and write to file
	formattedJSON, err := json.MarshalIndent(sumStruct, "", "  ") // Indent with two spaces
	if _, err := metaFile.Write(formattedJSON); err != nil {
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
