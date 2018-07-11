package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewPluginConfigSchemaV1(t *testing.T) {
	Convey("Given a schema version and a map of services and their swagger URLs", t, func() {
		version := "1"
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         "http://sevice-api.com/swagger.yaml",
				InsecureSkipVerify: true,
			},
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
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         "http://sevice-api.com/swagger.yaml",
				InsecureSkipVerify: true,
			},
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
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         "http://sevice-api.com/swagger.yaml",
				InsecureSkipVerify: true,
			},
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
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         "htpt:/non-valid-url",
				InsecureSkipVerify: true,
			},
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

func TestPluginConfigSchemaV1GetServiceConfig(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		version := "1"
		expectedURL := "http://sevice-api.com/swagger.yaml"
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: true,
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(version, services)
		Convey("When GetServiceConfig method is called with a service described in the configuration", func() {
			serviceConfig, err := pluginConfigSchema.GetServiceConfig("test")
			Convey("Then the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfig should not be nil", func() {
				So(serviceConfig, ShouldNotBeNil)
			})
			Convey("And the url returned should be equal to the one in the service configuration", func() {
				So(serviceConfig.GetSwaggerURL(), ShouldEqual, expectedURL)
			})
		})
		Convey("When GetServiceConfig method is called with a service that DOES NOT exist in the plugin configuration", func() {
			_, err := pluginConfigSchema.GetServiceConfig("non-existing-service")
			Convey("Then the error returned should not be nil as provider specified does not exist in configuration file", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestPluginConfigSchemaV1GetVersion(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		version := "1"
		expectedURL := "http://sevice-api.com/swagger.yaml"
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: true,
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(version, services)
		Convey("When GetVersion method is called", func() {
			configVersion, err := pluginConfigSchema.GetVersion()
			Convey("Then the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfig should not be nil", func() {
				So(configVersion, ShouldEqual, "1")
			})
		})
	})
}

func TestPluginConfigSchemaV1GetAllServiceConfigurations(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		version := "1"
		expectedURL := "http://sevice-api.com/swagger.yaml"
		serviceConfigName := "test"
		services := map[string]*ServiceConfigV1{
			serviceConfigName: {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: true,
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(version, services)
		Convey("When GetVersion method is called", func() {
			serviceConfigurations, err := pluginConfigSchema.GetAllServiceConfigurations()
			Convey("Then the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the serviceConfigurations contain 1 configuration", func() {
				So(len(serviceConfigurations), ShouldEqual, 1)
			})
			Convey("And the serviceConfigurations item should be test", func() {
				So(serviceConfigurations[serviceConfigName], ShouldNotBeNil)
			})
		})
	})
}
