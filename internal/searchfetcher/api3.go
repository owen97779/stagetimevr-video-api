package searchfetcher

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type api3 struct {
	Name string
	key  string
	url  url.URL
}

type SearchQuery struct {
	Search string
}

func API3_Retriever(name, key, fullURL string) (SearchResultsRetriever, error) {
	if name == "" || key == "" || fullURL == "" {
		return nil, fmt.Errorf("searc api 3: env variables not set")
	}

	url, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("search api 3: error parsing URL: %v", err)
	}

	return &api3{
		Name: name,
		key:  key,
		url:  *url,
	}, nil
}

func (a *api3) GetSearchResults(ctx context.Context, searchQuery SearchQuery) (string, error) {
	// SearchFetcher is a package that provides functionality to fetch search results from various search engines.
	// Call the API
	apiEndpoint := a.url.String() + "?q=" + searchQuery.Search

	req, err := http.NewRequest("GET", apiEndpoint, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("x-rapidapi-key", a.key)
	req.Header.Add("X-rapidapi-host", a.url.Host)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	return string(body), nil
}
