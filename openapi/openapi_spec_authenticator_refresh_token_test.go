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

	refreshTokenAuthenticator := newAPIRefreshTokenAuthenticator("my_fancy_name", fakeRefreshToken, accessTokenFakeServer.URL, "my_fancy_name")

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

func Test_ApiKeyRefreshT8999999999999okenAuthenticator_Fails_To_Prepare_Authorization(t *testing.T) {
	t.Run("crappy path -- the API Server providing the access token does not return the expected Authorization header containing the access token", func(t *testing.T) {
		fakeRefreshToken := `eyJ[...]RW.eyJ[...]WQi.eyd[...]SWr`
		accessTokenBrokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		refreshTokenAuthenticator := newAPIRefreshTokenAuthenticator("my_fancy_name", fakeRefreshToken, accessTokenBrokenServer.URL, "my_fancy_name")
		ctx := &authContext{}
		err := refreshTokenAuthenticator.prepareAuth(ctx)

		assert.Equal(t, err.Error(), fmt.Sprintf("refresh token POST response '%s' is missing the access token", accessTokenBrokenServer.URL))
		assert.Empty(t, ctx.headers[authorizationHeader])
	})

	t.Run("crappy path -- the API Server providing the access token returns a non expected response status code", func(t *testing.T) {
		fakeRefreshToken := `eyJ[...]RW.eyJ[...]WQi.eyd[...]SWr`
		accessTokenBrokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		refreshTokenAuthenticator := newAPIRefreshTokenAuthenticator("my_fancy_name", fakeRefreshToken, accessTokenBrokenServer.URL, "my_fancy_name")
		ctx := &authContext{}
		err := refreshTokenAuthenticator.prepareAuth(ctx)

		assert.Equal(t, err.Error(), fmt.Sprintf("refresh token POST response '%s' status code '500' not matching expected response status code [200, 204]", accessTokenBrokenServer.URL))
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

func TestAPIRefreshTokenAuthenticatorValidate(t *testing.T) {
	testCases := []struct {
		name                         string
		apiRefreshTokenAuthenticator apiRefreshTokenAuthenticator
		expectedError                error
	}{
		{
			name: "validate passes since api key value is populated",
			apiRefreshTokenAuthenticator: apiRefreshTokenAuthenticator{
				apiKey: apiKey{
					name:  "Authorization",
					value: "some refresh token",
				},
				terraformConfigurationName: "api_token",
			},
			expectedError: nil,
		},
		{
			name: "validate does not pass since api key value is NOT populated/empty",
			apiRefreshTokenAuthenticator: apiRefreshTokenAuthenticator{
				apiKey: apiKey{
					name:  "Authorization",
					value: "",
				},
				terraformConfigurationName: "api_token",
			},
			expectedError: errors.New("required security definition 'api_token' is missing the value. Please make sure the property 'api_token' is configured with a value in the provider's terraform configuration"),
		},
	}

	for _, tc := range testCases {
		err := tc.apiRefreshTokenAuthenticator.validate()
		assert.Equal(t, tc.expectedError, err, tc.name)
	}
}
