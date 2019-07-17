package openapi

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-openapi/spec"

	"github.com/dikhan/http_goclient"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProviderClient(t *testing.T) {
	Convey("Given a SpecBackendConfiguration, HttpClient, providerConfiguration and specAuthenticator", t, func() {
		var openAPIBackendConfiguration SpecBackendConfiguration
		providerConfiguration := providerConfiguration{}
		var apiAuthenticator specAuthenticator
		Convey("When ProviderClient method is constructed", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: openAPIBackendConfiguration,
				httpClient:                  &http_goclient.HttpClientStub{},
				providerConfiguration:       providerConfiguration,
				apiAuthenticator:            apiAuthenticator,
			}
			Convey("Then the providerClient should comply with ClientOpenAPI interface", func() {
				var _ ClientOpenAPI = providerClient
			})
		})
	})
}

func TestAppendOperationHeaders(t *testing.T) {
	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
		operationHeader := "operationHeader"
		operationHeaderTfName := "operation_header_tf_name"
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{},
			httpClient:                  &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration{
				Headers: map[string]string{
					operationHeaderTfName: "operationHeaderValue",
				},
			},
			apiAuthenticator: &specStubAuthenticator{},
		}
		Convey("When appendOperationHeaders with an operation headers, the provider config containing the values of the headers and a map that should contain the final result", func() {
			resourcePostOperation := &specResourceOperation{
				HeaderParameters: SpecHeaderParameters{
					{
						Name:          operationHeader,
						TerraformName: operationHeaderTfName,
					},
				},
				responses:       specResponses{},
				SecuritySchemes: SpecSecuritySchemes{},
			}
			headersMap := map[string]string{
				"someHeaderAlreadyPresent": "someValue",
			}
			providerClient.appendOperationHeaders(resourcePostOperation.HeaderParameters, providerClient.providerConfiguration, headersMap)
			Convey("And the headersMap should contain whatever headers where already in the map", func() {
				So(headersMap, ShouldContainKey, "someHeaderAlreadyPresent")
				So(headersMap["someHeaderAlreadyPresent"], ShouldEqual, "someValue")
			})
			Convey("And the headersMap should contain the new ones added from the operation headers", func() {
				So(headersMap, ShouldContainKey, operationHeader)
				So(headersMap[operationHeader], ShouldEqual, "operationHeaderValue")
			})
		})
	})
}

func TestAppendUserAgentHeader(t *testing.T) {
	Convey("Given a providerClient and user agent header value", t, func() {
		providerClient := &ProviderClient{}
		expectedHeaderValue := "some user agent header value"
		Convey("When appendUserAgentHeader with empty header map and header value", func() {
			headers := map[string]string{}
			providerClient.appendUserAgentHeader(headers, expectedHeaderValue)
			Convey("Then the user agent header value should exist in the header map with correct value", func() {
				value, exists := headers[userAgent]
				So(exists, ShouldBeTrue)
				So(value, ShouldEqual, expectedHeaderValue)
			})
		})
		Convey("When appendUserAgentHeader with non-empty header map and header value", func() {
			headers := map[string]string{"Some-Header": "some header value"}
			providerClient.appendUserAgentHeader(headers, expectedHeaderValue)
			Convey("Then the user agent header should exist in the header map with correct value", func() {
				value, exists := headers[userAgent]
				So(exists, ShouldBeTrue)
				So(value, ShouldEqual, expectedHeaderValue)
				So(headers["Some-Header"], ShouldEqual, "some header value")
			})
		})
		Convey("When appendUserAgentHeader with header map containing User-Agent and new header value", func() {
			headers := map[string]string{userAgent: "some existing user agent header value"}
			providerClient.appendUserAgentHeader(headers, expectedHeaderValue)
			Convey("Then the user agent header should exist in the header map with correct value", func() {
				value, exists := headers[userAgent]
				So(exists, ShouldBeTrue)
				So(value, ShouldEqual, expectedHeaderValue)
			})
		})
	})
}

