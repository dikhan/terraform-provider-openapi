package openapi

import (
	"fmt"
	"github.com/dikhan/http_goclient"
	"net/http"
)

// TODO: add tests here

// Api Key Header Auth
type apiRefreshTokenAuthenticator struct {
	apiKey
	refreshTokenURL string
	httpClient      http_goclient.HttpClientIface
}

func newAPIRefreshTokenAuthenticator(name, refreshToken, refreshTokenURL string) apiRefreshTokenAuthenticator {
	return apiRefreshTokenAuthenticator{
		apiKey: apiKey{
			name:  name,
			value: refreshToken,
		},
		refreshTokenURL: refreshTokenURL,
		httpClient:      &http_goclient.HttpClient{HttpClient: &http.Client{}},
	}
}

func (a apiRefreshTokenAuthenticator) getContext() interface{} {
	return a.apiKey
}

func (a apiRefreshTokenAuthenticator) getType() authType {
	return authTypeAPIKeyHeader
}

// prepareAuth will send a post request to the refreshTokenURL and get the access token from the response Authorization
// header. Otherwise, it will fail.
func (a apiRefreshTokenAuthenticator) prepareAuth(authContext *authContext) error {
	apiKey := a.getContext().(apiKey)
	headers := map[string]string{apiKey.name: apiKey.value}
	r, err := a.httpClient.PostJson(a.refreshTokenURL, headers, nil, nil)
	if err != nil {
		return err
	}
	accessToken := r.Header.Get(authorizationHeader)
	if accessToken == "" {
		return fmt.Errorf("refresh token POST response '%s' is missing the access token", a.refreshTokenURL)
	}
	authContext.headers[authorizationHeader] = accessToken
	return nil
}
