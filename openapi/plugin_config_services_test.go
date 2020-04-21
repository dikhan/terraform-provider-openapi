package openapi

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestNewServiceConfigV1(t *testing.T) {
	Convey("Given a service swagger url and whether the API has insecureSkipVerifyEnabled", t, func() {
		url := "http://host.com/swagger.json"
		insecureSkipVerifyEnabled := true
		Convey("When NewServiceConfigV1 method is called", func() {
			pluginConfigSchemaV1 := NewServiceConfigV1(url, insecureSkipVerifyEnabled, &TelemetryConfig{})
			Convey("And the pluginConfigSchema returned should implement PluginConfigSchema interface", func() {
				var _ ServiceConfiguration = pluginConfigSchemaV1
			})
			Convey("And the pluginConfigSchema should contain the configured telemetry", func() {
				So(pluginConfigSchemaV1.TelemetryConfig, ShouldNotBeNil)
			})
		})
	})
}

func TestServiceConfigV1GetSwaggerURL(t *testing.T) {
	Convey("Given a ServiceConfigV1 containing a swagger file", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedSwaggerURL := "http://sevice-api.com/swagger.yaml"
		serviceConfiguration = NewServiceConfigV1(expectedSwaggerURL, false, nil)
		Convey("When GetSwaggerURL method is called", func() {
			swaggerURL := serviceConfiguration.GetSwaggerURL()
			Convey("Then the swagger url returned should be equal to expected one", func() {
				So(swaggerURL, ShouldEqual, expectedSwaggerURL)
			})
		})
	})
}

func TestServiceConfigV1GetPluginVersion(t *testing.T) {
	Convey("Given a ServiceConfigV1 containing a specific plugin version", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedPluginVersion := "0.14.0"
		serviceConfiguration = &ServiceConfigV1{
			PluginVersion: expectedPluginVersion,
		}
		Convey("When GetPluginVersion method is called", func() {
			pluginVersion := serviceConfiguration.GetPluginVersion()
			Convey("Then the plugin version returned should be equal to expected one", func() {
				So(pluginVersion, ShouldEqual, expectedPluginVersion)
			})
		})
	})
}

func TestServiceConfigV1IsSecureSkipVerifyEnabled(t *testing.T) {
	Convey("Given a ServiceConfigV1 containing the insecure_skip_verify enabled", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedSwaggerURL := "http://sevice-api.com/swagger.yaml"
		expectedIsSecureSkipVerifyEnabled := true
		serviceConfiguration = NewServiceConfigV1(expectedSwaggerURL, expectedIsSecureSkipVerifyEnabled, nil)
		Convey("When IsInsecureSkipVerifyEnabled method is called", func() {
			isInsecureSkipVerifyEnabled := serviceConfiguration.IsInsecureSkipVerifyEnabled()
			Convey("Then the IsSecureSkipVerifyEnabled returned should be equal to expected one", func() {
				So(isInsecureSkipVerifyEnabled, ShouldEqual, expectedIsSecureSkipVerifyEnabled)
			})
		})
	})
}