func TestGetResourceIDURL(t *testing.T) {
	Convey("Given a providerClient", t, func() {
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration{},
		}
		Convey("When getResourceIDURL is called with a specResource and ID", func() {
			expectedID := "1234"
			expectedPath := "/v1/resource"
			r := &SpecV2Resource{
				Path: expectedPath,
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			}
			resourceURL, err := providerClient.getResourceIDURL(r, []string{}, expectedID)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then the resourceURL returned should be built from the schemes, host, base path, and path in the client and the ID passed", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s/%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath, expectedID))
			})
		})

		Convey("When getResourceIDURL is called with a specResource containing trailing / in the path and an ID", func() {
			expectedID := "1234"
			expectedPath := "/v1/resource/"
			r := &SpecV2Resource{
				Path: expectedPath,
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			}
			resourceURL, err := providerClient.getResourceIDURL(r, []string{}, expectedID)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath, expectedID))
			})
		})

		Convey("When getResourceIDURL is called with a specResource containing a parameterised path and a parent ID and instance ID", func() {
			expectedID := "5678"
			parentIDs := []string{"1234"}
			r := &SpecV2Resource{
				Path: "/v1/resource/{resource_id}/subresource",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			}
			resourceURL, err := providerClient.getResourceIDURL(r, parentIDs, expectedID)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s/v1/resource/%s/subresource/%s", expectedProtocol, expectedHost, expectedBasePath, parentIDs[0], expectedID))
			})
		})

		Convey("When getResourceIDURL is called with a specResource containing a parameterized path and instance ID but missing a parent ID", func() {
			r := &SpecV2Resource{
				Path: "/v1/resource/{resource_id}/subresource",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			}
			resourceURL, err := providerClient.getResourceIDURL(r, []string{}, "5678")
			Convey("Then an error should be returned", func() {
				//TODO: possible bug-- should the message contain all this noise: ([[/{resource_id}/ {resource_id}]]) ?
				So(err.Error(), ShouldEqual, "could not resolve sub-resource path correctly '/v1/resource/{resource_id}/subresource' ([[/{resource_id}/ {resource_id}]]) with the given ids - missing ids to resolve the path params properly: []")
			})
			Convey("And then resourceURL should be empty", func() {
				So(resourceURL, ShouldBeEmpty)
			})
		})

		Convey("When getResourceIDURL is called with an empty ID", func() {
			r := &SpecV2Resource{
				Path: "whatever",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			}
			_, err := providerClient.getResourceIDURL(r, []string{}, "")
			Convey("Then the error returned should match the expected one", func() {
				So(err.Error(), ShouldEqual, "could not build the resourceIDURL: required instance id value is missing")
			})
		})
	})
}

