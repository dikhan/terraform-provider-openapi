package openapi

// specAPIKeyAuthenticator defines the behaviour for api key type authenticators (e,g: header/query)
type specAPIKeyAuthenticator interface {
	getContext() interface{}
	prepareAuth(*authContext) error
	getType() authType
}

func createAPIKeyAuthenticator(secDef SpecSecurityDefinition, value string) specAPIKeyAuthenticator {
	switch secDef.getAPIKey().In {
	case inHeader:
		// TODO: add tests here
		if secDef.getType() == securityDefinitionAPIKeyRefreshToken {
			return newAPIRefreshTokenAuthenticator(secDef.getAPIKey().Name, secDef.buildValue(value), secDef.getAPIKey().Metadata[refreshTokenURLKey].(string))
		}
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
