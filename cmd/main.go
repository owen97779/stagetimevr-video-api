package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/owen97779/stagetimevr-video-api/internal/searchfetcher"
	"github.com/owen97779/stagetimevr-video-api/internal/urlshortener"
	"github.com/owen97779/stagetimevr-video-api/internal/videoendpoint"
)

type SearchFetcherService struct {
	apis []searchfetcher.SearchResultsRetriever
}

func NewSearchFetcherService(apis []searchfetcher.SearchResultsRetriever) *SearchFetcherService {
	return &SearchFetcherService{apis: apis}
}

func (sfs *SearchFetcherService) callRetrievers(ctx context.Context, query searchfetcher.SearchQuery) (string, error) {
	for _, retriever := range sfs.apis {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		results, err := retriever.GetSearchResults(ctx, query)
		if err == nil {
			return results, nil
		}
		log.Printf("Search Service: Error calling API '%v'. Trying next API.", err)
	}
	return "", fmt.Errorf("all APIs failed to retrieve search results")
}

func searchVideoHandler(searchFetcher *SearchFetcherService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var searchQuery searchfetcher.SearchQuery
		if err := json.NewDecoder(r.Body).Decode(&searchQuery); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		results, err := searchFetcher.callRetrievers(r.Context(), searchQuery)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to retrieve search results: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// VideoService manages the API calls, calls all the retrievers and returns the first successful response
type VideoEndpointService struct {
	apis         []videoendpoint.VideoEndpointRetriever
	urlshortener urlshortener.URLShortener
}

func NewVideoEndpointService(apis []videoendpoint.VideoEndpointRetriever, urlshortener urlshortener.URLShortener) *VideoEndpointService {
	return &VideoEndpointService{apis: apis, urlshortener: urlshortener}
}

func (vs *VideoEndpointService) callRetrievers(ctx context.Context, params videoendpoint.VideoParameters) (string, error) {
	for _, retriever := range vs.apis {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		endpoint, err := retriever.GetVideoEndpoint(ctx, params)
		if err == nil {
			return endpoint, nil
		}
		log.Printf("Video Endpoint Service: Error calling API '%v'. Trying next API.", err)
	}
	return "", fmt.Errorf("all APIs failed to retrieve the video endpoint")
}

func sendShortenedEndpointToClient(w http.ResponseWriter, endpoint string) {
	response := map[string]string{"shortened-url": endpoint}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HTTP handler for video requests
func videoRequestHandler(videoService *VideoEndpointService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var videoParams videoendpoint.VideoParameters
		if err := json.NewDecoder(r.Body).Decode(&videoParams); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		// Check if the short video ID endpoint exists
		exists, err := videoService.urlshortener.CheckIfShortenedURLExists(videoParams.ID)
		if exists {
			sendShortenedEndpointToClient(w, videoParams.ID)
			return
		}
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to check if shortened URL exists: %v", err), http.StatusInternalServerError)
			return
		}

		// Retrieve the long video endpoint
		endpoint, err := videoService.callRetrievers(r.Context(), videoParams)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to retrieve video endpoint: %v", err), http.StatusInternalServerError)
			return
		}

		// Find long url expiry parameter
		expiration, err := urlshortener.ExtractExpiryParameter(endpoint)
		log.Printf("Expiry: %v", expiration)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to extract expiry parameter: %v", err), http.StatusInternalServerError)
		}

		// Create request to shorten the video endpoint
		shortenLinkRequest, err := videoService.urlshortener.ShortLinkURLRequest(endpoint, videoParams.ID, expiration)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create request to shorten video endpoint: %v", err), http.StatusInternalServerError)
			return
		}
		// Shorten the video endpoint
		shortenedEndpoint, err := videoService.urlshortener.ShortenUrl(shortenLinkRequest)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to shorten video endpoint: %v", err), http.StatusInternalServerError)
			return
		}

		sendShortenedEndpointToClient(w, shortenedEndpoint)
	}
}

func main() {
	// Create a new API instance
	video_api_list := make([]videoendpoint.VideoEndpointRetriever, 0)
	api1, err := videoendpoint.API1_Retriever(
		os.Getenv("API1_NAME"),
		os.Getenv("API1_KEY"),
		os.Getenv("API1_URL"))
	if err != nil {
		log.Fatalf("Error creating API1 instance: %v", err)
	}
	video_api_list = append(video_api_list, api1)

	// Create a new URL shortener instance
	urlshortener, err := urlshortener.NewURLShortener(
		os.Getenv("URL_SHORTENER_URL"),
		os.Getenv("URL_SHORTENER_API_KEY"))

	if err != nil {
		log.Fatalf("Error creating URL shortener instance: %v", err)
	}
	videoService := NewVideoEndpointService(video_api_list, *urlshortener)
	http.HandleFunc("/video", videoRequestHandler(videoService))

	search_api_list := make([]searchfetcher.SearchResultsRetriever, 0)
	api3, err := searchfetcher.API3_Retriever(
		os.Getenv("API3_NAME"),
		os.Getenv("API3_KEY"),
		os.Getenv("API3_URL"))
	if err != nil {
		log.Fatalf("Error creating API3 instance: %v", err)
	}
	search_api_list = append(search_api_list, api3)
	searchFetcher := NewSearchFetcherService(search_api_list)
	http.HandleFunc("/search", searchVideoHandler(searchFetcher))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
