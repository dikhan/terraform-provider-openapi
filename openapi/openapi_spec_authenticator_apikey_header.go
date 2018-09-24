package openapi

// Api Key Header Auth
type apiKeyHeaderAuthenticator struct {
	apiKey
}

func (a apiKeyHeaderAuthenticator) getContext() interface{} {
	return a.apiKey
}

func (a apiKeyHeaderAuthenticator) getType() authType {
	return authTypeAPIKeyHeader
}

// prepareAPIKeyAuthentication adds to the map the auth header required for apikey header authentication. The url
// remains the same
func (a apiKeyHeaderAuthenticator) prepareAuth(authContext *authContext) error {
	apiKey := a.getContext().(apiKey)
	authContext.headers[apiKey.name] = apiKey.value
	return nil
}
