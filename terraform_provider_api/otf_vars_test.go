package main

import (
	"fmt"
	"os"
	"testing"

	"strings"

	. "github.com/smartystreets/goconvey/convey"
)

const PROVIDER_NAME = "test"
const OTF_VAR_SWAGGER_URL_VALUE = "http://host.com/swagger.yaml"

var OTF_VAR_NAME_LC = fmt.Sprintf(OTF_VAR_SWAGGER_URL, PROVIDER_NAME)
var OTF_VAR_NAME_UC = fmt.Sprintf(OTF_VAR_SWAGGER_URL, strings.ToUpper(PROVIDER_NAME))

func TestGetServiceProviderSwaggerUrlLowerCase(t *testing.T) {
	Convey("Given OTF_VAR_test_SWAGGER_URL is set using lower case provider name", t, func() {
		os.Setenv(OTF_VAR_NAME_LC, OTF_VAR_SWAGGER_URL_VALUE)
		Convey("When GetServiceProviderSwaggerUrl is called with provider name 'test'", func() {
			apiDiscoveryUrl, err := GetServiceProviderSwaggerUrl(PROVIDER_NAME)
			Convey("The apiDiscoveryUrl returned should contain the URL and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryUrl, ShouldEqual, OTF_VAR_SWAGGER_URL_VALUE)
			})
		})
		os.Unsetenv(OTF_VAR_NAME_LC)
	})
}

func TestGetServiceProviderSwaggerUrlUpperCase(t *testing.T) {
	Convey("Given OTF_VAR_TEST_SWAGGER_URL is set using upper case provider name", t, func() {
		os.Setenv(OTF_VAR_NAME_UC, OTF_VAR_SWAGGER_URL_VALUE)
		Convey("When GetServiceProviderSwaggerUrl is called with provider name 'test'", func() {
			apiDiscoveryUrl, err := GetServiceProviderSwaggerUrl(PROVIDER_NAME)
			Convey("The apiDiscoveryUrl returned should contain the URL and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryUrl, ShouldEqual, OTF_VAR_SWAGGER_URL_VALUE)
			})
		})
		os.Unsetenv(OTF_VAR_NAME_UC)
	})
}

func TestGetServiceProviderSwaggerUrlEnvNotSet(t *testing.T) {
	Convey(fmt.Sprintf("Given OTF_VAR_test_SWAGGER_URL env variable is not set"), t, func() {
		Convey("When GetServiceProviderSwaggerUrl is called with provider name 'test'", func() {
			_, err := GetServiceProviderSwaggerUrl(PROVIDER_NAME)
			Convey("The error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
