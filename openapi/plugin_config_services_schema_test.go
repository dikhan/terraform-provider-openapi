package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
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

func TestNewServiceConfigV1WithDefaults(t *testing.T) {
	Convey("Given a service swagger url", t, func() {
		url := "http://host.com/swagger.json"
		Convey("When NewServiceConfigV1WithDefaults method is called", func() {
			pluginConfigSchemaV1 := NewServiceConfigV1WithDefaults(url)
			Convey("And the pluginConfigSchema returned should implement PluginConfigSchema interface", func() {
				var _ ServiceConfiguration = pluginConfigSchemaV1
			})
		})
	})
}

func TestPluginConfigSchemaV1GetSwaggerURL(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var serviceConfiguration ServiceConfiguration
		expectedSwaggerURL := "http://sevice-api.com/swagger.yaml"
		serviceConfiguration = NewServiceConfigV1WithDefaults(expectedSwaggerURL)
		Convey("When GetSwaggerURL method is called", func() {
			swaggerURL := serviceConfiguration.GetSwaggerURL()
			Convey("Then the swagger url returned should be equal to expected one", func() {
				So(swaggerURL, ShouldEqual, expectedSwaggerURL)
			})
		})
	})
}

func TestPluginConfigSchemaV1IsSecureSkipVerifyEnabled(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
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
