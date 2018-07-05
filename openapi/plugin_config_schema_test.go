package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewPluginConfigSchemaV1(t *testing.T) {
	Convey("Given a schema version and a map of services and their swagger URLs", t, func() {
		version := "1"
		services := map[string]string{
			"test": "http://sevice-api.com/swagger.yaml",
		}
		Convey("When NewPluginConfigSchemaV1 method is called", func() {
			pluginConfigSchemaV1 := NewPluginConfigSchemaV1(version, services)
			Convey("And the pluginConfigSchema returned should implement PluginConfigSchema interface", func() {
				var _ PluginConfigSchema = pluginConfigSchemaV1
			})
		})
	})
}

func TestPluginConfigSchemaV1Validate(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		version := "1"
		services := map[string]string{
			"test": "http://sevice-api.com/swagger.yaml",
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(version, services)
		Convey("When Validate method is called", func() {
			err := pluginConfigSchema.Validate()
			Convey("Then the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
			})
		})
	})
	Convey("Given a PluginConfigSchemaV1 containing a version that is NOT supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		version := "2"
		services := map[string]string{
			"test": "http://sevice-api.com/swagger.yaml",
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(version, services)
		Convey("When Validate method is called", func() {
			err := pluginConfigSchema.Validate()
			Convey("Then the error returned should NOT be nil as version is not supported", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a PluginConfigSchemaV1 containing a version that is supported but some services contain NON valid URLs", t, func() {
		var pluginConfigSchema PluginConfigSchema
		version := "2"
		services := map[string]string{
			"test": "htpt:/non-valid-url",
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(version, services)
		Convey("When Validate method is called", func() {
			err := pluginConfigSchema.Validate()
			Convey("Then the error returned should NOT be nil as service URL is not correct", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestPluginConfigSchemaV1GetProviderURL(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		version := "1"
		expectedURL := "http://sevice-api.com/swagger.yaml"
		services := map[string]string{
			"test": expectedURL,
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(version, services)
		Convey("When GetProviderURL method is called with a service described in the configuration", func() {
			swaggerURL, err := pluginConfigSchema.GetProviderURL("test")
			Convey("Then the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the url returned should be equal to the one in the plugin configuration", func() {
				So(swaggerURL, ShouldEqual, expectedURL)
			})
		})
		Convey("When GetProviderURL method is called with a service that DOES NOT exist in the plugin configuration", func() {
			_, err := pluginConfigSchema.GetProviderURL("non-existing-service")
			Convey("Then the error returned should be nil as configuration is correct", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
