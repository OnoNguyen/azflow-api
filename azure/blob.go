package azure

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

// UploadFile uploads a file to an Azure Blob Storage container
func UploadFile(containerName, blobName, filePath string) error {
	var (
		accountName = os.Getenv("AZURE_STORAGE_ACCOUNT")
		accountKey  = os.Getenv("AZURE_STORAGE_KEY")
	)

	// Create a default request pipeline using your account name and account key.
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return fmt.Errorf("failed to create credential: %v", err)
	}
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// Create a URL to the target container
	URL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", accountName, containerName))
	if err != nil {
		return fmt.Errorf("failed to parse container URL: %w", err)
	}
	containerURL := azblob.NewContainerURL(*URL, pipeline)

	// Check if the container exists and create it if not
	_, err = containerURL.Create(context.Background(), azblob.Metadata{}, azblob.PublicAccessNone)
	if err != nil {
		// Check if the error is not due to the container already existing
		if serr, ok := err.(azblob.StorageError); ok && serr.ServiceCode() != azblob.ServiceCodeContainerAlreadyExists {
			return fmt.Errorf("failed to create container: %v", err)
		}
	}

	// Create a URL to the target blob
	blobURL := containerURL.NewBlockBlobURL(blobName)

	fmt.Println("blobURL: ", blobURL)
	// Open the file to upload
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Get the file size and read the file content into a buffer
	fileInfo, _ := file.Stat()
	var size = fileInfo.Size()
	buffer := make([]byte, size)
	file.Read(buffer)

	ctx := context.Background() // This example uses a never-expiring context
	// Upload the file to the Blob Storage
	_, err = azblob.UploadBufferToBlockBlob(ctx, buffer, blobURL, azblob.UploadToBlockBlobOptions{})
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	fmt.Println("File uploaded successfully")
	return nil
}
