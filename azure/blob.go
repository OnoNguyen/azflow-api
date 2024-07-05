package azure

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

var (
	// AccountName is the name of the Azure Blob Storage account
	AccountName string
	// AccountKey is the key for the Azure Blob Storage account
	AccountKey string
	Pipeline   pipeline.Pipeline
	Credential *azblob.SharedKeyCredential
)

func Init() {
	AccountName = os.Getenv("AZURE_STORAGE_ACCOUNT")
	AccountKey = os.Getenv("AZURE_STORAGE_KEY")
	// Create a default request pipeline using the account name and account key.
	credential, err := azblob.NewSharedKeyCredential(AccountName, AccountKey)
	if err != nil {
		panic(err)
	}
	Credential = credential
	Pipeline = azblob.NewPipeline(Credential, azblob.PipelineOptions{})
}

// UploadFile uploads a file to an Azure Blob Storage container
func UploadFile(containerName, blobName, filePath string) error {
	// Create a URL to the target container
	URL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", AccountName, containerName))
	if err != nil {
		return fmt.Errorf("failed to parse container URL: %w", err)
	}
	containerURL := azblob.NewContainerURL(*URL, Pipeline)

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

// GetFileUrls gets URLs to all blobs in the specified container with a shared access signature (SAS) for limited read access for one day.
func GetFileUrls(containerName string) ([]string, error) {
	// Create a URL to the target container
	URL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", AccountName, containerName))
	if err != nil {
		return nil, fmt.Errorf("failed to parse container URL: %w", err)
	}
	containerURL := azblob.NewContainerURL(*URL, Pipeline)

	// List blobs in the container
	var blobURLs []string
	for marker := (azblob.Marker{}); marker.NotDone(); {
		listBlob, err := containerURL.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs: %v", err)
		}

		// Generate SAS token and construct URLs for each blob
		for _, blobInfo := range listBlob.Segment.BlobItems {
			blobURL := containerURL.NewBlobURL(blobInfo.Name)

			// Create SAS token
			sasQueryParams, err := azblob.BlobSASSignatureValues{
				Protocol:      azblob.SASProtocolHTTPS, // Users must use HTTPS (not HTTP)
				ExpiryTime:    time.Now().Add(24 * time.Hour),
				ContainerName: containerName,
				BlobName:      blobInfo.Name,
				Permissions:   azblob.BlobSASPermissions{Read: true}.String(),
			}.NewSASQueryParameters(Credential)
			if err != nil {
				return nil, fmt.Errorf("failed to create SAS query parameters: %v", err)
			}

			// Construct SAS URL
			u := blobURL.URL()
			uString := u.String()
			sasURL := fmt.Sprintf("%s?%s", uString, sasQueryParams.Encode())
			blobURLs = append(blobURLs, sasURL)
		}
		marker = listBlob.NextMarker
	}

	return blobURLs, nil
}
