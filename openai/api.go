package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var (
	Token  string
	Client *openai.Client
)

func Init() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	Token = "Bearer " + apiKey
	Client = openai.NewClient(apiKey)
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

// TextToSpeech sends a POST request to the OpenAI API and takes a string input and a voice string and returns the responded voice as an io.ReadCloser if successful.
func TextToSpeech(input string, voice string, outFilePath string) error {
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
		return err
	}

	req, err1 := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err1 != nil {
		return err1
	}

	req.Header.Set("Authorization", Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err2 := client.Do(req)
	if err2 != nil {
		return err2
	}
	defer resp.Body.Close()

	// print error if status code is not 200
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Non-OK HTTP status: %s\nResponse body: %s\n", resp.Status, string(body))
	}

	outFile, err3 := os.Create(outFilePath)
	if err3 != nil {
		return err3
	}

	// Write the response body to the file
	if _, err = io.Copy(outFile, resp.Body); err != nil {
		return err
	}

	fmt.Printf("File saved successfully: %s", outFilePath)
	return nil

}

// ImageGen generates an image using the DALL-E API and saves it into the images folder
func ImageGen(prompt string) (string, error) {
	url := "https://api.openai.com/v1/images/generations"

	requestBody, err := json.Marshal(map[string]interface{}{
		"model":  "dall-e-3",
		"prompt": prompt,
		"n":      1,
		"size":   "1024x1024",
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

	// Check for non-200 status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Non-OK HTTP status: %s\nResponse body: %s\n", resp.Status, string(body))
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Extract the image URL from the response
	data := result["data"].([]interface{})
	imageURL := data[0].(map[string]interface{})["url"].(string)

	// Download the image from the URL
	imageResp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	if imageResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: %s", imageResp.Status)
	}

	// Create the images folder if it doesn't exist
	imageDir := "images"
	if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
		return "", err
	}

	// Create the image file
	imagePath := filepath.Join(imageDir, sanitizeFilename(prompt)+".png")
	file, err := os.Create(imagePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Save the image to the file
	_, err = io.Copy(file, imageResp.Body)
	if err != nil {
		return "", err
	}

	return imagePath, nil
}

// sanitizeFilename ensures the file name is valid by removing invalid characters
func sanitizeFilename(filename string) string {
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}
	return filename
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