func TestGetResourceURL_table_driven(t *testing.T) {
	providerClient := &ProviderClient{
		openAPIBackendConfiguration: &specStubBackendConfiguration{
			host:        "wwww.host.com",
			basePath:    "/api",
			httpSchemes: []string{"http"},
		},
	}

	r := &SpecV2Resource{
		RootPathItem: spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		},
	}

	t.Run("no parent ids", func(t *testing.T) {
		t.Run("happy paths", func(t *testing.T) {
			testcases := []struct {
				name                string
				path                string
				id                  string
				parentIDs           []string
				expectedResourceURL string
			}{
				{name: "no trailing slash", path: "/v1/resource", id: "1234", expectedResourceURL: "http://wwww.host.com/api/v1/resource/1234"},
				{name: "different id", path: "/v1/resource", id: "42", expectedResourceURL: "http://wwww.host.com/api/v1/resource/42"},
				{name: "with a parent id", path: "/v1/resource/{parent_id}/v17/subresource", id: "42", parentIDs: []string{"3.14159"}, expectedResourceURL: "http://wwww.host.com/api/v1/resource/3.14159/v17/subresource/42"},
				{name: "with a parent id with mustaches", path: "/v1/resource/{parent_id}/v17/subresource", id: "42", parentIDs: []string{"{3.14159}"}, expectedResourceURL: "http://wwww.host.com/api/v1/resource/{3.14159}/v17/subresource/42"},        //TODO: possible bug?
				{name: "with a parent id with a slash", path: "/v1/resource/{parent_id}/v17/subresource", id: "42", parentIDs: []string{"3.14/159"}, expectedResourceURL: "http://wwww.host.com/api/v1/resource/3.14/159/v17/subresource/42"},            //TODO: possible bug?
				{name: "with a token with double mustaches", path: "/v1/resource/{{parent_id}}/v17/subresource", id: "42", parentIDs: []string{"3.14159"}, expectedResourceURL: "http://wwww.host.com/api/v1/resource/{{parent_id}}/v17/subresource/42"}, //TODO: possible bug?
				{name: "with a parent id but no tokens", path: "/v1/resource", id: "42", parentIDs: []string{"pi"}, expectedResourceURL: "http://wwww.host.com/api/v1/resource/42"},
				{name: "trailing slash", path: "/v1/resource/", id: "1337", expectedResourceURL: "http://wwww.host.com/api/v1/resource/1337"},
				{name: "id with a slash", path: "/v1/resource/", id: "13/37", expectedResourceURL: "http://wwww.host.com/api/v1/resource/13/37"},     //TODO: possible bug?
				{name: "id with mustaches", path: "/v1/resource/", id: "1{33}7", expectedResourceURL: "http://wwww.host.com/api/v1/resource/1{33}7"}, //TODO: possible bug?
			}
			for _, tc := range testcases {
				r.Path = tc.path
				actualResourceURL, err := providerClient.getResourceIDURL(r, tc.parentIDs, tc.id)
				assert.NoError(t, err, tc.name)
				assert.Equal(t, tc.expectedResourceURL, actualResourceURL, tc.name)
			}
		})

		t.Run("crappy paths", func(t *testing.T) {
			testcases := []struct {
				name          string
				path          string
				id            string
				parentIDs     []string
				expectedError string
			}{
				{name: "empty id", path: "/v1/resource/", id: "", expectedError: "could not build the resourceIDURL: required instance id value is missing"},
				{name: "double trailing slash", path: "/v1/resource//", id: "1337", expectedError: "could not resolve sub-resource path correctly '/v1/resource//' ([[// ]]) with the given ids - missing ids to resolve the path params properly: []"},                                                                                                   //TODO: possible bug? remove ([[// ]]) ?
				{name: "double leading slash", path: "//v1/resource/", id: "1337", expectedError: "could not resolve sub-resource path correctly '//v1/resource/' ([[// ]]) with the given ids - missing ids to resolve the path params properly: []"},                                                                                                    //TODO: possible bug? remove ([[// ]]) ?
				{name: "double slash in the middle", path: "/v1//resource/", id: "1337", expectedError: "could not resolve sub-resource path correctly '/v1//resource/' ([[// ]]) with the given ids - missing ids to resolve the path params properly: []"},                                                                                              //TODO: possible bug? remove ([[// ]]) ?
				{name: "with a missing parent id", path: "/v1/resource/{parent_id}/v17/subresource", id: "42", parentIDs: []string{}, expectedError: "could not resolve sub-resource path correctly '/v1/resource/{parent_id}/v17/subresource' ([[/{parent_id}/ {parent_id}]]) with the given ids - missing ids to resolve the path params properly: []"}, //TODO: possible bug? remove ([[/{parent_id}/ {parent_id}]]) and [] ?
				{name: "with extra parent ids", path: "/v1/resource/{parent_id}/v17/subresource", id: "42", parentIDs: []string{"-1", "-2"}, expectedError: "could not resolve sub-resource path correctly '/v1/resource/{parent_id}/v17/subresource' ([[/{parent_id}/ {parent_id}]]) with the given ids - more ids than path params: [-1 -2]"},           //TODO: possible bug? remove ([[/{parent_id}/ {parent_id}]]) and [-1, -2] ?
			}
			for _, tc := range testcases {
				r.Path = tc.path
				actualResourceURL, err := providerClient.getResourceIDURL(r, tc.parentIDs, tc.id)
				assert.EqualError(t, err, tc.expectedError, tc.name)
				assert.Empty(t, actualResourceURL, tc.name)
			}
		})
	})

}

