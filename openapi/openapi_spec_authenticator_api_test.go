package openapi

import (
	"errors"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestApiAuth(t *testing.T) {
	Convey("Given a list of globalSecuritySchemes", t, func() {
		globalSecuritySchemes := &SpecSecuritySchemes{}
		Convey("When apiAuth method is constructed", func() {
			apiAuth := &apiAuth{
				globalSecuritySchemes: globalSecuritySchemes,
			}
			Convey("Then the apiAuth should comply with specAuthenticator interface", func() {
				var _ specAuthenticator = apiAuth
			})
		})
	})
}

func TestAuthRequired(t *testing.T) {
	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_header_auth' and an operation that requires the 'apikey_auth' authentication", t, func() {
		securityPolicyName := "apikey_header_auth"
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: securityPolicyName}}
		url := "https://www.host.com/v1/resource"
		oa := apiAuth{
			globalSecuritySchemes: &SpecSecuritySchemes{
				SpecSecurityScheme{
					Name: securityPolicyName,
				},
			},
		}
		Convey("When authRequired method is called", func() {
			authRequired, operationSecurityPolicies := oa.authRequired(url, operationSecuritySchemes)
			Convey("Then the value returned should be true", func() {
				So(authRequired, ShouldBeTrue)
			})
			Convey("And the name of the security policy 'apikey_header_auth'", func() {
				So(operationSecurityPolicies[0].Name, ShouldEqual, securityPolicyName)
			})
		})
	})

	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_auth' and an operation that DOES NOT require any authentication", t, func() {
		operationSecuritySchemes := SpecSecuritySchemes{}
		url := "https://www.host.com/v1/resource"
		oa := apiAuth{
			globalSecuritySchemes: &SpecSecuritySchemes{},
		}
		Convey("When authRequired method is called", func() {
			authRequired, operationSecurityPolicies := oa.authRequired(url, operationSecuritySchemes)
			Convey("Then the values returned should be false and the name of the security policy should be empty", func() {
				So(authRequired, ShouldBeFalse)
				So(operationSecurityPolicies, ShouldBeEmpty)
			})
		})
	})
}

func TestFetchRequiredAuthenticators(t *testing.T) {
	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_auth' and an operation that requires api key header authentication", t, func() {
		securityPolicyName := "apikey_auth"
		expectedAPIKey := apiKey{
			name:  authorizationHeader,
			value: "superSecretKey",
		}
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
				securityPolicyName: apiKeyHeaderAuthenticator{
					apiKey: expectedAPIKey,
				},
			},
		}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: securityPolicyName}}
		oa := apiAuth{}
		Convey("When fetchRequiredAuthenticators method with a security policy which is also defined in the security definitions", func() {
			authenticators, err := oa.fetchRequiredAuthenticators(operationSecuritySchemes, providerConfig)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the list of authenticators returned should contain the required auths", func() {
				So(authenticators, ShouldNotBeEmpty)
				So(authenticators[0].getContext().(apiKey).name, ShouldEqual, expectedAPIKey.name)
				So(authenticators[0].getContext().(apiKey).value, ShouldEqual, expectedAPIKey.value)
			})

		})
	})

	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_auth' and an operation that requires api key header authentication", t, func() {
		securityPolicyName := "apikey_auth"
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
				securityPolicyName: apiKeyHeaderAuthenticator{
					apiKey: apiKey{
						name:  authorizationHeader,
						value: "superSecretKey",
					},
				},
			},
		}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: "non_defined_security_policy"}}
		oa := apiAuth{}
		Convey("When fetchRequiredAuthenticators method with a security policy which is NOT defined in the security definitions", func() {
			authenticators, err := oa.fetchRequiredAuthenticators(operationSecuritySchemes, providerConfig)
			Convey("Then the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the authenticators returned should be empty", func() {
				So(authenticators, ShouldBeEmpty)
			})
		})
	})
}

