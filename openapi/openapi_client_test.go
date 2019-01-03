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

func TestGetResourceIDURL(t *testing.T) {
	Convey("Given a providerClient set up with stub auth that injects some headers to the request", t, func() {
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
		Convey("When getResourceIDURL with a specResource and and ID", func() {
			expectedID := "1234"
			expectedPath := "/v1/resource"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceIDURL(specStubResource, expectedID)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then resourceURL should equal", func() {
				expectedProtocol := providerClient.openAPIBackendConfiguration.getHTTPSchemes()[0]
				expectedHost, _ := providerClient.openAPIBackendConfiguration.getHost()
				expectedBasePath := providerClient.openAPIBackendConfiguration.getBasePath()
				So(resourceURL, ShouldEqual, fmt.Sprintf("%s://%s%s%s/%s", expectedProtocol, expectedHost, expectedBasePath, expectedPath, expectedID))
			})
		})

		Convey("When getResourceIDURL with a specResource containing trailing / in the path and and ID", func() {
			expectedID := "1234"
			expectedPath := "/v1/resource/"
			specStubResource := &specStubResource{
				path: expectedPath,
				resourcePostOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
					responses:        specResponses{},
					SecuritySchemes:  SpecSecuritySchemes{},
				},
			}
			resourceURL, err := providerClient.getResourceIDURL(specStubResource, expectedID)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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

		Convey("When getResourceURL with a specResource with a resource path and overrides the global host", func() {
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			_, err := providerClient.getResourceURL(specStubResource)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
		regionProperty := newStringSchemaDefinitionPropertyWithDefaults(providerPropertyRegion, "", true, false, expectedRegion)
		s := newTestSchema(regionProperty)
		providerConfiguration := providerConfiguration{
			data: s.getResourceData(t),
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
		emptyRegionProvidedByUser := ""
		expectedRegion := "us-east1"
		regionProperty := newStringSchemaDefinitionPropertyWithDefaults(providerPropertyRegion, "", true, false, emptyRegionProvidedByUser)
		s := newTestSchema(regionProperty)
		providerConfiguration := providerConfiguration{
			data: s.getResourceData(t),
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
			resourceURL, err := providerClient.getResourceURL(specStubResource)
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
			_, err := providerClient.getResourceURL(specStubResource)
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
		emptyRegionProvidedByUser := ""
		regionProperty := newStringSchemaDefinitionPropertyWithDefaults(providerPropertyRegion, "", true, false, emptyRegionProvidedByUser)
		s := newTestSchema(regionProperty)
		providerConfiguration := providerConfiguration{
			data: s.getResourceData(t),
		}
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
			_, err := providerClient.getResourceURL(specStubResource)
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
		regionProperty := newStringSchemaDefinitionPropertyWithDefaults(providerPropertyRegion, "", true, false, "us-east1")
		s := newTestSchema(regionProperty)
		providerConfiguration := providerConfiguration{
			data: s.getResourceData(t),
		}
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
			_, err := providerClient.getResourceURL(specStubResource)
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
			_, err := providerClient.getResourceURL(specStubResource)
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
			httpClient:                  httpClient,
			providerConfiguration:       providerConfiguration,
			apiAuthenticator:            apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourcePutOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
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
			httpClient:                  httpClient,
			providerConfiguration:       providerConfiguration,
			apiAuthenticator:            apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourceGetOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
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
			httpClient:                  httpClient,
			providerConfiguration:       providerConfiguration,
			apiAuthenticator:            apiAuthenticator,
		}
		Convey("When providerClient PUT method is called with a specStubResource that does not override the host, a requestPayload and an empty responsePayload", func() {
			specStubResource := &specStubResource{
				path: "/v1/resource",
				resourceDeleteOperation: &specResourceOperation{
					HeaderParameters: SpecHeaderParameters{},
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
			Convey("And then client should have received the right Headers with the right values", func() {
				So(httpClient.Headers, ShouldContainKey, expectedHeader)
				So(httpClient.Headers[expectedHeader], ShouldEqual, expectedHeaderValue)
			})
		})
	})
}
