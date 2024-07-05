package tts

import (
	"azflow-api/azure"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var (
	AudioFolderName = "audio"
)

func GetTrack() string {
	return "https://azflowresources.blob.core.windows.net/audio/speech.mp3?sp=r&st=2024-06-20T07:58:16Z&se=2024-06-20T15:58:16Z&spr=https&sv=2022-11-02&sr=b&sig=%2Fcp3XkF8N49KxseP0sSoDtD0oUHTtvmb5G4k5rz9ie0%3D"
}

func GetTrackUrls(userId string) ([]string, error) {
	return azure.GetFileUrls(fmt.Sprintf("%s/%s", AudioFolderName, userId))
}

// Tts takes a string input and a voice string and returns the path to the generated speech file.
func Tts(input string, voice string, userId string) (string, error) {
	if voice == "" {
		voice = "onyx"
	}

	url := "https://api.openai.com/v1/audio/speech"
	token := "Bearer " + os.Getenv("OPENAI_API_KEY")

	requestBody, err := json.Marshal(map[string]string{
		"model": "tts-1",
		"input": input,
		"voice": voice,
	})
	if err != nil {
		fmt.Println("Error marshalling request body:", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error creating request:", err)
		os.Exit(1)
	}

	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Non-OK HTTP status: %s\nResponse body: %s\n", resp.Status, string(body))
		os.Exit(1)
	}

	// Get the current timestamp
	timestamp := time.Now().Format("20060102-150405")
	// Create the file name with timestamp
	fileName := fmt.Sprintf("%s.mp3", timestamp)

	speechFilePath := filepath.Join("", fileName)
	outFile, err := os.Create(speechFilePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	defer outFile.Close()
	defer os.Remove(outFile.Name())

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		os.Exit(1)
	}

	fmt.Println("MP3 file successfully saved as", speechFilePath)

	err = azure.UploadFile(fmt.Sprintf("%s/%s", AudioFolderName, userId), fileName, speechFilePath)
	if err != nil {
		return "", err
	}
	return speechFilePath, nil
}
