package openapi

import (
	"fmt"
	"os"
	"testing"

	"strings"

	. "github.com/smartystreets/goconvey/convey"
)

const providerName = "test"
const otfVarSwaggerURLValue = "http://host.com/swagger.yaml"

var otfVarNameLc = fmt.Sprintf(otfVarSwaggerURL, providerName)
var otfVarNameUc = fmt.Sprintf(otfVarSwaggerURL, strings.ToUpper(providerName))

func TestGetServiceProviderSwaggerUrlLowerCase(t *testing.T) {
	Convey("Given a PluginConfiguration for 'test' provider and a OTF_VAR_test_SWAGGER_URL is set using lower case provider name", t, func() {
		pluginConfiguration, _ := NewPluginConfiguration(providerName)
		os.Setenv(otfVarNameLc, otfVarSwaggerURLValue)
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			apiDiscoveryURL, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The apiDiscoveryURL returned should contain the URL and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameLc)
	})
	Convey("Given a PluginConfiguration for 'test' provider and a OTF_VAR_TEST_SWAGGER_URL is set using upper case provider name", t, func() {
		pluginConfiguration, _ := NewPluginConfiguration(providerName)
		os.Setenv(otfVarNameUc, otfVarSwaggerURLValue)
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			apiDiscoveryURL, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The apiDiscoveryURL returned should contain the URL and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameUc)
	})
	Convey(fmt.Sprintf("Given a PluginConfiguration for 'test' provider and a OTF_VAR_test_SWAGGER_URL env variable is not set"), t, func() {
		pluginConfiguration, _ := NewPluginConfiguration(providerName)
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			_, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider, a OTF_VAR_test_SWAGGER_URL is set using lower case provider name and a plugin configuration file containing a service called 'test'", t, func() {
		pluginConfig := fmt.Sprintf(`version: '1'
services:
    %s: %s`, providerName, "http://some-other-api/swagger.yaml")
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		os.Setenv(otfVarNameLc, otfVarSwaggerURLValue)
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			apiDiscoveryURL, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The apiDiscoveryURL returned should contain the URL of the environment variable as it takes preference and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameLc)
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing a service called 'test' and OTF_VAR_test_SWAGGER_URL not being set", t, func() {
		pluginConfig := fmt.Sprintf(`version: '1'
services:
    %s: %s`, providerName, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			apiDiscoveryURL, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The apiDiscoveryURL returned should contain the URL and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
	})
	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration that DOES NOT contain a service called 'test'", t, func() {
		pluginConfig := `version: '1'
services:
    some-other-service: http://some-other-service-api/swagger.yaml`
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			_, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "'test' not found in provider's services configuration")
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
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			_, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to unmarshall terraform-provider-openapi.yaml configuration file - error = yaml: found character that cannot start any token")
			})
		})
	})

	Convey("Given a PluginConfiguration for 'test' provider and a plugin configuration file containing a wrong version", t, func() {
		pluginConfig := fmt.Sprintf(`version: '3'
services:
    %s: %s`, providerName, otfVarSwaggerURLValue)
		configReader := strings.NewReader(pluginConfig)
		pluginConfiguration := PluginConfiguration{
			ProviderName:  providerName,
			Configuration: configReader,
		}
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			_, err := pluginConfiguration.getServiceProviderSwaggerURL()
			Convey("The the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "error occurred while validating terraform-provider-openapi.yaml - error = provider configuration version not matching current implementation, please use version '1' of provider configuration specification")
			})
		})
	})
}
