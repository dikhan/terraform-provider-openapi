package openapi

type apiKeyIn string

const (
	inHeader apiKeyIn = "header"
	inQuery  apiKeyIn = "query"
)

type specAPIKey struct {
	In   apiKeyIn
	Name string
}

func newAPIKeyHeader(name string) specAPIKey {
	return newAPIKey(name, inHeader)
}

func newAPIKeyQuery(name string) specAPIKey {
	return newAPIKey(name, inQuery)
}

func newAPIKey(name string, in apiKeyIn) specAPIKey {
	return specAPIKey{
		Name: name,
		In:   in,
	}
}
