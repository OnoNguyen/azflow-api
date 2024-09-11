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

// GetAudios gets the urls to all audio files for a user if userEmail has value
// otherwise gets the urls to all audio files
func GetAudios(userEmail string) ([]*TAudio, error) {
	fis, err := storage.GetFileInfos(ContainerName, userEmail)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(fis))
	urls := make(map[string]string, len(fis))
	for _, fi := range fis {
		names = append(names, fi.Name)
		urls[fi.Name] = fi.Url
	}

	infos, err1 := getInfoFromDB(names)
	if err1 != nil {
		return nil, err1
	}

	var audios []*TAudio
	for _, info := range infos {
		name := info.Title
		if name == "" {
			name = info.Name
		}
		audios = append(audios, &TAudio{
			Url:  urls[info.Name],
			Name: name,
			Id:   info.Id,
		})
	}

	return audios, nil
}

// getInfoFromDB gets a list of audio info from db via CTE,
// where input is names array,
// and output is a list of matching titles and ids
func getInfoFromDB(names []string) ([]*TAudioInfo, error) {
	query := `
        WITH name_cte AS (
            SELECT unnest($1::text[]) AS name
        )
        SELECT n.name, COALESCE(a.title, '') AS title, COALESCE(a.id, -1) AS id 
        FROM name_cte n
        JOIN audio a ON n.name = a.ext_id;
    `

	var audios []*TAudioInfo
	err := pgxscan.Select(context.Background(), db.Conn, &audios, query, pq.Array(names))
	if err != nil {
		return nil, err
	}
	return audios, nil
}

// CreateAudio does openai tts, upload file, persist do db, and return path
func CreateAudio(memberEmail string, memberExtId string, text string, voice string, title string) (string, error) {
	MinTextLength := 800
	MaxTextLength := 4000
	l := len(text)
	if l < MinTextLength {
		return "", fmt.Errorf("content too short (%d < %d)", l, MinTextLength)
	}
	if l > MaxTextLength {
		return "", fmt.Errorf("content too long (%d > %d)", l, MaxTextLength)
	}

	fileName, filePath, outFile := createFile()
	defer outFile.Close()
	// defer os.Remove(outFile.Name())

	openai.Tts(text, voice, outFile)

	azureAudioFilePath := fmt.Sprintf("%s/%s", memberEmail, fileName)
	azureCaptionFilePath := azureAudioFilePath + ".txt"

	fmt.Println("Uploading MP3 file to Azure Blob Storage...", azureAudioFilePath)
	err := storage.UploadFile(ContainerName, azureAudioFilePath, filePath)
	if err != nil {
		return "", err
	}

	fmt.Println("Uploading caption file to Azure Blob Storage...", azureCaptionFilePath)
	// Create a temporary file for the caption
	captionFile, err := os.CreateTemp("", "caption-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary caption file: %w", err)
	}
	defer os.Remove(captionFile.Name()) // Clean up the temporary file when done

	// Write the text content to the temporary file
	_, err = captionFile.WriteString(text)
	if err != nil {
		return "", fmt.Errorf("failed to write caption to temporary file: %w", err)
	}
	captionFile.Close() // Close the file to ensure all data is written

	// Upload the caption file to Azure Blob Storage
	err = storage.UploadFile(ContainerName, azureCaptionFilePath, captionFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to upload caption file: %w", err)
	}

	// note: audioExtId is azureAudioFilePath
	_, err = insertAudio(memberEmail, memberExtId, azureAudioFilePath, fileName, title)
	if err != nil {
		return "", err
	}

	return azureAudioFilePath, nil
}

// EditAudio edits audio title
func EditAudio(id int, title string) (*model.Audio, error) {

	var (
		a model.Audio
	)

	err := pgxscan.Get(context.Background(), db.Conn, &a, "UPDATE audio SET title = $1 WHERE id = $2 RETURNING id, title", title, id)
	return &a, err
}

// insertAudio inserts audio info into the database
func insertAudio(memberEmail string, memberExtId string, audioExtId string, fileName string, title string) (int, error) {
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
	RETURNING id;
	`

	var audioID int
	err := pgxscan.Get(ctx, db.Conn, &audioID, query, memberEmail, memberExtId, audioExtId, fileName, title)
	if err != nil {
		return 0, fmt.Errorf("failed to insert audio info: %w", err)
	}
	return audioID, nil
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
