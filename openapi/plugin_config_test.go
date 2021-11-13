package openapi

import (
	"errors"
	"fmt"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"log"
	"net"
	"os"
	"strings"
	"testing"
)

const providerName = "test"
const otfVarSwaggerURLValue = "http://host.com/swagger.yaml"
const otfVarPluginConfigurationFileValue = "/some/path/terraform-provider-openapi.yaml"

var otfVarNameLc = fmt.Sprintf(otfVarSwaggerURL, providerName)
var otfVarNameUc = strings.ToUpper(otfVarNameLc)

func TestGetPluginConfigurationPath(t *testing.T) {
	var otfVarPluginConfigurationFileLc = fmt.Sprintf(otfVarPluginConfigurationFile, providerName)
	var otfVarPluginConfigurationFileUc = strings.ToUpper(otfVarPluginConfigurationFileLc)
	Convey("Given an environment variable set using lower case provider name with the plugin configuration file path", t, func() {
		os.Setenv(otfVarPluginConfigurationFileLc, otfVarPluginConfigurationFileValue)
		Convey("When getPluginConfigurationPath is called", func() {
			pluginConfigurationFile, err := getPluginConfigurationPath(providerName)
			Convey("Then the error returned should be nil and the pluginConfigurationFile returned should be match the env variable value", func() {
				So(err, ShouldBeNil)
				So(pluginConfigurationFile, ShouldEqual, otfVarPluginConfigurationFileValue)
			})
		})
		os.Unsetenv(otfVarPluginConfigurationFileLc)
	})
	Convey("Given an environment variable set using lower case provider name with the plugin configuration file path", t, func() {
		os.Setenv(otfVarPluginConfigurationFileUc, otfVarPluginConfigurationFileValue)
		Convey("When getPluginConfigurationPath is called", func() {
			pluginConfigurationFile, err := getPluginConfigurationPath(providerName)
			Convey("Then the error returned should be nil and the pluginConfigurationFile returned should be match the env variable value", func() {
				So(err, ShouldBeNil)
				So(pluginConfigurationFile, ShouldEqual, otfVarPluginConfigurationFileValue)
			})
		})
		os.Unsetenv(otfVarPluginConfigurationFileUc)
	})
	Convey("Given no environment variables set for the plugin configuration file", t, func() {
		Convey("When getPluginConfigurationPath is called", func() {
			pluginConfigurationFile, err := getPluginConfigurationPath(providerName)
			Convey("Then the error returned should be nil and the returned config file should be the default location", func() {
				So(err, ShouldBeNil)
				So(pluginConfigurationFile, ShouldContainSubstring, ".terraform.d/plugins/terraform-provider-openapi.yaml")
			})
		})
	})
}

type errReader struct {
	errorMessage string
}

func (e errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New(e.errorMessage)
}