func TestPrepareAuth(t *testing.T) {
	testCases := []struct {
		name                          string
		apiAuthenticator              specAuthenticator
		inputURL                      string
		inputOperationSecuritySchemes SpecSecuritySchemes
		inputProviderConfig           providerConfiguration
		expectedHeaders               map[string]string
		expectedURL                   string
		expectedError                 error
	}{
		{
			name:                          "apiAuthenticator set up with no global security schemes and the operation contains a security scheme 'apikey_header_auth' of type apiKeyHeader that matches one defined in the provider configuration (which contains the value)",
			apiAuthenticator:              newAPIAuthenticator(nil),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{SpecSecurityScheme{Name: "apikey_header_auth"}},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"apikey_header_auth": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  authorizationHeader,
							value: "superSecretKey",
						},
					},
				},
			},
			expectedHeaders: map[string]string{authorizationHeader: "superSecretKey"},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   nil,
		},
		{
			name:                          "apiAuthenticator set up with no global security schemes and the operation contains a security scheme 'apikey_query_auth' of type apiKeyQuery that matches one defined in the provider configuration (which contains the value)",
			apiAuthenticator:              newAPIAuthenticator(nil),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{SpecSecurityScheme{Name: "apikey_query_auth"}},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"apikey_query_auth": apiKeyQueryAuthenticator{
						apiKey: apiKey{
							name:  authorizationHeader,
							value: "superSecretKey",
						},
					},
				},
			},
			expectedHeaders: map[string]string{},
			expectedURL:     fmt.Sprintf("https://www.host.com/v1/resource?%s=%s", authorizationHeader, "superSecretKey"),
			expectedError:   nil,
		},
		{
			name:                          "apiAuthenticator set up with no global security schemes and the operation containing multiple mixed security schemes (apikey_header_auth and apikey_query_auth) that matches security definitions defined in the provider configuration (containing their value)",
			apiAuthenticator:              newAPIAuthenticator(nil),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{SpecSecurityScheme{Name: "apikey_header_auth"}, SpecSecurityScheme{Name: "apikey_query_auth"}},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"apikey_header_auth": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  authorizationHeader,
							value: "superSecretKeyInHeader",
						},
					},
					"apikey_query_auth": apiKeyQueryAuthenticator{
						apiKey: apiKey{
							name:  "someQueryParam",
							value: "superSecretKeyInQuery",
						},
					},
				},
			},
			expectedHeaders: map[string]string{authorizationHeader: "superSecretKeyInHeader"},
			expectedURL:     fmt.Sprintf("https://www.host.com/v1/resource?%s=%s", "someQueryParam", "superSecretKeyInQuery"),
			expectedError:   nil,
		},
		{
			name:                          "apiAuthenticator set up with no global security schemes and the operation containing multiple apiKey security schemes (api_key and app_id) that matches security definitions defined in the provider configuration (containing their value)",
			apiAuthenticator:              newAPIAuthenticator(nil),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{SpecSecurityScheme{Name: "api_key"}, SpecSecurityScheme{Name: "app_id"}},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					// provider config keys are always terraform name compliant - snake case
					"api_key": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY",
							value: "superSecretKeyForApiKey",
						},
					},
					"app_id": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-APP-ID",
							value: "superSecretKeyForAppId",
						},
					},
				},
			},
			expectedHeaders: map[string]string{"X-API-KEY": "superSecretKeyForApiKey", "X-APP-ID": "superSecretKeyForAppId"},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   nil,
		},
		{
			name:                          "apiAuthenticator set up with global security schemes that match security definitions defined in the provider configuration and the operation does not override the global security",
			apiAuthenticator:              newAPIAuthenticator(&SpecSecuritySchemes{SpecSecurityScheme{Name: "api_key"}}),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"api_key": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY",
							value: "superSecretKeyForApiKey",
						},
					},
				},
			},
			expectedHeaders: map[string]string{"X-API-KEY": "superSecretKeyForApiKey"},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   nil,
		},
		{
			name:                          "apiAuthenticator set up with global security schemes 'api_key' that match security definitions defined in the provider configuration and the operation overrides the global security schemes with (apiKeyOverride)",
			apiAuthenticator:              newAPIAuthenticator(&SpecSecuritySchemes{SpecSecurityScheme{Name: "api_key"}}),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{SpecSecurityScheme{Name: "apiKeyOverride"}},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"api_key": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY",
							value: "superSecretKeyForApiKey",
						},
					},
					"api_key_override": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY_OVERRIDE",
							value: "superSecretKeyForSpecialOperationApiKey",
						},
					},
				},
			},
			expectedHeaders: map[string]string{"X-API-KEY_OVERRIDE": "superSecretKeyForSpecialOperationApiKey"},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   nil,
		},
		{
			name:                          "apiAuthenticator set up with global security schemes 'apiKey' that are not defined in the provider configuration and the operation does not have any specific security scheme",
			apiAuthenticator:              newAPIAuthenticator(&SpecSecuritySchemes{SpecSecurityScheme{Name: "not_defined_scheme"}}),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"api_key": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY",
							value: "superSecretKeyForApiKey",
						},
					},
				},
			},
			expectedHeaders: map[string]string{},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   errors.New("operation's security policy '{not_defined_scheme}' is not defined, please make sure the swagger file contains a security definition named '{not_defined_scheme}' under the securityDefinitions section"),
		},
		{
			name:                          "apiAuthenticator set up with no global security schemes and the operation having specific security scheme that are not defined in the provider configuration ",
			apiAuthenticator:              newAPIAuthenticator(nil),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{SpecSecurityScheme{Name: "not_defined_scheme"}},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"api_key": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY",
							value: "superSecretKeyForApiKey",
						},
					},
				},
			},
			expectedHeaders: map[string]string{},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   errors.New("operation's security policy '{not_defined_scheme}' is not defined, please make sure the swagger file contains a security definition named '{not_defined_scheme}' under the securityDefinitions section"),
		},
		{
			name:                          "apiAuthenticator set up with global security schemes 'api_key' that match security definitions defined in the provider configuration but it's missing the value",
			apiAuthenticator:              newAPIAuthenticator(&SpecSecuritySchemes{SpecSecurityScheme{Name: "api_key"}}),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"api_key": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY",
							value: "", //The provider was configured with the 'api_key' but its missing the value
						},
						terraformConfigurationName: "api_key",
					},
				},
			},
			expectedHeaders: map[string]string{},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   errors.New("required security definition 'api_key' is missing the value. Please make sure the property 'api_key' is configured with a value in the provider's terraform configuration"),
		},
		{
			name:                          "apiAuthenticator set up with no global security schemes and the operation has a security scheme that matches one security definition defined in the provider configuration but it's missing the value",
			apiAuthenticator:              newAPIAuthenticator(nil),
			inputURL:                      "https://www.host.com/v1/resource",
			inputOperationSecuritySchemes: SpecSecuritySchemes{SpecSecurityScheme{Name: "api_key"}},
			inputProviderConfig: providerConfiguration{
				SecuritySchemaDefinitions: map[string]specAPIKeyAuthenticator{
					"api_key": apiKeyHeaderAuthenticator{
						apiKey: apiKey{
							name:  "X-API-KEY",
							value: "", //The provider was configured with the 'api_key' but its missing the value
						},
						terraformConfigurationName: "api_key",
					},
				},
			},
			expectedHeaders: map[string]string{},
			expectedURL:     "https://www.host.com/v1/resource",
			expectedError:   errors.New("required security definition 'api_key' is missing the value. Please make sure the property 'api_key' is configured with a value in the provider's terraform configuration"),
		},
	}

	for _, tc := range testCases {
		authContext, err := tc.apiAuthenticator.prepareAuth(tc.inputURL, tc.inputOperationSecuritySchemes, tc.inputProviderConfig)
		assert.Equal(t, tc.expectedError, err, tc.name)
		assert.Equal(t, tc.expectedHeaders, authContext.headers, tc.name)
		assert.Equal(t, tc.expectedURL, authContext.url, tc.name)
	}

}
