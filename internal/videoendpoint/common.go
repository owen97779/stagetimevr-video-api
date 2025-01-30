package videoendpoint

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type VideoParameters struct {
	ID string `json:"id"`
}

type ReturnedEndpoint struct {
	Endpoint string `json:"url"`
}

type VideoEndpointRetriever interface {
	GetVideoEndpoint(context.Context, VideoParameters) (string, error)
}

func CheckEndpoint(endpoint string) error {
	// Step 1: Parse the URL to check if it is valid
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("invalid URL: %v", err)
	}

	// Optional: Ensure the scheme is HTTP or HTTPS
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	// Step 2: Make a HEAD request to check reachability
	client := http.Client{
		Timeout: 5 * time.Second, // Set a timeout for the request
	}

	resp, err := client.Head(endpoint)
	if err != nil {
		return fmt.Errorf("failed to reach endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Step 3: Check if the status code indicates success
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("endpoint returned non-success status code: %d", resp.StatusCode)
	}

	// If all checks pass, the endpoint is valid and reachable
	return nil
}