func TestGetResourceURL(t *testing.T) {
	Convey("Given a providerClient set up with stub auth that injects some headers to the request and is not multiregion", t, func() {
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration{},
			apiAuthenticator: &specStubAuthenticator{
				authContext: &authContext{
					url: "",
					headers: map[string]string{
						"Authentication": "Bearer secret!",
					},
				},
			},
		}
		Convey("When getResourceURL is called with a specResource with a resource path that is not parameterised", func() {
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL is called with a resource which blows up on getResourcePath", func() {
			specStubResource := &specStubResource{
				funcGetResourcePath: func(parentIDs []string) (string, error) { return "", errors.New("getResourcePath blew up") },
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should not be nil", func() {
				So(err.Error(), ShouldNotBeNil)
			})
			Convey("And then resourceURL should be empty", func() {
				So(resourceURL, ShouldBeEmpty)
			})
		})

		// Using SpecV2Resource in this specific case to validate this specific scenario. The stub does not have logic
		// to resolve parameters and it not a good idea to update the mock to have prod logic. Hence, using a real impl SpecV2Resource
		// in this case so we have the subresource use case covered too.
		Convey("When getResourceURL is called with a specResource with a resource path that is parameterised (e,g: subresource)", func() {
			expectedParentID := "parentID"
			specStubResource := &SpecV2Resource{
				Path: "/v1/resource/{resource_id}/subresource",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{expectedParentID})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() { //TODO: what should it equal?
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s/v1/resource/%s/subresource", expectedProtocol, expectedHost, expectedBasePath, expectedParentID))
			})
		})

		Convey("When getResourceURL with a specResource with a resource path that is not parameterised and overrides the global host", func() {
			expectedHost := "wwww.host-overriden.com"
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				host: expectedHost,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL is called but the backend config has empty value for host or the resource spec has empty value for path", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{
					host:        "",
					basePath:    "/api",
					httpSchemes: []string{"http"},
				},
				httpClient:            &http_goclient.HttpClientStub{},
				providerConfiguration: providerConfiguration{},
				apiAuthenticator:      &specStubAuthenticator{},
			}

			specStubResource := &specStubResource{}
			_, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message returned should be", func() {
				So(err.Error(), ShouldEqual, "host and path are mandatory attributes to get the resource URL - host[''], path['']")
			})
		})

		Convey("When getResourceURL is called but the backend config has no httpSchemes configured", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{
					host:        "wwww.host.com",
					basePath:    "/api",
					httpSchemes: []string{""},
				},
				httpClient:            &http_goclient.HttpClientStub{},
				providerConfiguration: providerConfiguration{},
				apiAuthenticator: &specStubAuthenticator{
					authContext: &authContext{},
				},
			}
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should default to http scheme", func() {
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", "http", expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL is called but the backend config has both http and https configured", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{
					host:        "wwww.host.com",
					basePath:    "/api",
					httpSchemes: []string{"http", "https"},
				},
				httpClient:            &http_goclient.HttpClientStub{},
				providerConfiguration: providerConfiguration{},
				apiAuthenticator: &specStubAuthenticator{
					authContext: &authContext{},
				},
			}
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should default to https scheme", func() {
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", "https", expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL with a specResource with a resource path that does not have leading /", func() {
			expectedPath := "v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s/%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL with a specResource with a resource path that does not have leading basePath is not empty AND basePath is not /", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{
					host:        "wwww.host.com",
					basePath:    "api", // basePath is not empty AND basePath is not /
					httpSchemes: []string{"http"},
				},
				httpClient:            &http_goclient.HttpClientStub{},
				providerConfiguration: providerConfiguration{},
				apiAuthenticator: &specStubAuthenticator{
					authContext: &authContext{},
				},
			}
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s/%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL with a specResource with a resource path that does not have leading basePath is not empty AND basePath is not does not start with /", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{
					host:        "wwww.host.com",
					basePath:    "api/otherpath", // basePath is not empty AND basePath is not /
					httpSchemes: []string{"http"},
				},
				httpClient:            &http_goclient.HttpClientStub{},
				providerConfiguration: providerConfiguration{},
				apiAuthenticator: &specStubAuthenticator{
					authContext: &authContext{},
				},
			}
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s/%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL with a specResource with a resource path that does not have leading basePath is not empty AND basePath is not does start with /", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{
					host:        "wwww.host.com",
					basePath:    "/api/otherpath", // basePath is not empty AND basePath is not /
					httpSchemes: []string{"http"},
				},
				httpClient:            &http_goclient.HttpClientStub{},
				providerConfiguration: providerConfiguration{},
				apiAuthenticator: &specStubAuthenticator{
					authContext: &authContext{},
				},
			}
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
		})

		Convey("When getResourceURL with a specResource with a resource path that does not have leading basePath is not empty AND basePath is /", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{
					host:        "wwww.host.com",
					basePath:    "/", // basePath is /
					httpSchemes: []string{"http"},
				},
				httpClient:            &http_goclient.HttpClientStub{},
				providerConfiguration: providerConfiguration{},
				apiAuthenticator: &specStubAuthenticator{
					authContext: &authContext{},
				},
			}
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s", expectedProtocol, expectedHost, expectedPath))
			})
		})

	})

	Convey("Given a providerClient set up with a backend configuration that is multi-region and the region field being filled in (pretending user provided us-west1 in the provider's region property)", t, func() {
		expectedRegion := "us-west1"
		providerConfiguration := providerConfiguration{
			Region: expectedRegion,
		}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.%s.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
				regions:     []string{expectedRegion, "someOtherRegion"},
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      &specStubAuthenticator{},
		}
		Convey("When getResourceURL with a specResource with a resource path", func() {
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedProtocol, fmt.Sprintf(expectedHost, expectedRegion), expectedBasePath, expectedPath))
			})
		})
	})

	Convey("Given a providerClient set up with a backend configuration that is multi-region and the region field being the default (pretending user did not provide value for provider's region property)", t, func() {
		expectedRegion := "us-east1"
		providerConfiguration := providerConfiguration{
			Region: "", //emptyRegionProvidedByUser
		}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.%s.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
				regions:     []string{expectedRegion},
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      &specStubAuthenticator{},
		}
		Convey("When getResourceURL with a specResource with a resource path", func() {
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedProtocol, fmt.Sprintf(expectedHost, expectedRegion), expectedBasePath, expectedPath))
			})
		})
	})

	Convey("Given a providerClient set up with a backend configuration that is multi-region but the openAPIBackendConfiguration isMultiRegion() call throws an error", t, func() {
		expectedError := "someError"
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.%s.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
				regions:     []string{""},
				err:         fmt.Errorf(expectedError),
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration{},
			apiAuthenticator:      &specStubAuthenticator{},
		}
		Convey("When getResourceURL with a specResource with a resource path", func() {
			specStubResource := &specStubResource{}
			_, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should match the expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
			})
		})
	})

	Convey("Given a providerClient set up with a backend configuration that is multi-region but the openAPIBackendConfiguration getDefaultRegion() call throws an error", t, func() {
		expectedError := "some error thrown by default region method"
		providerConfiguration := providerConfiguration{}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:             "wwww.%s.host.com",
				basePath:         "/api",
				httpSchemes:      []string{"http"},
				regions:          []string{"us-east1"},
				defaultRegionErr: fmt.Errorf(expectedError),
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      &specStubAuthenticator{},
		}
		Convey("When getResourceURL with a specResource with a resource path", func() {
			specStubResource := &specStubResource{}
			_, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should match the expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
			})
		})
	})

	Convey("Given a providerClient set up with a backend configuration that is multi-region but the openAPIBackendConfiguration getHostByRegion(region) call throws an error", t, func() {
		expectedError := "some error thrown by default host by region method"
		providerConfiguration := providerConfiguration{}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:            "wwww.%s.host.com",
				basePath:        "/api",
				httpSchemes:     []string{"http"},
				regions:         []string{"us-east1"},
				hostByRegionErr: fmt.Errorf(expectedError),
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      &specStubAuthenticator{},
		}
		Convey("When getResourceURL with a specResource with a resource path", func() {
			specStubResource := &specStubResource{}
			_, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should match the expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
			})
		})
	})

	Convey("Given a providerClient set up with a backend configuration but the openAPIBackendConfiguration getHost() call throws an error", t, func() {
		expectedError := "some error thrown by default host method"
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.%s.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
				regions:     []string{},
				hostErr:     fmt.Errorf(expectedError),
			},
			httpClient:            &http_goclient.HttpClientStub{},
			providerConfiguration: providerConfiguration{},
			apiAuthenticator:      &specStubAuthenticator{},
		}
		Convey("When getResourceURL with a specResource with a resource path", func() {
			specStubResource := &specStubResource{}
			_, err := providerClient.getResourceURL(specStubResource, []string{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should match the expected", func() {
				So(err.Error(), ShouldEqual, expectedError)
			})
		})
	})
}

