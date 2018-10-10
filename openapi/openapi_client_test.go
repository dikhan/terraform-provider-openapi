package openapi

import (
	"fmt"
	"github.com/dikhan/http_goclient"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
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

func TestPerformRequest(t *testing.T) {
	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
		httpClient := &http_goclient.HttpClientStub{}
		providerConfiguration := providerConfiguration{}
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
		Convey("When performRequest POST method is called with a resourceURL, a requestPayload and an empty responsePayload", func() {
			resourcePostOperation := &specResourceOperation{
				HeaderParameters: SpecHeaderParameters{},
				responses:  specResponses{},
				SecuritySchemes: SpecSecuritySchemes{},
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
			Convey("And then client should have received the right Headers with the right values", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
			Convey("And then client should have received the right request payload", func() {
				So(httpClient.In.(map[string]interface{}), ShouldContainKey, expectedReqPayloadProperty1)
				So(httpClient.In.(map[string]interface{})[expectedReqPayloadProperty1], ShouldEqual, expectedReqPayloadProperty1Value)
			})
		})
		Convey("When performRequest with a method that is not supported", func() {
			resourcePostOperation := &specResourceOperation{
				HeaderParameters: SpecHeaderParameters{},
				responses:  specResponses{},
				SecuritySchemes: SpecSecuritySchemes{},
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
				apiAuthenticator:      &specStubAuthenticator{
					authContext:&authContext{},
					err: fmt.Errorf("some error with prep auth"),
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
		providerConfiguration := providerConfiguration{}
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
					HeaderParameters: SpecHeaderParameters{},
					responses:  specResponses{},
					SecuritySchemes: SpecSecuritySchemes{},
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
			Convey("And then client should have received the right Headers with the right values", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
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
		providerConfiguration := providerConfiguration{}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := newStubAuthenticator(expectedHeader, expectedHeaderValue, nil)
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: newStubBackendConfiguration("wwww.host.com", "/api", []string{"http"}),
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourcePutOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:  specResponses{},
					SecuritySchemes: SpecSecuritySchemes{},
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
			Convey("And then client should have received the right Headers with the right values", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
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
		providerConfiguration := providerConfiguration{}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := newStubAuthenticator(expectedHeader, expectedHeaderValue, nil)
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: newStubBackendConfiguration("wwww.host.com", "/api", []string{"http"}),
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourceGetOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:  specResponses{},
					SecuritySchemes: SpecSecuritySchemes{},
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
			Convey("And then client should have received the right Headers with the right values", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
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
		providerConfiguration := providerConfiguration{}
		expectedHeader := "Authentication"
		expectedHeaderValue := "Bearer secret!"
		apiAuthenticator := newStubAuthenticator(expectedHeader, expectedHeaderValue, nil)
		providerClient := &ProviderClient{
			openAPIBackendConfiguration: newStubBackendConfiguration("wwww.host.com", "/api", []string{"http"}),
			httpClient:            httpClient,
			providerConfiguration: providerConfiguration,
			apiAuthenticator:      apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourceDeleteOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:  specResponses{},
					SecuritySchemes: SpecSecuritySchemes{},
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
			Convey("And then client should have received the right Headers with the right values", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
		})
	})
}