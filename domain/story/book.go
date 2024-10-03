package story

import "azflow-api/openai"

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

//
//func CreateImageIdeas(content string) (string, error) {
//	messages := []map[string]string{
//		{"role": "system", "content": "summarize this book with less than 4000 words, in the language of user's content"},
//		{"role": "user", "content": title},
//	}
//	summary, err := openai.Chat(messages)
//	if err != nil {
//		return "", err
//	}
//	return summary, nil
//}