func TestPerformRequest(t *testing.T) {
	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		headerParameter := SpecHeaderParam{"Operation-Specific-Header", "operation_specific_header"}
		providerConfiguration := providerConfiguration{
			Headers: map[string]string{headerParameter.TerraformName: "some-value"},
		}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := &specStubAuthenticator{
			authContext: &authContext{
				url: "",
				headers: map[string]string{
					expectedHeader: expectedHeaderValue,
				},
			},
		}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      apiAuthenticator,
		}
		Convey("When performRequest POST method is called with a resourceURL, a requestPayload, an empty responsePayload, and header parameters", func() {
			resourcePostOperation := &specResourceOperation{
				HeaderParameters: SpecHeaderParameters{headerParameter},
				responses:        specResponses{},
				SecuritySchemes:  SpecSecuritySchemes{},
			}
			expectedReqPayloadProperty1 := "property1"
			expectedReqPayloadProperty1Value := "someValue"
			requestPayload := map[string]interface{}{
				expectedReqPayloadProperty1: expectedReqPayloadProperty1Value,
			}
			responsePayload := map[string]interface{}{}

			expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
			expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
			expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
			expectedPath := "/v1/resource"
			resourceURL := fmt.Sprintf("%s://%s%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath)

			_, err := providerClient.performRequest("POST", resourceURL, resourcePostOperation, requestPayload, responsePayload)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
			Convey("And then client should have received the right Authentication header and expected value", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
			Convey("And then client should have received the right operation header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, headerParameter.Name)
				So(httpClient.Headers[headerParameter.Name], ShouldEqual, providerConfiguration.Headers[headerParameter.TerraformName])
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
			Convey("And then client should have received the right request payload", func() {
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty1)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty1], ShouldEqual, expectedReqPayloadProperty1Value)
			})
		})
		Convey("When performRequest with a method that is not supported", func() {
			resourcePostOperation := &specResourceOperation{
				HeaderParameters: SpecHeaderParameters{},
				responses:        specResponses{},
				SecuritySchemes:  SpecSecuritySchemes{},
			}
			_, err := providerClient.performRequest("NotSupportedMethod", "", resourcePostOperation, nil, nil)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message returned should be", func() {
				So(err.Error(), ShouldEqual, "method 'NotSupportedMethod' not supported")
			})
		})
		Convey("When performRequest prepareAuth returns an error", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: &specStubBackendConfiguration{},
				apiAuthenticator: &specStubAuthenticator{
					authContext: &authContext{},
					err:         fmt.Errorf("some error with prep auth"),
				},
			}
			_, err := providerClient.performRequest("POST", "", &specResourceOperation{}, nil, nil)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message returned should be", func() {
				So(err.Error(), ShouldEqual, "some error with prep auth")
			})
		})
	})
}

