package openapi

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApiKeyHeaderAuthenticator(t *testing.T) {
	Convey("Given a name and a value", t, func() {
		name := ""
		value := ""
		Convey("When specV2Analyser method is constructed", func() {
			apiKeyHeaderAuthenticator := &apiKeyHeaderAuthenticator{
				apiKey: apiKey{
					name:  name,
					value: value,
				},
			}
			Convey("Then the apiKeyHeaderAuthenticator should comply with specAPIKeyAuthenticator interface", func() {
				var _ specAPIKeyAuthenticator = apiKeyHeaderAuthenticator
			})
		})
	})
}

func TestApiKeyHeaderAuthenticatorGetContext(t *testing.T) {
	Convey("Given an apiKeyHeaderAuthenticator", t, func() {
		name := "name"
		value := "value"
		apiKeyHeaderAuthenticator := &apiKeyHeaderAuthenticator{
			apiKey: apiKey{
				name:  name,
				value: value,
			},
		}
		Convey("When getContext method is called", func() {
			key := apiKeyHeaderAuthenticator.getContext()
			Convey("Then the key returned  should match the one the apiKeyHeaderAuthenticator was set up with", func() {
				So(key.(apiKey).name, ShouldEqual, apiKeyHeaderAuthenticator.apiKey.name)
				So(key.(apiKey).value, ShouldEqual, apiKeyHeaderAuthenticator.apiKey.value)
			})
		})
	})
}

func TestApiKeyHeaderAuthenticatorGetType(t *testing.T) {
	Convey("Given an apiKeyHeaderAuthenticator", t, func() {
		name := "name"
		value := "value"
		apiKeyHeaderAuthenticator := &apiKeyHeaderAuthenticator{
			apiKey: apiKey{
				name:  name,
				value: value,
			},
		}
		Convey("When getType method is called", func() {
			authType := apiKeyHeaderAuthenticator.getType()
			Convey("Then the authType returned  should be api key header", func() {
				So(authType, ShouldEqual, authTypeAPIKeyHeader)
			})
		})
	})
}

func TestApiKeyHeaderAuthenticatorPrepareAuth(t *testing.T) {
	Convey("Given an apiKeyHeaderAuthenticator", t, func() {
		name := "name"
		value := "value"
		apiKeyHeaderAuthenticator := &apiKeyHeaderAuthenticator{
			apiKey: apiKey{
				name:  name,
				value: value,
			},
		}
		Convey("When prepareAuth method is called with a authContext", func() {
			expectedURL := "http://www.backend.com"
			ctx := &authContext{
				headers: map[string]string{},
				url:     expectedURL,
			}
			err := apiKeyHeaderAuthenticator.prepareAuth(ctx)
			Convey("Then the err returned  should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then the context url should remain the same", func() {
				So(ctx.url, ShouldEqual, expectedURL)
			})
			Convey("Then the context header should be populated with the apiKey info", func() {
				So(ctx.headers, ShouldNotBeEmpty)
				So(ctx.headers[apiKeyHeaderAuthenticator.name], ShouldEqual, apiKeyHeaderAuthenticator.value)
			})
		})
	})
}

func TestApiKeyHeaderAuthenticatorValidate(t *testing.T) {
	testCases := []struct {
		name                      string
		apiKeyHeaderAuthenticator apiKeyHeaderAuthenticator
		expectedError             error
	}{
		{
			name: "validate passes since api key value is populated",
			apiKeyHeaderAuthenticator: apiKeyHeaderAuthenticator{
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
			apiKeyHeaderAuthenticator: apiKeyHeaderAuthenticator{
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
		err := tc.apiKeyHeaderAuthenticator.validate()
		assert.Equal(t, tc.expectedError, err, tc.name)
	}
}
