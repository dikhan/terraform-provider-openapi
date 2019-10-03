package openapi

import (
	"fmt"
)

// Api Key Header Auth
type apiRefreshTokenAuthenticator struct {
	apiKey
	refreshTokenURL string
}

func newAPIRefreshTokenAuthenticator(name, refreshToken, refreshTokenURL string) apiRefreshTokenAuthenticator {
	return apiRefreshTokenAuthenticator{
		apiKey: apiKey{
			name:  name,
			value: refreshToken,
		},
		refreshTokenURL: refreshTokenURL,
	}
}

func (a apiRefreshTokenAuthenticator) getContext() interface{} {
	return a.apiKey
}

func (a apiRefreshTokenAuthenticator) getType() authType {
	return authTypeAPIKeyHeader
}

// prepareAPIKeyAuthentication adds to the map the auth header required for apikey header authentication. The url
// remains the same
func (a apiRefreshTokenAuthenticator) prepareAuth(authContext *authContext) error {
	apiKey := a.getContext().(apiKey)

	authorizationHeaderValue := apiKey.value
	fmt.Println(authorizationHeaderValue)
	// TODO: call refresh token API (POST) a.refreshTokenURL including the refresh token passed in as Authorization header with value apiKey.value
	//  Get the access token from the response header Authorization
	//  Strip out the bearer scheme from the header value and return the string

	authContext.headers[apiKey.name] = "access token"
	return nil
}
