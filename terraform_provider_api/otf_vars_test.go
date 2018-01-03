package main

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
	Convey("Given OTF_VAR_test_SWAGGER_URL is set using lower case provider name", t, func() {
		os.Setenv(otfVarNameLc, otfVarSwaggerURLValue)
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			apiDiscoveryURL, err := getServiceProviderSwaggerURL(providerName)
			Convey("The apiDiscoveryURL returned should contain the URL and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameLc)
	})
}

func TestGetServiceProviderSwaggerUrlUpperCase(t *testing.T) {
	Convey("Given OTF_VAR_TEST_SWAGGER_URL is set using upper case provider name", t, func() {
		os.Setenv(otfVarNameUc, otfVarSwaggerURLValue)
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			apiDiscoveryURL, err := getServiceProviderSwaggerURL(providerName)
			Convey("The apiDiscoveryURL returned should contain the URL and error should be nil", func() {
				So(err, ShouldBeNil)
				So(apiDiscoveryURL, ShouldEqual, otfVarSwaggerURLValue)
			})
		})
		os.Unsetenv(otfVarNameUc)
	})
}

func TestGetServiceProviderSwaggerUrlEnvNotSet(t *testing.T) {
	Convey(fmt.Sprintf("Given OTF_VAR_test_SWAGGER_URL env variable is not set"), t, func() {
		Convey("When getServiceProviderSwaggerURL is called with provider name 'test'", func() {
			_, err := getServiceProviderSwaggerURL(providerName)
			Convey("The error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
