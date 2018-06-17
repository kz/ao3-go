package ao3

import (
	"net/http"
	"time"
)

const baseURL = "https://archiveofourown.org/"
const defaultTimeout = 5 * time.Second

// AO3Client contains configuration parameters for the package
type AO3Client struct {
	HttpClient *http.Client
}

func InitAO3Client(client *http.Client) *AO3Client {
	if client == nil {
		client = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	return &AO3Client{
		HttpClient: client,
	}
}
