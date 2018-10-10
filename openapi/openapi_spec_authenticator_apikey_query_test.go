package openapi

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestApiKeyQueryAuthenticator(t *testing.T) {
	Convey("Given a name and a value", t, func() {
		name := ""
		value := ""
		Convey("When specV2Analyser method is constructed", func() {
			apiKeyQueryAuthenticator := &apiKeyQueryAuthenticator{
				apiKey{
					name:  name,
					value: value,
				},
			}
			Convey("Then the apiKeyQueryAuthenticator should comply with specAPIKeyAuthenticator interface", func() {
				var _ specAPIKeyAuthenticator = apiKeyQueryAuthenticator
			})
		})
	})
}

func TestApiKeyQueryAuthenticatorGetContext(t *testing.T) {
	Convey("Given an apiKeyQueryAuthenticator", t, func() {
		name := "name"
		value := "value"
		apiKeyQueryAuthenticator := &apiKeyQueryAuthenticator{
			apiKey: apiKey{
				name:  name,
				value: value,
			},
		}
		Convey("When getContext method is called", func() {
			key := apiKeyQueryAuthenticator.getContext()
			Convey("Then the key returned  should match the one the apiKeyQueryAuthenticator was set up with", func() {
				So(key.(apiKey).name, ShouldEqual, apiKeyQueryAuthenticator.apiKey.name)
				So(key.(apiKey).value, ShouldEqual, apiKeyQueryAuthenticator.apiKey.value)
			})
		})
	})
}

func TestApiKeyQueryAuthenticatorGetType(t *testing.T) {
	Convey("Given an apiKeyQueryAuthenticator", t, func() {
		name := "name"
		value := "value"
		apiKeyHeaderAuthenticator := &apiKeyQueryAuthenticator{
			apiKey: apiKey{
				name:  name,
				value: value,
			},
		}
		Convey("When getType method is called", func() {
			authType := apiKeyHeaderAuthenticator.getType()
			Convey("Then the authType returned  should be api key header", func() {
				So(authType, ShouldEqual, authTypeAPIQuery)
			})
		})
	})
}

func TestApiKeyQueryAuthenticatorPrepareAuth(t *testing.T) {
	Convey("Given an apiKeyQueryAuthenticator", t, func() {
		name := "name"
		value := "value"
		apiKeyQueryAuthenticator := &apiKeyQueryAuthenticator{
			apiKey: apiKey{
				name:  name,
				value: value,
			},
		}
		Convey("When prepareAuth method is called with a authContext", func() {
			expectedURL := "http://www.backend.com"
			expectedHeaders := map[string]string{}
			ctx := &authContext{
				headers: expectedHeaders,
				url:     expectedURL,
			}
			err := apiKeyQueryAuthenticator.prepareAuth(ctx)
			Convey("Then the err returned  should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then the context url should have the query auth", func() {
				So(ctx.url, ShouldEqual, fmt.Sprintf("%s?%s=%s", expectedURL, apiKeyQueryAuthenticator.name, apiKeyQueryAuthenticator.value))
			})
			Convey("Then the context headers should remain the same", func() {
				So(ctx.headers, ShouldBeEmpty)
				So(ctx.headers, ShouldEqual, expectedHeaders)
			})
		})
	})
}
