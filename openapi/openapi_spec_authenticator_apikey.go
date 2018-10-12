package openapi

// specAPIKeyAuthenticator defines the behaviour for api key type authenticators (e,g: header/query)
type specAPIKeyAuthenticator interface {
	getContext() interface{}
	prepareAuth(*authContext) error
	getType() authType
}

func createAPIKeyAuthenticator(secDef SpecSecurityDefinition, value string) specAPIKeyAuthenticator {
	switch secDef.apiKey.In {
	case inHeader:
		return apiKeyHeaderAuthenticator{apiKey{secDef.apiKey.Name, value}}
	case inQuery:
		return apiKeyQueryAuthenticator{apiKey{secDef.apiKey.Name, value}}
	}
	return nil
}

type apiKey struct {
	name  string
	value string
}
