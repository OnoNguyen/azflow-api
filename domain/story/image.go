package story

import "azflow-api/openai"

func GenerateImage(prompt string) (string, error) {
	return openai.ImageGen("Create an image for this idea: " + prompt)
}
