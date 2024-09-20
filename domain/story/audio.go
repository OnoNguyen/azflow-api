package story

import (
	"azflow-api/azure/storage"
	"azflow-api/db"
	"azflow-api/gql/model"
	"azflow-api/openai"
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/lib/pq"
	"os"
	"path/filepath"
	"time"
)

var (
	ContainerName = "audio"
)

func GetAudio(id int) (*DmAudio, error) {
	var path *string
	err := pgxscan.Get(context.Background(), db.Conn, &path, "SELECT ext_id FROM audio WHERE id = $1", id)
	if err != nil {
		return nil, err
	}

	audio, err1 := GetAudios(*path)

	if err1 != nil {
		return nil, err1
	}

	if len(audio) == 0 {
		return nil, fmt.Errorf("audio not found")
	}

	if len(audio) > 1 {
		return nil, fmt.Errorf("more than one audio found")
	}

	return audio[0], nil

}

// GetAudios retrieves audio files from Azure Storage
// and returns them as DmAudio objects.
// path can be an empty string to retrieve all audio files
// or it can be a path to a specific folder or file
func GetAudios(path string) ([]*DmAudio, error) {
	fis, err := storage.GetFileInfos(ContainerName, path)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(fis))
	urlMap := make(map[string]string, len(fis))
	for _, fi := range fis {
		names = append(names, fi.Name)
		urlMap[fi.Name] = fi.Url
	}

	infos, err1 := getInfoFromDB(names)
	if err1 != nil {
		return nil, err1
	}

	var audios []*DmAudio
	for _, info := range infos {
		title := info.FileName
		if title == "" {
			title = info.Title
		}

		audios = append(audios, &DmAudio{
			Url:           urlMap[info.ExtId],
			TranscriptUrl: urlMap[info.ExtId+".txt"],
			Title:         title,
			Id:            info.Id,
		})
	}

	return audios, nil
}

// getInfoFromDB gets a list of audio info from db via CTE,
// where input is names array,
// and output is a list of matching titles and ids
func getInfoFromDB(names []string) ([]*DbAudio, error) {
	query := `
        WITH name_cte AS (
            SELECT unnest($1::text[]) AS ext_id
        )
        SELECT n.ext_id, COALESCE(a.title, '') AS title, COALESCE(a.id, -1) AS id 
        FROM name_cte n
        JOIN audio a ON n.ext_id = a.ext_id;
    `

	var audios []*DbAudio
	err := pgxscan.Select(context.Background(), db.Conn, &audios, query, pq.Array(names))
	if err != nil {
		return nil, err
	}
	return audios, nil
}

// CreateAudio does openai tts, upload file, persist do db, and return path
func CreateAudio(memberEmail string, memberExtId string, text string, voice string, title string) (*model.Audio, error) {
	MinTextLength := 800
	MaxTextLength := 4000
	l := len(text)
	if l < MinTextLength {
		return nil, fmt.Errorf("content too short (%d < %d)", l, MinTextLength)
	}
	if l > MaxTextLength {
		return nil, fmt.Errorf("content too long (%d > %d)", l, MaxTextLength)
	}

	fileName, filePath, outFile := createFile()
	defer outFile.Close()
	// defer os.Remove(outFile.Title())

	openai.Tts(text, voice, outFile)

	azureAudioFilePath := fmt.Sprintf("%s/%s", memberEmail, fileName)
	azureTranscriptFilePath := azureAudioFilePath + ".txt"

	fmt.Println("Uploading MP3 file to Azure Blob Storage...", azureAudioFilePath)
	err := storage.UploadFile(ContainerName, azureAudioFilePath, filePath)
	if err != nil {
		return nil, err
	}

	err = uploadTranscript(azureTranscriptFilePath, text)
	if err != nil {
		return nil, err
	}

	// note: audioExtId is azureAudioFilePath
	return insertAudio(memberEmail, memberExtId, azureAudioFilePath, fileName, title)
}

func uploadTranscript(azureTranscriptFilePath string, text string) error {
	fmt.Println("Uploading transcript file to Azure Blob Storage...", azureTranscriptFilePath)
	// Create a temporary file for the transcript
	transcriptFile, err := os.CreateTemp("", "transcript-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temporary transcript file: %w", err)
	}
	defer os.Remove(transcriptFile.Name()) // Clean up the temporary file when done

	// Write the text content to the temporary file
	_, err = transcriptFile.WriteString(text)
	if err != nil {
		return fmt.Errorf("failed to write transcript to temporary file: %w", err)
	}
	transcriptFile.Close() // Close the file to ensure all data is written

	// Upload the transcript file to Azure Blob Storage
	err = storage.UploadFile(ContainerName, azureTranscriptFilePath, transcriptFile.Name())
	if err != nil {
		return fmt.Errorf("failed to upload transcript file: %w", err)
	}
	return nil
}

// EditAudio edits audio title
func EditAudio(id int, title string, transcript string) (*model.Audio, error) {

	var audioInfo DbAudio
	err := pgxscan.Get(context.Background(), db.Conn, &audioInfo, "UPDATE audio SET title = $1 WHERE id = $2 RETURNING id, title, ext_id", title, id)

	if err != nil {
		return nil, err
	}

	azureTranscriptFilePath := fmt.Sprintf("%s.txt", audioInfo.ExtId)
	err = uploadTranscript(azureTranscriptFilePath, transcript)
	if err != nil {
		return nil, err
	}

	audio := model.Audio{Title: audioInfo.Title, ID: audioInfo.Id}
	return &audio, err
}

// insertAudio inserts audio info into the database
func insertAudio(memberEmail string, memberExtId string, audioExtId string, fileName string, title string) (*model.Audio, error) {
	ctx := context.Background()

	query := `
	WITH member_cte AS (
		INSERT INTO member (email, ext_id)
		SELECT $1, $2::text
		WHERE NOT EXISTS (
			SELECT 1 FROM member WHERE ext_id = $2::text
		)
		RETURNING id
	),
	member_data AS (
		SELECT id FROM member_cte
		UNION ALL
		SELECT id FROM member WHERE ext_id = $2::text
	)
	INSERT INTO audio (member_id, ext_id, file_name, title)
	VALUES ((SELECT id FROM member_data), $3, $4, $5)
	RETURNING id, title;
	`

	var a model.Audio
	err := pgxscan.Get(ctx, db.Conn, &a, query, memberEmail, memberExtId, audioExtId, fileName, title)
	return &a, err
}

// createFile creates a temporary file
func createFile() (string, string, *os.File) {
	// Get the current timestamp
	timestamp := time.Now().Format("20060102-150405")
	// Create the file name with timestamp
	fileName := fmt.Sprintf("%s.mp3", timestamp)

	filePath := filepath.Join("", fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		os.Exit(1)
	}
	return fileName, filePath, outFile
}
