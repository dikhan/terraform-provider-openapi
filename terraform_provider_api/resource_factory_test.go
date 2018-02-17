package main

import (
	"fmt"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPrepareAPIKeyAuthentication(t *testing.T) {
	Convey("Given a provider configuration containing a header 'apiKey' type security definition with name 'apikey_header_auth', an operation that requires the 'apikey_auth' authentication and the resource URL", t, func() {
		securityPolicyName := "apikey_header_auth"
		providerConfig := providerConfig{
			SecuritySchemaDefinitions: map[string]apiKeyAuthentication{
				securityPolicyName: apiKeyHeader{
					apiKey{
						name:  "Authorization",
						value: "superSecretKey",
					},
				},
			},
		}
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Security: []map[string][]string{
					{
						securityPolicyName: {},
					},
				},
			},
		}
		url := "https://www.host.com/v1/resource"
		Convey("When prepareAPIKeyAuthentication method is called with the operation, providerConfig and the service url", func() {
			r := resourceFactory{}
			headers, updatedURL := r.prepareAPIKeyAuthentication(operation, providerConfig, url)
			Convey("Then the map returned should contain a key 'Authorization'", func() {
				So(headers, ShouldContainKey, "Authorization")
			})
			Convey("And the value of the 'Authorization' entry should be superSecretKey", func() {
				So(headers["Authorization"], ShouldEqual, "superSecretKey")
			})
			Convey("And the url returned should be the same as the input parameter ", func() {
				So(url, ShouldEqual, updatedURL)
			})
		})
	})

	Convey("Given a provider configuration containing a query 'apiKey' type security definition with name 'apikey_query_auth', an operation that requires the 'apikey_auth' authentication and the resource URL", t, func() {
		securityPolicyName := "apikey_query_auth"
		providerConfig := providerConfig{
			SecuritySchemaDefinitions: map[string]apiKeyAuthentication{
				securityPolicyName: apiKeyQuery{
					apiKey{
						name:  "Authorization",
						value: "superSecretKey",
					},
				},
			},
		}
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Security: []map[string][]string{
					{
						securityPolicyName: {},
					},
				},
			},
		}
		url := "https://www.host.com/v1/resource"
		Convey("When prepareAPIKeyAuthentication method is called with the operation, providerConfig and the service url", func() {
			r := resourceFactory{}
			headers, updatedURL := r.prepareAPIKeyAuthentication(operation, providerConfig, url)
			Convey("Then the map returned should be empty", func() {
				So(headers, ShouldBeEmpty)
			})
			Convey("And the url returned should be the same as the input parameter ", func() {
				So(updatedURL, ShouldEqual, fmt.Sprintf("%s?%s=%s", url, "Authorization", "superSecretKey"))
			})
		})
	})

	Convey("Given a provider configuration containing multiple 'apiKey' type security definitions (apikey_header_auth and apikey_query_auth), an operation that requires either 'apikey_header_auth' or 'apikey_query_auth' authentication and the resource URL", t, func() {
		providerConfig := providerConfig{
			SecuritySchemaDefinitions: map[string]apiKeyAuthentication{
				"apikey_header_auth": apiKeyHeader{
					apiKey{
						name:  "Authorization",
						value: "superSecretKeyInHeader",
					},
				},
				"apikey_query_auth": apiKeyQuery{
					apiKey{
						name:  "Authorization",
						value: "superSecretKeyInQuery",
					},
				},
			},
		}
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Security: []map[string][]string{
					{
						"apikey_header_auth": {},
						"apikey_query_auth":  {},
					},
				},
			},
		}
		url := "https://www.host.com/v1/resource"
		Convey("When prepareAPIKeyAuthentication method is called with the operation, providerConfig and the service url", func() {
			r := resourceFactory{}
			headers, returnedURL := r.prepareAPIKeyAuthentication(operation, providerConfig, url)
			Convey("Then one of the authentication mechanisms should be used with no preference, the option selected would be the first one found in the map", func() {
				// Checking whether the apiKey query mechanism has been picked; otherwise apiKey header must be present - either or
				if len(headers) == 0 {
					So(returnedURL, ShouldEqual, fmt.Sprintf("%s?Authorization=superSecretKeyInQuery", url))
				} else {
					So(headers, ShouldContainKey, "Authorization")
					So(headers["Authorization"], ShouldEqual, "superSecretKeyInHeader")
					So(returnedURL, ShouldEqual, url)
				}
			})
		})
	})
}

func TestAuthRequired(t *testing.T) {
	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_header_auth' and an operation that requires the 'apikey_auth' authentication", t, func() {
		securityPolicyName := "apikey_header_auth"
		providerConfig := providerConfig{
			SecuritySchemaDefinitions: map[string]apiKeyAuthentication{
				securityPolicyName: apiKeyHeader{
					apiKey{
						name:  "Authorization",
						value: "superSecretKey"},
				},
			},
		}
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Security: []map[string][]string{
					{
						securityPolicyName: {},
					},
				},
			},
		}
		Convey("When authRequired method is called", func() {
			r := resourceFactory{}
			authRequired, securityDefinitionName := r.authRequired(operation, providerConfig)
			Convey("Then the values returned should be true and the name of the security policy 'apikey_header_auth'", func() {
				So(authRequired, ShouldBeTrue)
				So(securityDefinitionName, ShouldEqual, securityPolicyName)
			})
		})
	})

	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_auth' and an operation that DOES NOT require any authentication", t, func() {
		securityPolicyName := "apikey_auth"
		providerConfig := providerConfig{
			SecuritySchemaDefinitions: map[string]apiKeyAuthentication{
				securityPolicyName: apiKeyHeader{
					apiKey{
						name:  "Authorization",
						value: "superSecretKey",
					},
				},
			},
		}
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Security: []map[string][]string{
					{
					// No auth is required for this operation
					},
				},
			},
		}
		Convey("When authRequired method is called", func() {
			r := resourceFactory{}
			authRequired, securityDefinitionName := r.authRequired(operation, providerConfig)
			Convey("Then the values returned should be false and the name of the security policy should be empty", func() {
				So(authRequired, ShouldBeFalse)
				So(securityDefinitionName, ShouldBeEmpty)
			})
		})
	})
}
