package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"testing"
)

func TestNewServiceConfigV1(t *testing.T) {
	Convey("Given a service swagger url and whether the API has insecureSkipVerifyEnabled", t, func() {
		url := "http://host.com/swagger.json"
		insecureSkipVerifyEnabled := true
		Convey("When NewServiceConfigV1 method is called", func() {
			pluginConfigSchemaV1 := NewServiceConfigV1(url, insecureSkipVerifyEnabled)
			Convey("And the pluginConfigSchema returned should implement PluginConfigSchema interface", func() {
				var _ ServiceConfiguration = pluginConfigSchemaV1
			})
		})
	})
}

func TestServiceConfigV1GetSwaggerURL(t *testing.T) {
	Convey("Given a ServiceConfigV1 containing a swagger file", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedSwaggerURL := "http://sevice-api.com/swagger.yaml"
		serviceConfiguration = NewServiceConfigV1(expectedSwaggerURL, false)
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
		serviceConfiguration = NewServiceConfigV1(expectedSwaggerURL, expectedIsSecureSkipVerifyEnabled)
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