func TestGetServiceProviderConfiguration(t *testing.T) {
	Convey("Given a PluginConfiguration for 'test' provider and a OTF_VAR_test_SWAGGER_URL is set using lower case provider name", t, func() {
		pluginConfiguration, _ := NewPluginConfiguration(providerName)
		os.Setenv(otfVarNameLc, otfVarSwaggerURLValue)
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the serviceConfiguration returned should contain a service 'test' with a swagger URL and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(serviceConfiguration, ShouldNotBeNil)
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameLc)
	})

	Convey("Given a PluginConfiguration for 'test' provider and a OTF_VAR_TEST_SWAGGER_URL is set using upper case provider name", t, func() {
		pluginConfiguration, _ := NewPluginConfiguration(providerName)
		os.Setenv(otfVarNameUc, otfVarSwaggerURLValue)
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the pluginConfig returned should contain a service 'test' with a swagger URL and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(serviceConfiguration, ShouldNotBeNil)
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameUc)
	})

	Convey(fmt.Sprintf("Given a PluginConfiguration for 'test' provider and a OTF_VAR_test_SWAGGER_URL env variable is not set"), t, func() {
		pluginConfiguration, _ := NewPluginConfiguration(providerName)
		Convey("When getServiceConfiguration is called", func() {
			_, err := pluginConfiguration.getServiceConfiguration()
			Convey("The error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider, a OTF_VAR_test_SWAGGER_URL is set using lower case provider name and a plugin configuration file containing a service called 'test'", t, func() {
		pluginConfig := fmt.Sprintf(`version: '1'
services:
   %s:
       swagger-url: %s`, providerName, "http://some-other-api/swagger.yaml")
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		os.Setenv(otfVarNameLc, otfVarSwaggerURLValue)
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the serviceConfiguration returned should contain the URL of the environment variable as it takes preference and error should be nil and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(serviceConfiguration, ShouldNotBeNil)
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameLc)
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing a service called 'test' and OTF_VAR_test_SWAGGER_URL not being set", t, func() {
		pluginConfig := fmt.Sprintf(`version: '1'
services:
   %s:
       swagger-url: %s`, providerName, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the serviceConfiguration returned should contain the URL and error should be nil and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(serviceConfiguration, ShouldNotBeNil)
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing telemetry configured and a service called 'test'", t, func() {
		expectedTelemetryHost := "some-host"
		expectedTelemetryPort := 2654
		pluginConfig := fmt.Sprintf(`version: '1'
services:
    %s:
      telemetry:
        graphite:
          host: %s
          port: %d
          prefix: openapi
      swagger-url: %s`, providerName, expectedTelemetryHost, expectedTelemetryPort, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then serviceConfiguration returned should contain the URL and contain the expected graphite telemetry configuration and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(serviceConfiguration, ShouldNotBeNil)
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
				expectedGraphiteProvider := &TelemetryProviderGraphite{
					Host:   expectedTelemetryHost,
					Port:   expectedTelemetryPort,
					Prefix: "openapi",
				}
				So(serviceConfiguration.GetTelemetryConfiguration(), ShouldResemble, TelemetryProvider(expectedGraphiteProvider))
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing http_endpoint telemetry configured and a service called 'test'", t, func() {
		expectedTelemetryHost := "http://some-host/v1/metrics"
		expectedPrefix := "openapi"
		pluginConfig := fmt.Sprintf(`version: '1'
services:
    %s:
      telemetry:
        http_endpoint:
          url: %s
          prefix: %s
      swagger-url: %s`, providerName, expectedTelemetryHost, expectedPrefix, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the serviceConfiguration shoudl be the expected and the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(serviceConfiguration, ShouldNotBeNil)
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
				expectedHTTPEndpointProvider := &TelemetryProviderHTTPEndpoint{
					URL:    expectedTelemetryHost,
					Prefix: expectedPrefix,
				}
				So(serviceConfiguration.GetTelemetryConfiguration(), ShouldResemble, TelemetryProvider(expectedHTTPEndpointProvider))
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration that DOES NOT contain a service called 'test'", t, func() {
		pluginConfig := fmt.Sprintf(`version: '1'
services:
    some-other-service:
        swagger-url: http://some-other-service-api/swagger.yaml`)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			_, err := pluginConfiguration.getServiceConfiguration()
			Convey("The error returned should containing the following message", func() {
				So(err.Error(), should.ContainSubstring, "'test' not found in provider's services configuration")
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing wrong formatter yaml", t, func() {
		pluginConfig := `	wrong yaml`
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			_, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the error should containing the following message", func() {
				So(err.Error(), should.ContainSubstring, "failed to unmarshall terraform-provider-openapi.yaml configuration file - error = yaml: found character that cannot start any token")
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing a non supported version", t, func() {
		pluginConfig := fmt.Sprintf(`version: '3'
services:
    %s:
        swagger-url: %s`, providerName, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			_, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the error should containing the following message", func() {
				So(err.Error(), should.ContainSubstring, "error occurred while validating 'terraform-provider-openapi.yaml' - error = provider configuration version not matching current implementation, please use version '1' of provider configuration specification")
			})
		})
	})

	Convey("Given a PluginConfiguration that fails to read", t, func() {
		errReader := errReader{"some error"}
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: errReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			_, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the error should containing the following message", func() {
				So(err.Error(), should.ContainSubstring, "failed to read terraform-provider-openapi.yaml configuration file")
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing a service called 'test' with non valid configuration", t, func() {
		pluginConfig := fmt.Sprintf(`version: '1'
services:
   %s:
       swagger-url: non-valid-url`, providerName)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			_, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the error should containing the following message", func() {
				So(err.Error(), should.ContainSubstring, "service configuration for 'test' not valid: service swagger URL configuration not valid ('non-valid-url'). URL must be either a valid formed URL or a path to an existing swagger file stored in the disk")
			})
		})
	})

}

func udpServer(metricChannel chan string) (net.PacketConn, string, string) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:")
	if err != nil {
		log.Fatal(err)
	}
	telemetryServer := pc.LocalAddr().String()
	telemetryHost := strings.Split(telemetryServer, ":")[0]
	telemetryPort := strings.Split(telemetryServer, ":")[1]
	go func() {
		for {
			buf := make([]byte, 1024)
			n, _, err := pc.ReadFrom(buf)
			if err != nil {
				continue
			}
			body := string(buf[:n])
			metricChannel <- body
		}
	}()
	return pc, telemetryHost, telemetryPort
}
