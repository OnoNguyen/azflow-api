package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var Token string

func Init() {
	Token = "Bearer " + os.Getenv("OPENAI_API_KEY")
}
func GetTrack() string {
	return "https://azflowresources.blob.core.windows.net/audio/speech.mp3?sp=r&st=2024-06-20T07:58:16Z&se=2024-06-20T15:58:16Z&spr=https&sv=2022-11-02&sr=b&sig=%2Fcp3XkF8N49KxseP0sSoDtD0oUHTtvmb5G4k5rz9ie0%3D"
}

// Tts stands for text to speech
// It sends a POST request to the OpenAI API and takes a string input and a voice string and returns the responded voice as an io.ReadCloser if successful.
func Tts(input string, voice string, outFile *os.File) {
	if voice == "" {
		voice = "onyx"
	}

	url := "https://api.openai.com/v1/audio/speech"

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

	req.Header.Set("Authorization", Token)
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

func Chat(messages []map[string]string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":    "gpt-4o-mini",
		"messages": messages,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// print error if status code is not 200
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Non-OK HTTP status: %s\nResponse body: %s\n", resp.Status, string(body))
	}

	body, err1 := io.ReadAll(resp.Body)
	if err1 != nil {
		return "", err
	}

	var chatCompletion ChatCompletion
	err = json.Unmarshal(body, &chatCompletion)
	if err != nil {
		return "", err
	}

	if len(chatCompletion.Choices) > 0 {
		content := chatCompletion.Choices[0].Message.Content
		fmt.Println("Content:", content)
		fmt.Println("Finish Reason:", chatCompletion.Choices[0].FinishReason)

		return content, nil

	} else {
		return "", fmt.Errorf("no choices found in the response")
	}
}
