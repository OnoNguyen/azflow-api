package azure

import (
	"io/ioutil"
	"os"
	"testing"
)

// Mock the necessary Azure Blob Storage components
//type MockBlobURL struct{}
//
//func (m *MockBlobURL) UploadBufferToBlockBlob(ctx context.Context, b []byte, o azblob.UploadToBlockBlobOptions) (*azblob.BlockBlobUploadResponse, error) {
//	// Simulate a successful upload
//	return &azblob.BlockBlobUploadResponse{}, nil
//}

func TestUploadFile(t *testing.T) {
	// Set up environment variables for testing
	os.Setenv("AZURE_STORAGE_ACCOUNT", "azflowresources")
	os.Setenv("AZURE_STORAGE_KEY", "3SF5rYTQYdcHrXJpPEODMwGj/dfPWzI5Dimwr2KmHUVhxUVxm+NGY049xFMsT9e64wM3KwAXcKWl+AStp9BCXw==")

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
			containerName: "test-1",
			blobName:      "testblob",
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