func TestProviderClientPost(t *testing.T) {

	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		headerParameter := SpecHeaderParam{"Operation-Specific-Header", "operation_specific_header"}
		providerConfiguration := providerConfiguration{
			Headers: map[string]string{headerParameter.TerraformName: "some-value"},
		}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := &specStubAuthenticator{
			authContext: &authContext{
				url: "",
				headers: map[string]string{
					expectedHeader: expectedHeaderValue,
				},
			},
		}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      apiAuthenticator,
		}
		Convey("When providerClient POST method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{headerParameter},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			expectedReqPayloadProperty1 := "property1"
			expectedReqPayloadProperty1Value := "someValue"
			expectedReqPayloadProperty2 := "property2"
			expectedReqPayloadProperty2Value := 2
			requestPayload := map[string]interface{}{
				expectedReqPayloadProperty1: expectedReqPayloadProperty1Value,
				expectedReqPayloadProperty2: expectedReqPayloadProperty2Value,
			}
			responsePayload := map[string]interface{}{}

			_, err := providerClient.Post(specStubResource, requestPayload, responsePayload)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				expectedPath := specStubResource.path
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath))
			})
			Convey("And then client should have received the right Authentication header and expected value", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
			Convey("And then client should have received the right operation header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, headerParameter.Name)
				So(httpClient.Headers[headerParameter.Name], ShouldEqual, providerConfiguration.Headers[headerParameter.TerraformName])
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
			Convey("And then client should have received the right request payload", func() {
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty1)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty1], ShouldEqual, expectedReqPayloadProperty1Value)
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty2)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty2], ShouldEqual, expectedReqPayloadProperty2Value)
			})
		})

	})

	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration{},
			apiAuthenticator: &specStubAuthenticator{
				authContext: &authContext{
					headers: map[string]string{},
				},
			},
		}
		Convey("When providerClient POST method is called with a SpecV2Resource that has a subresource path, a requestPayload, an empty responsePayload and the resource parentID", func() {
			specv2Resource := &SpecV2Resource{
				Path: "/v1/resource/{id}/subresource",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							OperationProps: spec.OperationProps{
								Responses: &spec.Responses{},
							},
						},
					},
				},
			}
			expectedReqPayloadProperty1 := "property1"
			expectedReqPayloadProperty1Value := "someValue"
			expectedReqPayloadProperty2 := "property2"
			expectedReqPayloadProperty2Value := 2
			requestPayload := map[string]interface{}{
				expectedReqPayloadProperty1: expectedReqPayloadProperty1Value,
				expectedReqPayloadProperty2: expectedReqPayloadProperty2Value,
			}
			responsePayload := map[string]interface{}{}

			_, err := providerClient.Post(specv2Resource, requestPayload, responsePayload, "parentID")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s/v1/resource/parentID/subresource", expectedProtocol, expectedHost, expectedBasePath))
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
			Convey("And then client should have received the right request payload", func() {
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty1)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty1], ShouldEqual, expectedReqPayloadProperty1Value)
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty2)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty2], ShouldEqual, expectedReqPayloadProperty2Value)
			})
		})

	})
}

