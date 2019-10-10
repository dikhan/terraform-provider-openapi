package openapi

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dikhan/http_goclient"

	"github.com/stretchr/testify/assert"
)

func Test_ApiKeyRefreshTokenAuthenticator_Successfully_Prepares_Authorization(t *testing.T) {
	accessTokenExpectedReturn := `eyKT[...]er.eyJ[...]IUh.eyd[...]BvR`
	fakeRefreshToken := `eyJ[...]RW.eyJ[...]WQi.eyd[...]SWr`
	accessTokenFakeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fakeRefreshToken, r.Header.Get("my_fancy_name"))
		w.Header().Add(authorizationHeader, accessTokenExpectedReturn)
	}))

	refreshTokenAuthenticator := newAPIRefreshTokenAuthenticator("my_fancy_name", fakeRefreshToken, accessTokenFakeServer.URL)

	t.Run("happy path -- Successful AuthContext is populated with an Access Token when the authContext have no headers map", func(t *testing.T) {
		ctx := &authContext{}
		err := refreshTokenAuthenticator.prepareAuth(ctx)

		assert.NoError(t, err)
		assert.Equal(t, accessTokenExpectedReturn, ctx.headers[authorizationHeader])
	})

	t.Run("happy path -- Successful AuthContext is populated with an Access Token", func(t *testing.T) {
		ctx := &authContext{
			headers: map[string]string{},
		}
		err := refreshTokenAuthenticator.prepareAuth(ctx)

		assert.NoError(t, err)
		assert.Equal(t, accessTokenExpectedReturn, ctx.headers[authorizationHeader])
	})

}

func Test_ApiKeyRefreshTokenAuthenticator_Fails_To_Prepare_Authorization(t *testing.T) {
	t.Run("crappy path -- the API Server providing the access token Fails", func(t *testing.T) {
		fakeRefreshToken := `eyJ[...]RW.eyJ[...]WQi.eyd[...]SWr`
		accessTokenBrokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		refreshTokenAuthenticator := newAPIRefreshTokenAuthenticator("my_fancy_name", fakeRefreshToken, accessTokenBrokenServer.URL)
		ctx := &authContext{}
		err := refreshTokenAuthenticator.prepareAuth(ctx)

		assert.Equal(t, err.Error(), fmt.Sprintf("refresh token POST response '%s' is missing the access token", accessTokenBrokenServer.URL))
		assert.Empty(t, ctx.headers[authorizationHeader])
	})

	t.Run("crappy path -- the HTTP Client Fails in PostJSON", func(t *testing.T) {
		httpStub := http_goclient.HttpClientStub{
			Error: errors.New("postJSON failed"),
		}

		refreshTokenAuthenticator := apiRefreshTokenAuthenticator{
			httpClient: &httpStub,
		}
		ctx := &authContext{}
		err := refreshTokenAuthenticator.prepareAuth(ctx)
		assert.EqualError(t, err, "postJSON failed")
	})
}
