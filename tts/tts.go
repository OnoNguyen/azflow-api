package tts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

func Tts(input string, voice string) (string, error) {
	if voice == "" {
		return "Hello, World!", nil
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

	speechFilePath := filepath.Join("", "speech.mp3")
	outFile, err := os.Create(speechFilePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		os.Exit(1)
	}

	fmt.Println("MP3 file successfully saved as", speechFilePath)

	return speechFilePath, nil
}
