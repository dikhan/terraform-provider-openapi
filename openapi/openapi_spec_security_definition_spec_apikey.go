package openapi

type apiKeyIn string

const (
	inHeader apiKeyIn = "header"
	inQuery  apiKeyIn = "query"
)

type apiKeyMetadataKey string

const (
	refreshTokenURLKey apiKeyMetadataKey = "refreshTokenURL"
)

type specAPIKey struct {
	In       apiKeyIn
	Name     string
	Metadata map[apiKeyMetadataKey]interface{}
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
