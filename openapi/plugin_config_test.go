package openapi

import (
	"bytes"
	"fmt"
	"github.com/smartystreets/assertions/should"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the pluginConfigurationFile returned should not be nil ", func() {
				So(pluginConfigurationFile, ShouldEqual, otfVarPluginConfigurationFileValue)
			})
		})
		os.Unsetenv(otfVarPluginConfigurationFileLc)
	})
	Convey("Given an environment variable set using lower case provider name with the plugin configuration file path", t, func() {
		os.Setenv(otfVarPluginConfigurationFileUc, otfVarPluginConfigurationFileValue)
		Convey("When getPluginConfigurationPath is called", func() {
			pluginConfigurationFile, err := getPluginConfigurationPath(providerName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the pluginConfigurationFile returned should not be nil ", func() {
				So(pluginConfigurationFile, ShouldEqual, otfVarPluginConfigurationFileValue)
			})
		})
		os.Unsetenv(otfVarPluginConfigurationFileUc)
	})
	Convey("Given no environment variables set for the plugin configuration file", t, func() {
		Convey("When getPluginConfigurationPath is called", func() {
			pluginConfigurationFile, err := getPluginConfigurationPath(providerName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the pluginConfigurationFile returned should not be nil ", func() {
				So(pluginConfigurationFile, ShouldContainSubstring, ".terraform.d/plugins/terraform-provider-openapi.yaml")
			})
		})
	})
}

func TestGetServiceProviderConfiguration(t *testing.T) {
	Convey("Given a PluginConfiguration for 'test' provider and a OTF_VAR_test_SWAGGER_URL is set using lower case provider name", t, func() {
		pluginConfiguration, _ := NewPluginConfiguration(providerName)
		os.Setenv(otfVarNameLc, otfVarSwaggerURLValue)
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfiguration returned should not be nil ", func() {
				So(serviceConfiguration, ShouldNotBeNil)
			})
			Convey("And the serviceConfiguration returned should contain a service 'test' with a swagger URL", func() {
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfiguration returned should not be nil ", func() {
				So(serviceConfiguration, ShouldNotBeNil)
			})
			Convey("And the pluginConfig returned should contain a service 'test' with a swagger URL", func() {
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfiguration returned should not be nil ", func() {
				So(serviceConfiguration, ShouldNotBeNil)
			})
			Convey("And the serviceConfiguration returned should contain the URL of the environment variable as it takes preference and error should be nil", func() {
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfiguration returned should not be nil ", func() {
				So(serviceConfiguration, ShouldNotBeNil)
			})
			Convey("And the serviceConfiguration returned should contain the URL and error should be nil", func() {
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing telemetry configured and a service called 'test'", t, func() {
		metricChannel := make(chan string)
		pc, telemetryHost, telemetryPort := udpServer(metricChannel)
		defer pc.Close()

		pluginConfig := fmt.Sprintf(`version: '1'
services:
    %s:
      telemetry:
        graphite:
          host: %s
          port: %s
          prefix: openapi
      swagger-url: %s`, providerName, telemetryHost, telemetryPort, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			serviceConfiguration, err := pluginConfiguration.getServiceConfiguration()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfiguration returned should not be nil ", func() {
				So(serviceConfiguration, ShouldNotBeNil)
			})
			Convey("And the serviceConfiguration returned should contain the URL and error should be nil", func() {
				serviceSwaggerURL := serviceConfiguration.GetSwaggerURL()
				So(serviceSwaggerURL, ShouldEqual, otfVarSwaggerURLValue)
			})
			Convey("And the serviceConfiguration contains the expected graphite telemetry configuration", func() {
				port, _ := strconv.Atoi(telemetryPort)
				expectedGraphiteProvider := &TelemetryProviderGraphite{
					Host:   telemetryHost,
					Port:   port,
					Prefix: "openapi",
				}
				So(serviceConfiguration.GetTelemetryConfiguration(), ShouldResemble, TelemetryProvider(expectedGraphiteProvider))
			})
			// TODO: This should be checked where the telemetry metric is moved
			//Convey("And the telemetry server should have been received the expected counter metrics increase", func() {
			//	assertExpectedMetric(t, metricChannel, "openapi.terraform.providers.test.total_runs:1|c")
			//	assertExpectedMetric(t, metricChannel, "openapi.terraform.openapi_plugin_version.dev.total_runs:1|c")
			//})
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
			Convey("The error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should containing the following message", func() {
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
			Convey("The error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should containing the following message", func() {
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
			Convey("The error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should containing the following message", func() {
				So(err.Error(), should.ContainSubstring, "error occurred while validating 'terraform-provider-openapi.yaml' - error = provider configuration version not matching current implementation, please use version '1' of provider configuration specification")
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing a service configuration that is invalid (e,g: plugin version is specified but does not match the version of the plugin executing)", t, func() {
		pluginConfig := fmt.Sprintf(`version: '1'
services:
    %s:
        swagger-url: %s
        plugin_version: 0.14.0`, providerName, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceConfiguration is called", func() {
			_, err := pluginConfiguration.getServiceConfiguration()
			Convey("The error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should containing the following message", func() {
				So(err.Error(), should.ContainSubstring, "service configuration for 'test' not valid: plugin version '0.14.0' in the plugin configuration file does not match the version of the OpenAPI plugin that is running 'dev'")
			})
		})
	})

}

func assertExpectedMetric(t *testing.T, metricChannel chan string, expectedMetric string) {
	assertExpectedMetricAndLogging(t, metricChannel, expectedMetric, "", "", nil)
}

func assertExpectedMetricAndLogging(t *testing.T, metricChannel chan string, expectedMetric, expectedLogMetricToSubmit, expectedLogMetricSuccess string, logging *bytes.Buffer) {
	select {
	case metricReceived := <-metricChannel:
		assert.Contains(t, metricReceived, expectedMetric)
		if expectedLogMetricToSubmit != "" {
			assert.Contains(t, logging.String(), expectedLogMetricToSubmit)
		}
		if expectedLogMetricSuccess != "" {
			assert.Contains(t, logging.String(), expectedLogMetricSuccess)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("[FAIL] '%s' not reveided within the expected timeframe (timed out)", expectedMetric)
	}
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