func TestServiceConfigV1Validate(t *testing.T) {
	Convey("Given a ServiceConfigV1 containing a valid swagger URL and a specific plugin version", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedPluginVersion := "0.14.0"
		expectedSwaggerURL := "http://sevice-api.com/swagger.yaml"
		serviceConfiguration = &ServiceConfigV1{
			SwaggerURL:    expectedSwaggerURL,
			PluginVersion: expectedPluginVersion,
		}
		Convey("When Validate method is called with a running version that matches the configured one", func() {
			runningPluginVersion := expectedPluginVersion
			err := serviceConfiguration.Validate(runningPluginVersion)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
	Convey("Given a ServiceConfigV1 containing an invalid swagger URL pointing at a file store in the disk", t, func() {
		var serviceConfiguration ServiceConfiguration
		swaggerFile, _ := ioutil.TempFile("", "")
		defer func(swaggerFile *os.File) {
			_ = swaggerFile.Close()
			_ = os.RemoveAll(swaggerFile.Name())
		}(swaggerFile)
		expectedSwaggerURL := swaggerFile.Name()
		serviceConfiguration = &ServiceConfigV1{
			SwaggerURL: expectedSwaggerURL,
		}
		Convey("When Validate method is called", func() {
			err := serviceConfiguration.Validate("0.14.0")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
	Convey("Given a ServiceConfigV1 containing an empty plugin version", t, func() {
		expectedSwaggerURL := "http://a.valid.url"
		serviceConfiguration := &ServiceConfigV1{
			SwaggerURL:    expectedSwaggerURL,
			PluginVersion: "",
		}
		Convey("When Validate method is called", func() {
			err := serviceConfiguration.Validate("0.14.0")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
	Convey("Given a ServiceConfigV1 containing an invalid swagger URL", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedSwaggerURL := "htpt:/non-valid-url"
		serviceConfiguration = &ServiceConfigV1{
			SwaggerURL: expectedSwaggerURL,
		}
		Convey("When Validate method is called", func() {
			err := serviceConfiguration.Validate("0.14.0")
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "service swagger URL configuration not valid ('htpt:/non-valid-url'). URL must be either a valid formed URL or a path to an existing swagger file stored in the disk")
			})
		})
	})

	Convey("Given a ServiceConfigV1 containing a valid swagger file and a specific plugin version", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedPluginVersion := "0.14.0"
		expectedSwaggerURL := "http://sevice-api.com/swagger.yaml"
		serviceConfiguration = &ServiceConfigV1{
			SwaggerURL:    expectedSwaggerURL,
			PluginVersion: expectedPluginVersion,
		}
		Convey("When Validate method is called with a running version that DOES NOT match the configured one", func() {
			runningPluginVersion := "0.15.0"
			err := serviceConfiguration.Validate(runningPluginVersion)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "plugin version '0.14.0' in the plugin configuration file does not match the version of the OpenAPI plugin that is running '0.15.0'")
			})
		})
	})
}

func TestGetTelemetryConfiguration(t *testing.T) {
	testCases := []struct {
		name            string
		serviceConfigV1 *ServiceConfigV1
		inputPluginName string
		expectedType    interface{}
		expectedError   string
		expectedLogging []string
	}{
		{
			name: "service is configured correctly with a graphite provider",
			serviceConfigV1: &ServiceConfigV1{
				TelemetryConfig: &TelemetryConfig{
					Graphite: &TelemetryProviderGraphite{
						Host: "my-graphite.com",
						Port: 8125,
					},
				},
			},
			inputPluginName: "pluginName",
			expectedType:    &TelemetryProviderGraphite{},
			expectedLogging: []string{"[DEBUG] graphite telemetry provider enabled"},
		},
		{
			name: "service is configured correctly with a httpendpoint provider",
			serviceConfigV1: &ServiceConfigV1{
				TelemetryConfig: &TelemetryConfig{
					HTTPEndpoint: &TelemetryProviderHTTPEndpoint{
						URL: "http://telemetry.myhost.com/v1/metrics",
					},
				},
			},
			inputPluginName: "pluginName",
			expectedType:    &TelemetryProviderHTTPEndpoint{},
			expectedLogging: []string{"[DEBUG] http endpoint telemetry provider enabled"},
		},
		{
			name: "service is configured correctly with graphite and httpendpoint providers",
			serviceConfigV1: &ServiceConfigV1{
				TelemetryConfig: &TelemetryConfig{
					Graphite: &TelemetryProviderGraphite{
						Host: "my-graphite.com",
						Port: 8125,
					},
					HTTPEndpoint: &TelemetryProviderHTTPEndpoint{
						URL: "http://telemetry.myhost.com/v1/metrics",
					},
				},
			},
			inputPluginName: "pluginName",
			expectedType:    nil,
			expectedLogging: []string{"[WARN] ignoring telemetry due multiple telemetry providers configured (graphite and http_endpoint): select only one"},
		},
		{
			name: "service skips graphite telemetry due to the validation not passing",
			serviceConfigV1: &ServiceConfigV1{
				TelemetryConfig: &TelemetryConfig{
					Graphite: &TelemetryProviderGraphite{
						Host: "", // Configuration is missing the required host
						//Port: 8125,
					},
				},
			},
			inputPluginName: "pluginName",
			expectedType:    nil,
			expectedLogging: []string{"[WARN] ignoring graphite telemetry due to the following validation error: graphite telemetry configuration is missing a value for the 'host property'"},
		},
		{
			name: "service skips httpendpoint telemetry due to the validation not passing",
			serviceConfigV1: &ServiceConfigV1{
				TelemetryConfig: &TelemetryConfig{
					HTTPEndpoint: &TelemetryProviderHTTPEndpoint{
						URL: "", // Configuration is missing the required url
					},
				},
			},
			inputPluginName: "pluginName",
			expectedType:    nil,
			expectedLogging: []string{"[WARN] ignoring http endpoint telemetry due to the following validation error: http endpoint telemetry configuration is missing a value for the 'url property'"},
		},
		{
			name: "TelemetryConfig is nil",
			serviceConfigV1: &ServiceConfigV1{
				TelemetryConfig: nil,
			},
			inputPluginName: "pluginName",
			expectedType:    nil,
			expectedLogging: []string{"[DEBUG] telemetry not configured"},
		},
	}
	for _, tc := range testCases {
		var buf bytes.Buffer
		log.SetOutput(&buf)
		telemetryHandler := tc.serviceConfigV1.GetTelemetryConfiguration()
		assert.IsType(t, tc.expectedType, telemetryHandler, tc.name)
		for _, log := range tc.expectedLogging {
			assert.Contains(t, buf.String(), log, tc.name)
		}
	}
}
