package story

import (
	"azflow-api/openai"
	"context"
	"fmt"
	"strings"
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

type ChapterSummaryStruct struct {
	Sentences []string `json:"sentences"`
}

func CreateChapterSummary(title string, chapter int) (*ChapterSummaryStruct, error) {
	if sum, err := openai.CreateChatCompletion(context.Background(), "You are book publisher."+
		" You summarise the book input from user and avoid using author name in the content."+
		" You end the content with a real world example.",
		fmt.Sprintf("Summarise title %s. Chapter %d.", title, chapter)); err != nil {
		return nil, err
	} else {
		// break up sum into an array of sentences, trim empty space before and after splitting.
		sum = strings.TrimSpace(sum)

		ss := strings.Split(sum, ".")

		// trim empty space and remove empty strings
		for i, s := range ss {
			ss[i] = strings.TrimSpace(s)
			if ss[i] == "" {
				ss = append(ss[:i], ss[i+1:]...)
			}
		}

		return &ChapterSummaryStruct{Sentences: ss}, nil
	}
}
