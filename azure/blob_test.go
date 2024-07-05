package azure

import (
	"github.com/Azure/azure-storage-blob-go/azblob"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

func initTest() {
	// Mocked values for testing
	AccountName = "azflowresources"
	AccountKey = "3SF5rYTQYdcHrXJpPEODMwGj/dfPWzI5Dimwr2KmHUVhxUVxm+NGY049xFMsT9e64wM3KwAXcKWl+AStp9BCXw=="
	credential, err := azblob.NewSharedKeyCredential(AccountName, AccountKey)
	if err != nil {
		panic(err)
	}
	Credential = credential
	Pipeline = azblob.NewPipeline(Credential, azblob.PipelineOptions{})
}

func TestUploadFile(t *testing.T) {
	initTest()
	// Create a temporary file for testing
	tmpfile, err := ioutil.TempFile("", "testfile")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write some data to the temporary file
	if _, err := tmpfile.Write([]byte("test data")); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	// Test cases
	tests := []struct {
		name          string
		containerName string
		blobName      string
		filePath      string
		wantErr       bool
	}{
		{
			name:          "successful upload",
			containerName: "audio",
			blobName:      "1a0b3b6f-52c6-4039-afca-d7f93f9ff963/20240705-181528.mp3",
			filePath:      tmpfile.Name(),
			wantErr:       false,
		},
		{
			name:          "file not found",
			containerName: "test",
			blobName:      "testblob",
			filePath:      "nonexistentfile.txt",
			wantErr:       true,
		},
	}

	// Execute test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := UploadFile(tt.containerName, tt.blobName, tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetFileUrls(t *testing.T) {
	initTest()

	// Mocked container name for testing
	containerName := "audio"
	path := "1a0b3b6f-52c6-4039-afca-d7f93f9ff963"

	// Run the function under test
	urls, err := GetFileUrls(containerName, path)

	// Check for errors
	if err != nil {
		t.Fatalf("GetFileUrls failed: %v", err)
	}

	t.Log(urls)

	// Verify each URL format
	for _, u := range urls {
		parsedURL, err := url.Parse(u)
		if err != nil {
			t.Errorf("Failed to parse URL %s: %v", u, err)
			continue
		}
		if parsedURL.Scheme != "https" {
			t.Errorf("Expected URL scheme to be 'https', got '%s'", parsedURL.Scheme)
		}
		if parsedURL.Query().Get("se") == "" {
			t.Error("Expected SAS token 'se' parameter, got empty")
		}
	}
}
