package openapi

import "fmt"

// Api Key Query Auth
type apiKeyQueryAuthenticator struct {
	apiKey
}

func (a apiKeyQueryAuthenticator) getContext() interface{} {
	return a.apiKey
}

func (a apiKeyQueryAuthenticator) getType() authType {
	return authTypeAPIQuery
}

// prepareAPIKeyAuthentication updates the url to insert the query api auth values. The map returned is not
// populated in this case as the auth is done via query parameters. However, having the ability to return the map
// provides the opportunity to inject some headers if needed.
func (a apiKeyQueryAuthenticator) prepareAuth(authContext *authContext) error {
	apiKey := a.getContext().(apiKey)
	authContext.url = fmt.Sprintf("%s?%s=%s", authContext.url, apiKey.name, apiKey.value)
	return nil
}
