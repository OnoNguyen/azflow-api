package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func GetTrack() string {
	return "https://azflowresources.blob.core.windows.net/audio/speech.mp3?sp=r&st=2024-06-20T07:58:16Z&se=2024-06-20T15:58:16Z&spr=https&sv=2022-11-02&sr=b&sig=%2Fcp3XkF8N49KxseP0sSoDtD0oUHTtvmb5G4k5rz9ie0%3D"
}

// Tts takes a string input and a voice string and returns the responsed voice as an io.ReadCloser if successful.
func Tts(input string, voice string, outFile *os.File) {
	if voice == "" {
		voice = "onyx"
	}

	url := "https://api.openai.com/v1/audio/speech"
	token := "Bearer " + os.Getenv("OPENAI_API_KEY")

	requestBody, err := json.Marshal(map[string]string{
		"model": "openai-1",
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

	// print error if status code is not 200
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Non-OK HTTP status: %s\nResponse body: %s\n", resp.Status, string(body))
		os.Exit(1)
	}

	// Write the response body to the file
	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		fmt.Println("Error saving file:", err)
		os.Exit(1)
	}

	fmt.Println("File saved successfully.")

}
