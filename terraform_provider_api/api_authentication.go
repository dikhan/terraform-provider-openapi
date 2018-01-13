package main

import "fmt"

type apiKeyAuthentication interface {
	getAPIKey() apiKey
	prepareAPIKeyAuthentication(string) (map[string]string, string)
}

type apiKey struct {
	name  string
	value string
}

// Api Key Header Auth
type apiKeyHeader struct {
	apiKey
}

func (a apiKeyHeader) getAPIKey() apiKey {
	return a.apiKey
}

// prepareAPIKeyAuthentication adds to the map the auth header required for apikey header authentication. The url
// remains the same
func (a apiKeyHeader) prepareAPIKeyAuthentication(url string) (map[string]string, string) {
	headers := map[string]string{}
	headers[a.getAPIKey().name] = a.getAPIKey().value
	return headers, url
}

// Api Key Query Auth
type apiKeyQuery struct {
	apiKey
}

func (a apiKeyQuery) getAPIKey() apiKey {
	return a.apiKey
}

// prepareAPIKeyAuthentication updates the url to insert the query api auth values. The map returned is not
// populated in this case as the auth is done via query parameters. However, having the ability to return the map
// provides the opportunity to inject some headers if needed.
func (a apiKeyQuery) prepareAPIKeyAuthentication(url string) (map[string]string, string) {
	url = fmt.Sprintf("%s?%s=%s", url, a.getAPIKey().name, a.getAPIKey().value)
	return nil, url
}

func createAPIKeyAuthenticator(apiKeyAuthType, name, value string) apiKeyAuthentication {
	switch apiKeyAuthType {
	case "header":
		return apiKeyHeader{apiKey{name, value}}
	case "query":
		return apiKeyQuery{apiKey{name, value}}
	}
	return nil
}
