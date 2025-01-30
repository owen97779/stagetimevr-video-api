package urlshortener

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type URLShortener struct {
	serverEndpoint string
	apiKey         string
}

type shlinkRequest struct {
	LongUrl    string `json:"longUrl"`
	Slug       string `json:"customSlug"`
	ValidUntil string `json:"validUntil"`
	Domain     string `json:"domain"`
}

type shlinkResponse struct {
	ShortUrl string `json:"shortUrl"`
}

func (u *URLShortener) ShortLinkURLRequest(longUrl string, slug string, validUntil string) (shlinkRequest, error) {
	urlShortenerEndpoint, err := url.Parse(u.serverEndpoint)
	if err != nil {
		return shlinkRequest{}, fmt.Errorf("failed to parse urlshortener hostname from env variable: %v", err)
	}
	domain := urlShortenerEndpoint.Hostname()
	return shlinkRequest{
		LongUrl:    longUrl,
		Slug:       slug,
		ValidUntil: validUntil,
		Domain:     domain,
	}, nil
}

// Needs to be in ISO 8601 format
func ExtractExpiryParameter(longURL string) (string, error) {
	parsedURL, err := url.Parse(longURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Get the "expire" parameter
	expireStr := parsedURL.Query().Get("expire")
	if expireStr == "" {
		fmt.Println("Expire parameter not found")
		return "", fmt.Errorf("expire parameter not found")
	}

	// Convert the expire value from string to int64
	var expireUnix int64
	_, err = fmt.Sscanf(expireStr, "%d", &expireUnix)
	if err != nil {
		return "", fmt.Errorf("error converting expire to int64: %w", err)
	}

	// Convert Unix timestamp to ISO 8601 format
	expireTime := time.Unix(expireUnix, 0).UTC().Format(time.RFC3339)
	return expireTime, nil
}

func NewURLShortener(serverEndpoint string, apiKey string) (*URLShortener, error) {
	if serverEndpoint == "" || apiKey == "" {
		return nil, fmt.Errorf("urlshortener server endpoint and api key must be provided")
	}

	return &URLShortener{
		serverEndpoint: serverEndpoint,
		apiKey:         apiKey,
	}, nil
}

func (u *URLShortener) ShortenUrl(shlinkRequest shlinkRequest) (string, error) {
	requestBody, err := json.Marshal(shlinkRequest)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest("POST", u.serverEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", u.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response %d: %v", resp.StatusCode, resp.Body)
	}

	var shlinkResponse shlinkResponse
	if err := json.NewDecoder(resp.Body).Decode(&shlinkResponse); err != nil {
		return "", fmt.Errorf("failed to decode response body: %w", err)
	}

	return shlinkResponse.ShortUrl, nil
}

func (u *URLShortener) CheckIfShortenedURLExists(slug string) (bool, error) {
	req, err := http.NewRequest("GET", u.serverEndpoint+"/"+slug, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-Api-Key", u.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else if resp.StatusCode == http.StatusOK {
		return true, nil
	}

	return false, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
}
