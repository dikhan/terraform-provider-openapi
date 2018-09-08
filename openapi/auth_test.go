package openapi

//
//import (
//	"fmt"
//	"github.com/go-openapi/spec"
//	. "github.com/smartystreets/goconvey/convey"
//	"testing"
//)
//
//func TestPrepareAuth(t *testing.T) {
//	Convey("Given a provider configuration containing a header 'apiKey' type security definition with name 'apikey_header_auth', an operation that requires the 'apikey_auth' authentication and the resource URL", t, func() {
//		securityPolicyName := "apikey_header_auth"
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				securityPolicyName: apiKeyHeader{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKey",
//					},
//				},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						securityPolicyName: {},
//					},
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(nil)
//		Convey("When prepareAuth method is called with a providerConfiguration", func() {
//			authContext, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("Then the map returned should contain a key 'Authorization'", func() {
//				So(authContext.headers, ShouldContainKey, "Authorization")
//			})
//			Convey("And the value of the 'Authorization' entry should be superSecretKey", func() {
//				So(authContext.headers["Authorization"], ShouldEqual, "superSecretKey")
//			})
//			Convey("And the url returned should be the same as the input parameter ", func() {
//				So(authContext.url, ShouldEqual, url)
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing a query 'apiKey' type security definition with name 'apikey_query_auth', an operation that requires the 'apikey_auth' authentication and the resource URL", t, func() {
//		securityPolicyName := "apikey_query_auth"
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				securityPolicyName: apiKeyQuery{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKey",
//					},
//				},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						securityPolicyName: {},
//					},
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(nil)
//		Convey("When prepareAPIKeyAuthentication method is called with the operation, providerConfiguration and the service url", func() {
//			authContext, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("Then the map returned should be empty", func() {
//				So(authContext.headers, ShouldBeEmpty)
//			})
//			Convey("And the url returned should be the same as the input parameter ", func() {
//				So(authContext.url, ShouldEqual, fmt.Sprintf("%s?%s=%s", url, "Authorization", "superSecretKey"))
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing multiple 'apiKey' type security definitions (apikey_header_auth and apikey_query_auth), an operation that requires either 'apikey_header_auth' or 'apikey_query_auth' authentication and the resource URL", t, func() {
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				"apikey_header_auth": apiKeyHeader{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKeyInHeader",
//					},
//				},
//				"apikey_query_auth": apiKeyQuery{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKeyInQuery",
//					},
//				},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						"apikey_header_auth": {},
//					},
//					{
//						"apikey_query_auth": {},
//					},
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(nil)
//		Convey("When prepareAuth method is called with the providerConfiguration", func() {
//			authContext, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("Then the security policy present in the first position of the security policies (apikey_header_auth), should be the one used as default mechanism for auth", func() {
//				So(authContext.headers, ShouldContainKey, "Authorization")
//				So(authContext.headers["Authorization"], ShouldEqual, "superSecretKeyInHeader")
//				So(authContext.url, ShouldEqual, url)
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing multiple 'apiKey' type security definitions (apikey_header_auth and apikey_query_auth), an operation that requires both 'apikey_header_auth' AND 'apikey_query_auth' authentication and the resource URL", t, func() {
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				"apikey_header_auth": apiKeyHeader{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKeyInHeader",
//					},
//				},
//				"apikey_query_auth": apiKeyQuery{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKeyInQuery",
//					},
//				},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						"apikey_header_auth": {},
//						"apikey_query_auth":  {},
//					},
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(nil)
//		Convey("When prepareAuth method is called with the providerConfiguration", func() {
//			authContext, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("Then both security policies (apikey_header_auth) and (apikey_query_auth), should be the used for auth", func() {
//				// Checking whether the apiKey query mechanism has been picked; otherwise apiKey header must be present - either or
//				So(authContext.url, ShouldEqual, fmt.Sprintf("%s?Authorization=superSecretKeyInQuery", url))
//				So(authContext.headers, ShouldContainKey, "Authorization")
//				So(authContext.headers["Authorization"], ShouldEqual, "superSecretKeyInHeader")
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing multiple 'apiKey' type security definitions (apiKey and appId), an operation that requires both 'apiKey' AND 'appId' authentication and the resource URL", t, func() {
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				"apiKey": apiKeyHeader{
//					apiKey{
//						name:  "X-API-KEY",
//						value: "superSecretKeyForApiKey",
//					},
//				},
//				"appId": apiKeyHeader{
//					apiKey{
//						name:  "X-APP-ID",
//						value: "superSecretKeyForAppId",
//					},
//				},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						"apiKey": {},
//						"appId":  {},
//					},
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(nil)
//		Convey("When prepareAuth method is called with the providerConfiguration", func() {
//			authContext, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("Then both api key security policies (apiKey) and (appId), should be the used for auth", func() {
//				// Checking whether the apiKey query mechanism has been picked; otherwise apiKey header must be present - either or
//				So(authContext.headers["X-API-KEY"], ShouldEqual, "superSecretKeyForApiKey")
//				So(authContext.headers["X-APP-ID"], ShouldEqual, "superSecretKeyForAppId")
//				So(authContext.url, ShouldEqual, url)
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing security definitions for the global security contains policies default and an operation that DOES NOT have any specific security scheme and the resource URL", t, func() {
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				"apiKey": apiKeyHeader{
//					apiKey{
//						name:  "X-API-KEY",
//						value: "superSecretKeyForApiKey",
//					},
//				},
//			},
//		}
//		globalSecuritySchemes := []map[string][]string{
//			{
//				"apiKey": {},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					// Operation DOES NOT have security schemes
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(globalSecuritySchemes)
//		Convey("When prepareAuth method is called with the providerConfiguration", func() {
//			authContext, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("Then the global security should be applied even though the operation DOES NOT have a security scheme", func() {
//				So(authContext.headers["X-API-KEY"], ShouldEqual, "superSecretKeyForApiKey")
//				So(authContext.url, ShouldEqual, url)
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing security definitions for both global security schemes and operation overrides and the resource URL", t, func() {
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				"apiKey": apiKeyHeader{
//					apiKey{
//						name:  "X-API-KEY",
//						value: "superSecretKeyForApiKey",
//					},
//				},
//				"apiKeyOverride": apiKeyHeader{
//					apiKey{
//						name:  "X-API-KEY_OVERRIDE",
//						value: "superSecretKeyForSpecialOperationApiKey",
//					},
//				},
//			},
//		}
//		globalSecuritySchemes := []map[string][]string{
//			{
//				"apiKey": {},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						"apiKeyOverride": {},
//					},
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(globalSecuritySchemes)
//		Convey("When prepareAuth method is called with the providerConfiguration", func() {
//			authContext, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//			Convey("Then the global security should be applied even though the operation DOES NOT have a security scheme", func() {
//				So(authContext.headers["X-API-KEY_OVERRIDE"], ShouldEqual, "superSecretKeyForSpecialOperationApiKey")
//				So(authContext.url, ShouldEqual, url)
//			})
//		})
//	})
//
//	Convey("Given a global security setting containing schemes which are not defined in the provider security definitions, and an operation with NO security schemes; therefore global security is applied", t, func() {
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				"apiKey": apiKeyHeader{
//					apiKey{
//						name:  "X-API-KEY",
//						value: "superSecretKeyForApiKey",
//					},
//				},
//			},
//		}
//		globalSecuritySchemes := []map[string][]string{
//			{
//				"not_defined_scheme": {},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					// Operation DOES NOT have security schemes
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := newAPIAuthenticator(globalSecuritySchemes)
//		Convey("When prepareAuth method is called with the providerConfiguration", func() {
//			_, err := oa.prepareAuth("CreateResourceV1", url, operation.Security, providerConfig)
//			Convey("Then err should NOT be nil as global schemes contain policies which are not defined", func() {
//				So(err, ShouldNotBeNil)
//			})
//		})
//	})
//}
//
//func TestAuthRequired(t *testing.T) {
//	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_header_auth' and an operation that requires the 'apikey_auth' authentication", t, func() {
//		securityPolicyName := "apikey_header_auth"
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						securityPolicyName: {},
//					},
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := apiAuth{}
//		Convey("When authRequired method is called", func() {
//			authRequired, operationSecurityPolicies := oa.authRequired("CreateResourceV1", url, operation.Security)
//			Convey("Then the values returned should be true and the name of the security policy 'apikey_header_auth'", func() {
//				So(authRequired, ShouldBeTrue)
//				So(operationSecurityPolicies, ShouldContainKey, securityPolicyName)
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_auth' and an operation that DOES NOT require any authentication", t, func() {
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					// No auth is required for this operation}
//				},
//			},
//		}
//		url := "https://www.host.com/v1/resource"
//		oa := apiAuth{}
//		Convey("When authRequired method is called", func() {
//			authRequired, operationSecurityPolicies := oa.authRequired("CreateResourceV1", url, operation.Security)
//			Convey("Then the values returned should be false and the name of the security policy should be empty", func() {
//				So(authRequired, ShouldBeFalse)
//				So(operationSecurityPolicies, ShouldBeEmpty)
//			})
//		})
//	})
//}
//
//func TestConfirmOperationSecurityPoliciesAreDefined(t *testing.T) {
//	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_auth' and an operation that requires api key header authentication", t, func() {
//		securityPolicyName := "apikey_auth"
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				securityPolicyName: apiKeyHeader{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKey",
//					},
//				},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						securityPolicyName: {},
//					},
//				},
//			},
//		}
//		oa := apiAuth{}
//		Convey("When confirmOperationSecurityPoliciesAreDefined method with a security policy which is also defined in the security definitions", func() {
//			err := oa.confirmOperationSecurityPoliciesAreDefined(operation.Security[0], providerConfig)
//			Convey("Then the err returned should be nil", func() {
//				So(err, ShouldBeNil)
//			})
//		})
//	})
//
//	Convey("Given a provider configuration containing an 'apiKey' type security definition with name 'apikey_auth' and an operation that requires api key header authentication", t, func() {
//		securityPolicyName := "apikey_auth"
//		providerConfig := providerConfiguration{
//			SecuritySchemaDefinitions: map[string]authenticator{
//				securityPolicyName: apiKeyHeader{
//					apiKey{
//						name:  "Authorization",
//						value: "superSecretKey",
//					},
//				},
//			},
//		}
//		operation := &spec.Operation{
//			OperationProps: spec.OperationProps{
//				Security: []map[string][]string{
//					{
//						"non_defined_security_policy": {},
//					},
//				},
//			},
//		}
//		oa := apiAuth{}
//		Convey("When confirmOperationSecurityPoliciesAreDefined method with a security policy which is NOT defined in the security definitions", func() {
//			err := oa.confirmOperationSecurityPoliciesAreDefined(operation.Security[0], providerConfig)
//			Convey("Then the err returned should not be nil", func() {
//				So(err, ShouldNotBeNil)
//			})
//		})
//	})
//}
