package story

import (
	"azflow-api/ffmpeg"
	"azflow-api/openai"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const VideoWorkDir = "video/zero-to-one-c4/"

func CreateVideoPreview(images []string, contentTrunks []string) (string, error) {
	fmt.Printf("CreateVideoPreview input content trunks %v\n", contentTrunks)

	for i := 0; i < len(contentTrunks); i++ {
		fmt.Printf("CreateAssFile with content %v\n", contentTrunks[i])
		CreateAssFile(VideoWorkDir, contentTrunks[i], i)
	}

	// get absolute path to this directory
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("expected no error, got %v", err)
	}

	// recreate video using ffmpeg
	if err := ffmpeg.ExecuteScript(filepath.Join(currentDir, VideoWorkDir)); err != nil {
		return "", fmt.Errorf("expected no error, got %v", err)
	}

	previewUrl := fmt.Sprintf("http://localhost:8080/%s/output.mp4", "video")

	return previewUrl, nil
}

// CreateAssFile creates an .ass file for the given text and id of the [id].mp3 file in the workDir
// Assuming the [id].mp3 file is in the workDir, it will distribute the text appearance time equally into different trunks of the .mp3 file
func CreateAssFile(workDir string, text string, id int) error {
	// read [id].mp3 file from workDir
	fileContent, err := os.ReadFile(filepath.Join(workDir, fmt.Sprintf("%d.mp3", id)))
	if err != nil {
		return fmt.Errorf("expected no error, got %v", err)
	}

	// Get length in seconds of the mp3 content in file
	seconds := float64(len(fileContent)) / 20_000
	wordCount := len(strings.Fields(text))
	secondsEachWord := seconds / float64(wordCount)

	// Create .ass file
	file, err := os.Create(fmt.Sprintf("%s/%d.ass", workDir, id))
	if err != nil {
		return err
	}
	defer file.Close()

	// Write header to .ass file
	header := fmt.Sprintf(`[Script Info]
Title: Zero to One - Chapter 4 - Sentence %d
ScriptType: v4.00+
WrapStyle: 0
ScaledBorderAndShadow: yes
Collisions: Normal
PlayDepth: 0
Timer: 100.0000

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding
Style: Default,Arial,20,&H00FFFFFF,&H000000FF,&H00444444,&H00000000,-1,0,0,0,100,100,0,0,1,1,0,5,10,10,10,1

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
`, id)
	file.WriteString(header)

	splitCount := 3
	// Split text into chunks of up to splitCount words
	words := strings.Fields(text)
	var startTime float64
	for i := 0; i < len(words); i += splitCount {
		// Create a chunk of up to splitCount words
		end := i + splitCount
		if end > len(words) {
			end = len(words)
		}
		chunkWords := words[i:end]

		// Calculate timing for each chunk
		chunkWordCount := len(chunkWords)
		duration := secondsEachWord * float64(chunkWordCount)
		endTime := startTime + duration

		// Convert times to `h:mm:ss.cs` format
		startTimeStr := formatTime(startTime)
		endTimeStr := formatTime(endTime)

		// Build karaoke effect with \k and sky blue color (\1c&HFFB6C1&) for each word
		coolText := ""
		for _, word := range chunkWords {
			wordDuration := secondsEachWord * 100 // convert to centiseconds
			coolText += fmt.Sprintf("{\\1c&H0000FF&\\t(\\1c&H00FFFF&)\\an5\\fscx0\\fscy0\\t(0,%.5f,\\fscx100\\fscy100)}%s ", wordDuration, word)
		}

		// Write dialogue line with fade-in and karaoke effect
		dialogue := fmt.Sprintf("Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n", startTimeStr, endTimeStr, strings.TrimSpace(coolText))
		file.WriteString(dialogue)

		// Update start time for next line
		startTime = endTime
	}

	return nil
}

// Helper function to format time as h:mm:ss.cs
func formatTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := int(seconds) % 3600 / 60
	secondsRemain := seconds - float64(hours*3600+minutes*60)
	return fmt.Sprintf("%d:%02d:%05.2f", hours, minutes, secondsRemain)
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
