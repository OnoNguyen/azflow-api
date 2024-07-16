package story

import (
	"azflow-api/azure/storage"
	"azflow-api/openai"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	ContainerName = "audio"
)

func GetAudioUrls(userId string) ([]string, error) {
	return storage.GetFileUrls(ContainerName, userId)
}

// CreateAudio an orchestrator func to do openai tts, upload file, and return path
func CreateAudio(userId string, text string, voice string) (string, error) {
	MinTextLength := 2500
	l := len(text)
	if l < MinTextLength {
		return "", fmt.Errorf("content too short (%d < %d)", l, MinTextLength)
	}

	fileName, filePath, outFile := CreateFile()
	defer outFile.Close()
	// defer os.Remove(outFile.Name())

	openai.Tts(text, voice, outFile)

	azureFilePath := fmt.Sprintf("%s/%s", userId, fileName)
	fmt.Println("Uploading MP3 file to Azure Blob Storage...", azureFilePath)

	err := storage.UploadFile(ContainerName, azureFilePath, filePath)
	if err != nil {
		return "", err
	}

	return azureFilePath, nil
}

func CreateFile() (string, string, *os.File) {
	// Get the current timestamp
	timestamp := time.Now().Format("20060102-150405")
	// Create the file name with timestamp
	fileName := fmt.Sprintf("%s.mp3", timestamp)

	filePath := filepath.Join("", fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	return fileName, filePath, outFile
}
