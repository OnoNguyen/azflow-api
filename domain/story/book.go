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
	Title             string   `json:"title"`
	KeyPoint          string   `json:"key_point"`
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
			" Each paragraph should be less than 500 words and sounds like a motivational talk, avoid mentioning the book title, the chapter number, and the author name. "+
			" Then create an image idea for the introduction, conclusion, and a list of image ideas for each of the paragraphs in the main summary list, for the purpose of image generation, and the number of images has to match the number of paragraphs.", title)
}

func CreateChapterSummaryAndImageIdeas(title string, chapter int) (*SummaryStruct, error) {
	return openai.CreateStructuredChatCompletion[SummaryStruct](context.Background(),
		"You are a helpful scholar, have been reading a lot of books in the world."+
			"Given a book title and a chapter number in the book, you help find out:"+
			"	1. The title of the chapter in the following form: '[Book Title] Chapter [Chapter Number] [Chapter Title]', for example: 'Zero to One, Chapter 7: Follow the Money'."+
			"	2. An introduction into the chapter."+
			"	3. The key point of the chapter in 1 sentence."+
			"	4. A list of short paragraphs to summarize the chapter. The paragraphs should be less than 200 words. The last paragraph is a real world example illustrating the key point of the chapter."+
			"	5. The conclusion of the chapter."+
			"	6. Don't use author name in all the paragraphs."+
			"Then create an image idea for the introduction, conclusion, and a list of image ideas for each of the paragraphs in the main summary list, for the purpose of image generation. The number of images has to match the number of paragraphs.",
		fmt.Sprintf("%s. Chapter %d", title, chapter))
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
