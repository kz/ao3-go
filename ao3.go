package ao3

import (
	"net/http"
	"time"
)

const baseURL = "https://archiveofourown.org/"
const defaultTimeout = 5 * time.Second

// AO3Client contains configuration parameters for the package
type AO3Client struct {
	HttpClient    *http.Client
	HtmlSanitizer *Sanitizer
}

// InitAO3Client optionally takes in two parameters:
// - client, the HTTP client to use (especially useful to configure timeouts and
//   use custom clients, for example, with Google Cloud Platform);
// - sanitizerStrength, the sanitization policy for sanitization of blurbs,e tc.
//   (NonePolicy performs no sanitization)
func InitAO3Client(client *http.Client, sanitizationPolicy SanitizationPolicy) (*AO3Client, *AO3Error) {
	if client == nil {
		client = &http.Client{
			Timeout: defaultTimeout,

		}
	}

	sanitizer, err := NewSanitizer(sanitizationPolicy)
	if err != nil {
		return nil, WrapError(http.StatusNotImplemented, err, "unable to create sanitizer")
	}

	return &AO3Client{
		HttpClient:    client,
		HtmlSanitizer: sanitizer,
	}, nil
}