func TestProviderClientPut(t *testing.T) {

	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		headerParameter := SpecHeaderParam{"Operation-Specific-Header", "operation_specific_header"}
		providerConfiguration := providerConfiguration{
			Headers: map[string]string{headerParameter.TerraformName: "some-value"},
		}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := newStubAuthenticator(expectedHeader, expectedHeaderValue, nil)
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: newStubBackendConfiguration("wwww.host.com", "/api", []string{"http"}),
			httpClient:                  httpClient,
			providerConfiguration:       providerConfiguration,
			apiAuthenticator:            apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourcePutOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{headerParameter},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			expectedReqPayloadProperty1 := "property1"
			expectedReqPayloadProperty1Value := "someValue"
			requestPayload := map[string]interface{}{
				expectedReqPayloadProperty1: expectedReqPayloadProperty1Value,
			}
			responsePayload := map[string]interface{}{}
			expectedID := "1234"
			_, err := providerClient.Put(specStubResource, expectedID, requestPayload, responsePayload)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				expectedPath := specStubResource.path
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s%s/%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath, expectedID))
			})
			Convey("And then client should have received the right Authentication header and expected value", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
			Convey("And then client should have received the right operation header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, headerParameter.Name)
				So(httpClient.Headers[headerParameter.Name], ShouldEqual, providerConfiguration.Headers[headerParameter.TerraformName])
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
			Convey("And then client should have received the right request payload", func() {
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty1)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty1], ShouldEqual, expectedReqPayloadProperty1Value)
			})
		})
	})

	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration{},
			apiAuthenticator: &specStubAuthenticator{
				authContext: &authContext{
					headers: map[string]string{},
				},
			},
		}
		Convey("When providerClient PUT  method is called with a SpecV2Resource that has a subresource path, a requestPayload, an empty responsePayload and the resource parentID", func() {
			specv2Resource := &SpecV2Resource{
				Path: "/v1/resource/{id}/subresource",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							OperationProps: spec.OperationProps{
								Responses: &spec.Responses{},
							},
						},
					},
				},
				InstancePathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Put: &spec.Operation{
							OperationProps: spec.OperationProps{
								Responses: &spec.Responses{},
							},
						},
					},
				},
			}
			expectedReqPayloadProperty1 := "property1"
			expectedReqPayloadProperty1Value := "someValue"
			requestPayload := map[string]interface{}{
				expectedReqPayloadProperty1: expectedReqPayloadProperty1Value,
			}
			responsePayload := map[string]interface{}{}
			expectedID := "1234"
			parentIDs := []string{"parentID"}
			_, err := providerClient.Put(specv2Resource, expectedID, requestPayload, responsePayload, parentIDs...)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s/v1/resource/%s/subresource/%s", expectedProtocol, expectedHost, expectedBasePath, parentIDs[0], expectedID))
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
			Convey("And then client should have received the right request payload", func() {
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty1)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty1], ShouldEqual, expectedReqPayloadProperty1Value)
			})
		})
	})

}

