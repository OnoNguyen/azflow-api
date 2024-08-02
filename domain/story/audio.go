package story

import (
	"azflow-api/azure/storage"
	"azflow-api/db"
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
// and output is a list of matching titles
func getInfoFromDB(names []string) ([]*TAudioInfo, error) {
	query := `
        WITH name_cte AS (
            SELECT unnest($1::text[]) AS name
        )
        SELECT n.name, COALESCE(a.title, '') AS title, COALESCE(a.id, -1) AS id 
        FROM name_cte n
        LEFT JOIN audio a ON a.ext_id = n.name;
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

	azureFilePath := fmt.Sprintf("%s/%s", memberEmail, fileName)
	fmt.Println("Uploading MP3 file to Azure Blob Storage...", azureFilePath)

	err := storage.UploadFile(ContainerName, azureFilePath, filePath)
	if err != nil {
		return "", err
	}

	// note: audioExtId is azureFilePath
	_, err = insertAudio(memberEmail, memberExtId, azureFilePath, fileName, title)
	if err != nil {
		return "", err
	}

	return azureFilePath, nil
}

// EditAudio edits audio title
func EditAudio(id int, title string) (string, error) {
	c, err := db.Conn.Exec(context.Background(), "UPDATE audio SET title = $1 WHERE id = $2", title, id)
	return c.String(), err
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
