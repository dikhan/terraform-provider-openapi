package openapi

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPrepareAuth(t *testing.T) {
	Convey("Given a provider configuration containing a header 'apiKey' type security definition with name 'apikey_header_auth', an operation that requires the 'apikey_auth' authentication and the resource URL", t, func() {
		securityPolicyName := "apikey_header_auth"
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				securityPolicyName: apiKeyHeader{
					apiKey{
						name:  "Authorization",
						value: "superSecretKey",
					},
				},
			},
		}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: securityPolicyName}}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(nil)
		Convey("When prepareAuth method is called with a providerConfiguration", func() {
			authContext, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should contain a key 'Authorization'", func() {
				So(authContext.headers, ShouldContainKey, "Authorization")
			})
			Convey("And the value of the 'Authorization' entry should be superSecretKey", func() {
				So(authContext.headers["Authorization"], ShouldEqual, "superSecretKey")
			})
			Convey("And the url returned should be the same as the input parameter ", func() {
				So(authContext.url, ShouldEqual, url)
			})
		})
	})

	Convey("Given a provider configuration containing a query 'apiKey' type security definition with name 'apikey_query_auth', an operation that requires the 'apikey_auth' authentication and the resource URL", t, func() {
		securityPolicyName := "apikey_query_auth"
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				securityPolicyName: apiKeyQuery{
					apiKey{
						name:  "Authorization",
						value: "superSecretKey",
					},
				},
			},
		}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: securityPolicyName}}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(nil)
		Convey("When prepareAPIKeyAuthentication method is called with the operation, providerConfiguration and the service url", func() {
			authContext, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should be empty", func() {
				So(authContext.headers, ShouldBeEmpty)
			})
			Convey("And the url returned should be the same as the input parameter ", func() {
				So(authContext.url, ShouldEqual, fmt.Sprintf("%s?%s=%s", url, "Authorization", "superSecretKey"))
			})
		})
	})

	Convey("Given a provider configuration containing multiple 'apiKey' type security definitions (apikey_header_auth and apikey_query_auth), an operation that requires both 'apikey_header_auth' AND 'apikey_query_auth' authentication and the resource URL", t, func() {
		apiKeyHeaderSecurityPolicyName := "apikey_header_auth"
		apiKeyQuerySecurityPolicyName := "apikey_query_auth"
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				apiKeyHeaderSecurityPolicyName: apiKeyHeader{
					apiKey{
						name:  "Authorization",
						value: "superSecretKeyInHeader",
					},
				},
				apiKeyQuerySecurityPolicyName: apiKeyQuery{
					apiKey{
						name:  "someQueryParam",
						value: "superSecretKeyInQuery",
					},
				},
			},
		}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: apiKeyHeaderSecurityPolicyName}, SpecSecurityScheme{Name: apiKeyQuerySecurityPolicyName}}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(nil)
		Convey("When prepareAuth method is called with the providerConfiguration", func() {
			authContext, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then both security policies (apikey_header_auth) and (apikey_query_auth), should be the used for auth", func() {
				// Checking whether the apiKey query mechanism has been picked; otherwise apiKey header must be present - either or
				So(authContext.url, ShouldEqual, fmt.Sprintf("%s?someQueryParam=superSecretKeyInQuery", url))
				So(authContext.headers, ShouldContainKey, "Authorization")
				So(authContext.headers["Authorization"], ShouldEqual, "superSecretKeyInHeader")
			})
		})
	})

	Convey("Given a provider configuration containing multiple 'apiKey' type security definitions (apiKey and appId), an operation that requires both 'apiKey' AND 'appId' header authentication and the resource URL", t, func() {
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				// provider config keys are always terraform name compliant - snake case
				"api_key": apiKeyHeader{
					apiKey{
						name:  "X-API-KEY",
						value: "superSecretKeyForApiKey",
					},
				},
				"app_id": apiKeyHeader{
					apiKey{
						name:  "X-APP-ID",
						value: "superSecretKeyForAppId",
					},
				},
			},
		}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: "apiKey"}, SpecSecurityScheme{Name: "appId"}}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(nil)
		Convey("When prepareAuth method is called with the providerConfiguration", func() {
			authContext, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then both api key security policies (apiKey) and (appId), should be the used for auth", func() {
				// Checking whether the apiKey query mechanism has been picked; otherwise apiKey header must be present - either or
				So(authContext.headers["X-API-KEY"], ShouldEqual, "superSecretKeyForApiKey")
				So(authContext.headers["X-APP-ID"], ShouldEqual, "superSecretKeyForAppId")
				So(authContext.url, ShouldEqual, url)
			})
		})
	})

	Convey("Given a provider configuration containing security definitions for the global security contains policies default and an operation that DOES NOT have any specific security scheme and the resource URL", t, func() {
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				"api_key": apiKeyHeader{
					apiKey{
						name:  "X-API-KEY",
						value: "superSecretKeyForApiKey",
					},
				},
			},
		}
		globalSecuritySchemes := &SpecSecuritySchemes{SpecSecurityScheme{Name: "apiKey"}}
		operationSecuritySchemes := SpecSecuritySchemes{
			// Operation DOES NOT have security schemes
		}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(globalSecuritySchemes)
		Convey("When prepareAuth method is called with the providerConfiguration", func() {
			authContext, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the global security should be applied even though the operation DOES NOT have a security scheme", func() {
				So(authContext.headers["X-API-KEY"], ShouldEqual, "superSecretKeyForApiKey")
				So(authContext.url, ShouldEqual, url)
			})
		})
	})

	Convey("Given a provider configuration containing security definitions for both global security schemes and operation overrides and the resource URL", t, func() {
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				"api_key": apiKeyHeader{
					apiKey{
						name:  "X-API-KEY",
						value: "superSecretKeyForApiKey",
					},
				},
				"api_key_override": apiKeyHeader{
					apiKey{
						name:  "X-API-KEY_OVERRIDE",
						value: "superSecretKeyForSpecialOperationApiKey",
					},
				},
			},
		}
		globalSecuritySchemes := &SpecSecuritySchemes{SpecSecurityScheme{Name: "apiKey"}}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: "apiKeyOverride"}}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(globalSecuritySchemes)
		Convey("When prepareAuth method is called with the providerConfiguration", func() {
			authContext, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the global security should be applied even though the operation DOES NOT have a security scheme", func() {
				So(authContext.headers["X-API-KEY_OVERRIDE"], ShouldEqual, "superSecretKeyForSpecialOperationApiKey")
				So(authContext.url, ShouldEqual, url)
			})
		})
	})

	Convey("Given a global security setting containing schemes which are not defined in the provider security definitions, and an operation with NO security schemes", t, func() {
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				"apiKey": apiKeyHeader{
					apiKey{
						name:  "X-API-KEY",
						value: "superSecretKeyForApiKey",
					},
				},
			},
		}
		globalSecuritySchemes := &SpecSecuritySchemes{SpecSecurityScheme{Name: "not_defined_scheme"}}
		operationSecuritySchemes := SpecSecuritySchemes{
			// Operation DOES NOT have security schemes}
		}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(globalSecuritySchemes)
		Convey("When prepareAuth method is called with the providerConfiguration", func() {
			_, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should NOT be nil as global schemes contain policies which are not defined", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the err message should be", func() {
				So(err.Error(), ShouldEqual, "operation's security policy '{not_defined_scheme}' is not defined, please make sure the swagger file contains a security definition named '{not_defined_scheme}' under the securityDefinitions section")
			})
		})
	})

	Convey("Given an operation security setting containing schemes which are not defined in the provider security definitions", t, func() {
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				"api_key": apiKeyHeader{
					apiKey{
						name:  "X-API-KEY",
						value: "superSecretKeyForApiKey",
					},
				},
			},
		}
		operationSecuritySchemes := SpecSecuritySchemes{SpecSecurityScheme{Name: "not_defined_scheme"}}
		url := "https://www.host.com/v1/resource"
		oa := newAPIAuthenticator(nil)
		Convey("When prepareAuth method is called with the providerConfiguration", func() {
			_, err := oa.prepareAuth(url, operationSecuritySchemes, providerConfig)
			Convey("Then err should NOT be nil as global schemes contain policies which are not defined", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the err message should be", func() {
				So(err.Error(), ShouldEqual, "operation's security policy '{not_defined_scheme}' is not defined, please make sure the swagger file contains a security definition named '{not_defined_scheme}' under the securityDefinitions section")
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
			name:  "Authorization",
			value: "superSecretKey",
		}
		providerConfig := providerConfiguration{
			SecuritySchemaDefinitions: map[string]authenticator{
				securityPolicyName: apiKeyHeader{
					expectedAPIKey,
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
			SecuritySchemaDefinitions: map[string]authenticator{
				securityPolicyName: apiKeyHeader{
					apiKey{
						name:  "Authorization",
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
