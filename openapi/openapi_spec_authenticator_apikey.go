package openapi

// SpecAPIKeyAuthenticator defines the behaviour for api key type authenticators (e,g: header/query)
type SpecAPIKeyAuthenticator interface {
	getContext() interface{}
	prepareAuth(*authContext) error
	getType() authType
}

func CreateAPIKeyAuthenticator(secDef SpecSecurityDefinition, value string) SpecAPIKeyAuthenticator {
	switch secDef.getAPIKey().In {
	case inHeader:
		return newAPIKeyHeaderAuthenticator(secDef.getAPIKey().Name, secDef.buildValue(value))
	case inQuery:
		return newAPIKeyQueryAuthenticator(secDef.getAPIKey().Name, secDef.buildValue(value))
	}
	return nil
}

type apiKey struct {
	name  string
	value string
}