func TestProviderClientGet(t *testing.T) {

	Convey("Given a providerClient set up with stub client that returns some response", t, func() {
		httpClient := &http_goclient.HttpClientStub{
			Response: &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`{"property1":"value1"}`)),
			},
		}
		headerParameter := SpecHeaderParam{"Operation-Specific-Header", "operation_specific_header"}
		providerConfiguration := providerConfiguration{
			Headers: map[string]string{headerParameter.TerraformName: "some-value"},
		}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := newStubAuthenticator(expectedHeader, expectedHeaderValue, nil)
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: newStubBackendConfiguration("wwww.host.com", "/api", []string{"http"}),
			httpClient:                  httpClient,
			providerConfiguration:       providerConfiguration,
			apiAuthenticator:            apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourceGetOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{headerParameter},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}

			responsePayload := map[string]interface{}{}
			expectedID := "1234"
			_, err := providerClient.Get(specStubResource, expectedID, responsePayload)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				expectedPath := specStubResource.path
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s%s/%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath, expectedID))
			})
			Convey("And then client should have received the right Authentication header and expected value", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
			Convey("And then client should have received the right operation header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, headerParameter.Name)
				So(httpClient.Headers[headerParameter.Name], ShouldEqual, providerConfiguration.Headers[headerParameter.TerraformName])
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
		})
	})

	Convey("Given a providerClient set up with stub client that returns some response", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration{},
			apiAuthenticator: &specStubAuthenticator{
				authContext: &authContext{
					headers: map[string]string{},
				},
			},
		}
		Convey("When providerClient GET  method is called with a SpecV2Resource that has a subresource path, a requestPayload, an empty responsePayload and the resource parentID", func() {
			specv2Resource := &SpecV2Resource{
				Path: "/v1/resource/{id}/subresource",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							OperationProps: spec.OperationProps{
								Responses: &spec.Responses{},
							},
						},
					},
				},
				InstancePathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Get: &spec.Operation{
							OperationProps: spec.OperationProps{
								Responses: &spec.Responses{},
							},
						},
					},
				},
			}
			responsePayload := map[string]interface{}{}
			parentIDs := []string{"parentID"}
			expectedID := "1234"
			_, err := providerClient.Get(specv2Resource, expectedID, responsePayload, parentIDs...)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s/v1/resource/%s/subresource/%s", expectedProtocol, expectedHost, expectedBasePath, parentIDs[0], expectedID))
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
		})
	})

}

func TestProviderClientDelete(t *testing.T) {

	Convey("Given a providerClient set up with stub client that returns some response", t, func() {
		httpClient := &http_goclient.HttpClientStub{
			Response: &http.Response{
				Body: ioutil.NopCloser(strings.NewReader(`{"property1":"value1"}`)),
			},
		}
		headerParameter := SpecHeaderParam{"Operation-Specific-Header", "operation_specific_header"}
		providerConfiguration := providerConfiguration{
			Headers: map[string]string{headerParameter.TerraformName: "some-value"},
		}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := newStubAuthenticator(expectedHeader, expectedHeaderValue, nil)
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: newStubBackendConfiguration("wwww.host.com", "/api", []string{"http"}),
			httpClient:                  httpClient,
			providerConfiguration:       providerConfiguration,
			apiAuthenticator:            apiAuthenticator,
		}
		Convey("When providerClient DELETE method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourceDeleteOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{headerParameter},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			expectedID := "1234"
			_, err := providerClient.Delete(specStubResource, expectedID)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				expectedPath := specStubResource.path
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s%s/%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath, expectedID))
			})
			Convey("And then client should have received the right Authentication header and expected value", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
			Convey("And then client should have received the right operation header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, headerParameter.Name)
				So(httpClient.Headers[headerParameter.Name], ShouldEqual, providerConfiguration.Headers[headerParameter.TerraformName])
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
		})
	})

	Convey("Given a providerClient set up with stub client that returns some response", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: &specStubBackendConfiguration{
				host:        "wwww.host.com",
				basePath:    "/api",
				httpSchemes: []string{"http"},
			},
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration{},
			apiAuthenticator: &specStubAuthenticator{
				authContext: &authContext{
					headers: map[string]string{},
				},
			},
		}
		Convey("When providerClient DELETE  method is called with a SpecV2Resource that has a subresource path, a requestPayload, an empty responsePayload and the resource parentID", func() {
			specv2Resource := &SpecV2Resource{
				Path: "/v1/resource/{id}/subresource",
				RootPathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							OperationProps: spec.OperationProps{
								Responses: &spec.Responses{},
							},
						},
					},
				},
				InstancePathItem: spec.PathItem{
					PathItemProps: spec.PathItemProps{
						Delete: &spec.Operation{
							OperationProps: spec.OperationProps{
								Responses: &spec.Responses{},
							},
						},
					},
				},
			}
			parentIDs := []string{"parentID"}
			expectedID := "1234"
			_, err := providerClient.Delete(specv2Resource, expectedID, parentIDs...)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then client should have received the right URL", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(httpClient.URL, ShouldEqual, fmt.Sprintf("%s://%s%s/v1/resource/%s/subresource/%s", expectedProtocol, expectedHost, expectedBasePath, parentIDs[0], expectedID))
			})
			Convey("And then client should have received the right User-Agent header and the expected value", func() {
				So(httpClient.Headers, ShouldContainKey, userAgent)
				So(httpClient.Headers[userAgent], ShouldContainSubstring, "OpenAPI Terraform Provider")
			})
		})
	})
}
