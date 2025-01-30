package searchfetcher

import "context"

type SearchResultsRetriever interface {
	GetSearchResults(context.Context, SearchQuery) (string, error)
}
