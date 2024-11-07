package openai

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	"io"
	"net/http"
	"os"
)

func CreateChatCompletion(ctx context.Context, systemMessage string, userMessage string) (string, error) {
	resp, err := Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4o,
		Temperature: 0.6,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("CreateChatCompletion error: %v", err)
	}
	return resp.Choices[0].Message.Content, nil
}

func CreateStructuredChatCompletion[T any](
	ctx context.Context,
	systemMessage string,
	userMessage string,
) (*T, error) {
	var result T
	schema, err := jsonschema.GenerateSchemaForType(result)
	if err != nil {
		return nil, fmt.Errorf("GenerateSchemaForType error: %v", err)
	}

	resp, err := Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4o,
		Temperature: 0,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userMessage,
			},
		},
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openai.ChatCompletionResponseFormatJSONSchema{
				Name:   "generic_response",
				Schema: schema,
				Strict: true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("CreateChatCompletion error: %v", err)
	}

	err = schema.Unmarshal(resp.Choices[0].Message.Content, &result)
	if err != nil {
		return nil, fmt.Errorf("unmarshal schema error: %v", err)
	}

	return &result, nil
}

func CreateImage(prompt string, imagePath string) (string, error) {
	request := openai.ImageRequest{
		Prompt: prompt,
		N:      1,
		Size:   openai.CreateImageSize1024x1792,
		Model:  openai.CreateImageModelDallE3,
	}

	response, err := Client.CreateImage(context.Background(), request)
	if err != nil {
		return "", err
	}

	imageURL := response.Data[0].URL

	// Download the image from the URL
	imageResp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer imageResp.Body.Close()

	if imageResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: %s", imageResp.Status)
	}

	// Create the image file
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
