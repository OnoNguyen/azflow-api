package story

import (
	"azflow-api/azure/storage"
	"azflow-api/db"
	"azflow-api/openai"
	"context"
	"fmt"
	"github.com/georgysavva/scany/v2/pgxscan"
	"os"
	"path/filepath"
	"time"
)

var (
	ContainerName = "audio"
)

// GetAudios gets the urls to all audio files for a user if userEmail has value
// otherwise gets the urls to all audio files
func GetAudios(userEmail string) ([]string, error) {
	return storage.GetFileUrls(ContainerName, userEmail)
}

// CreateAudio does openai tts, upload file, and return path
func CreateAudio(memberEmail string, memberExtId string, text string, voice string) (string, error) {
	MinTextLength := 2500
	l := len(text)
	if l < MinTextLength {
		return "", fmt.Errorf("content too short (%d < %d)", l, MinTextLength)
	}

	fileName, filePath, outFile := createFile()
	defer outFile.Close()
	// defer os.Remove(outFile.Name())

	openai.Tts(text, voice, outFile)

	azureFilePath := fmt.Sprintf("%s/%s", memberEmail, fileName)
	fmt.Println("Uploading MP3 file to Azure Blob Storage...", azureFilePath)

	err := storage.UploadFile(ContainerName, azureFilePath, filePath)
	if err != nil {
		return "", err
	}

	// note: audioExtId is azureFilePath for now
	_, err = persistAudio(memberEmail, memberExtId, azureFilePath, fileName)
	if err != nil {
		return "", err
	}

	return azureFilePath, nil
}

// persistAudio inserts or updates audio info into the database
func persistAudio(memberEmail string, memberExtId string, audioExtId string, fileName string) (int, error) {
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
	VALUES ((SELECT id FROM member_data), $3, $4, $4)
	ON CONFLICT (ext_id) DO UPDATE
	SET title = EXCLUDED.title
	RETURNING id;
	`

	var audioID int
	err := pgxscan.Get(ctx, db.Conn, &audioID, query, memberEmail, memberExtId, audioExtId, fileName)
	if err != nil {
		return 0, fmt.Errorf("failed to upsert audio info: %w", err)
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
