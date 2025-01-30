package videoendpoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// The api details are obscured for security reasons

type api1 struct {
	Name string
	key  string
	url  url.URL
}

func (a *api1) GetVideoEndpoint(ctx context.Context, params VideoParameters) (string, error) {
	// Call the API
	apiEndpoint := a.url.String() + "?id=" + params.ID + "&filter=" + "audioandvideo"

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

	// Parse the JSON response
	possibleVideoEndpoints := []ReturnedEndpoint{}
	if err := json.NewDecoder(res.Body).Decode(&possibleVideoEndpoints); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}
	// // Print out the response body
	// body, err := io.ReadAll(res.Body)
	// if err != nil {
	// 	return "", fmt.Errorf("error reading response body: %v", err)
	// }
	// fmt.Println(string(body))
	// fmt.Printf("Endpoints: %v\n", videoEndpoints)

	for _, endpoint := range possibleVideoEndpoints {
		if err := CheckEndpoint(endpoint.Endpoint); err == nil {
			fmt.Printf("Endpoint: %s\n", endpoint)
			return endpoint.Endpoint, nil
		}
	}

	return "", fmt.Errorf("no valid endpoints found")
}

func API1_Retriever(name, key, fullURL string) (VideoEndpointRetriever, error) {
	if name == "" || key == "" || fullURL == "" {
		return nil, fmt.Errorf("video endpoint api 1: env variables not set")
	}

	url, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("video endpoint api 1: error parsing URL: %v", err)
	}

	return &api1{
		Name: name,
		key:  key,
		url:  *url,
	}, nil
}
