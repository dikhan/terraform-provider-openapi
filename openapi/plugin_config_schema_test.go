package openapi

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPluginConfigSchemaV1(t *testing.T) {
	Convey("Given a list of services and version", t, func() {
		version := ""
		services := map[string]*ServiceConfigV1{}
		Convey("When PluginConfigSchemaV1 method is constructed", func() {
			pluginConfigSchemaV1 := &PluginConfigSchemaV1{
				Services: services,
				Version:  version,
			}
			Convey("Then the pluginConfigSchemaV1 should comply with PluginConfigSchema interface", func() {
				var _ PluginConfigSchema = pluginConfigSchemaV1
			})
		})
	})
}

func TestNewPluginConfigSchemaV1(t *testing.T) {
	Convey("Given a schema version and a map of services and their swagger URLs", t, func() {
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         "http://sevice-api.com/swagger.yaml",
				InsecureSkipVerify: true,
			},
		}
		Convey("When NewPluginConfigSchemaV1 method is called", func() {
			pluginConfigSchemaV1 := NewPluginConfigSchemaV1(services)
			Convey("And the pluginConfigSchema returned should implement PluginConfigSchema interface and have services configured", func() {
				var _ PluginConfigSchema = pluginConfigSchemaV1
				So(pluginConfigSchemaV1.Services, ShouldNotBeNil)
			})
		})
	})
}

func TestPluginConfigSchemaV1Validate(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         "http://sevice-api.com/swagger.yaml",
				InsecureSkipVerify: true,
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(services)
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
		pluginConfigSchema = &PluginConfigSchemaV1{
			Version:  version,
			Services: services,
		}
		Convey("When Validate method is called", func() {
			err := pluginConfigSchema.Validate()
			Convey("And the error returned be equal to", func() {
				So(err.Error(), ShouldEqual, "provider configuration version not matching current implementation, please use version '1' of provider configuration specification")
			})
		})
	})
}

func TestPluginConfigSchemaV1GetServiceConfig(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		expectedURL := "http://sevice-api.com/swagger.yaml"
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: true,
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(services)
		Convey("When GetServiceConfig method is called with a service described in the configuration", func() {
			serviceConfig, err := pluginConfigSchema.GetServiceConfig("test")
			Convey("Then serviceConfig should not be nil, the url returned should be equal to the one in the service configuration and the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
				So(serviceConfig, ShouldNotBeNil)
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
		expectedURL := "http://sevice-api.com/swagger.yaml"
		services := map[string]*ServiceConfigV1{
			"test": {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: true,
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(services)
		Convey("When GetVersion method is called", func() {
			configVersion, err := pluginConfigSchema.GetVersion()
			Convey("Then the serviceConfig should not be nil and error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
				So(configVersion, ShouldEqual, "1")
			})
		})
	})
}

func TestPluginConfigSchemaV1GetAllServiceConfigurations(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services", t, func() {
		var pluginConfigSchema PluginConfigSchema
		expectedURL := "http://sevice-api.com/swagger.yaml"
		serviceConfigName := "test"
		services := map[string]*ServiceConfigV1{
			serviceConfigName: {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: true,
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(services)
		Convey("When GetAllServiceConfigurations method is called", func() {
			serviceConfigurations, err := pluginConfigSchema.GetAllServiceConfigurations()
			Convey("Then the serviceConfigurations contain 1 configuration, serviceConfigurations item should be test and error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
				So(len(serviceConfigurations), ShouldEqual, 1)
				So(serviceConfigurations[serviceConfigName], ShouldNotBeNil)
			})
		})
	})
}

func TestPluginConfigSchemaV1Marshal(t *testing.T) {
	Convey("Given a PluginConfigSchemaV1 containing a version supported and some services containing telemetry", t, func() {
		var pluginConfigSchema PluginConfigSchema
		expectedURL := "http://sevice-api.com/swagger.yaml"
		serviceConfigName := "test"
		expectedInscureSkipVerify := true
		services := map[string]*ServiceConfigV1{
			serviceConfigName: {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: expectedInscureSkipVerify,
				TelemetryConfig: &TelemetryConfig{
					Graphite: &TelemetryProviderGraphite{
						Host:   "some-host.com",
						Port:   8080,
						Prefix: "some_prefix",
					},
					HTTPEndpoint: &TelemetryProviderHTTPEndpoint{
						URL:    "http://my-api.com/v1/metrics",
						Prefix: "some_prefix",
					},
				},
				SchemaConfigurationV1: []ServiceSchemaPropertyConfigurationV1{
					{
						SchemaPropertyName: "apikey_auth",
						DefaultValue:       "apiKeyValue",
						Command:            []string{"echo", "something"},
						CommandTimeout:     10,
						ExternalConfiguration: ServiceSchemaPropertyExternalConfigurationV1{
							File:        "some_file",
							KeyName:     "some_key_name",
							ContentType: "json",
						},
					},
				},
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(services)
		Convey("When Marshal method is called", func() {
			marshalConfig, err := pluginConfigSchema.Marshal()
			Convey("Then the marshalConfig should contain the right marshal configuration and the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
				expectedConfig := fmt.Sprintf(`version: "1"
services:
  test:
    swagger-url: %s
    insecure_skip_verify: %t
    schema_configuration:
    - schema_property_name: apikey_auth
      default_value: apiKeyValue
      cmd: [echo, something]
      cmd_timeout: 10
      schema_property_external_configuration:
        file: some_file
        key_name: some_key_name
        content_type: json
    telemetry:
      graphite:
        host: some-host.com
        port: 8080
        prefix: some_prefix
      http_endpoint:
        url: http://my-api.com/v1/metrics
        prefix: some_prefix
`, expectedURL, expectedInscureSkipVerify)
				So(string(marshalConfig), ShouldEqual, expectedConfig)
			})
		})
	})

	Convey("Given a PluginConfigSchemaV1 containing a version supported and a service that does not specify a fix plugin version", t, func() {
		var pluginConfigSchema PluginConfigSchema
		expectedURL := "http://sevice-api.com/swagger.yaml"
		serviceConfigName := "test"
		expectedInscureSkipVerify := true
		services := map[string]*ServiceConfigV1{
			serviceConfigName: {
				SwaggerURL: expectedURL,
				//PluginVersion: expectedPluginVersion, Note: This functionality has been deprecated as of OpenAPI version > 2.2.0. Nonetheless, this test confirms backwards compatibility with configurations that are still specifying the plugin_version property
				InsecureSkipVerify: expectedInscureSkipVerify,
				SchemaConfigurationV1: []ServiceSchemaPropertyConfigurationV1{
					{
						SchemaPropertyName: "apikey_auth",
						DefaultValue:       "apiKeyValue",
						Command:            []string{"echo", "something"},
						CommandTimeout:     10,
						ExternalConfiguration: ServiceSchemaPropertyExternalConfigurationV1{
							File:        "some_file",
							KeyName:     "some_key_name",
							ContentType: "json",
						},
					},
				},
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(services)
		Convey("When Marshal method is called", func() {
			marshalConfig, err := pluginConfigSchema.Marshal()
			Convey("Then the marshalConfig should contain the right marshal configuration (and the plugin_version property should not be present) and the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
				expectedConfig := fmt.Sprintf(`version: "1"
services:
  test:
    swagger-url: %s
    insecure_skip_verify: %t
    schema_configuration:
    - schema_property_name: apikey_auth
      default_value: apiKeyValue
      cmd: [echo, something]
      cmd_timeout: 10
      schema_property_external_configuration:
        file: some_file
        key_name: some_key_name
        content_type: json
`, expectedURL, expectedInscureSkipVerify)
				So(string(marshalConfig), ShouldEqual, expectedConfig)
			})
		})
	})
	Convey("Given a PluginConfigSchemaV1 containing a version supported and a SchemaConfigurationV1 without a Command, CommandTimeout, or ExternalConfiguration", t, func() {
		var pluginConfigSchema PluginConfigSchema
		expectedURL := "http://sevice-api.com/swagger.yaml"
		serviceConfigName := "test"
		expectedInsecureSkipVerify := true
		services := map[string]*ServiceConfigV1{
			serviceConfigName: {
				SwaggerURL:         expectedURL,
				InsecureSkipVerify: expectedInsecureSkipVerify,
				SchemaConfigurationV1: []ServiceSchemaPropertyConfigurationV1{
					{
						SchemaPropertyName: "apikey_auth",
						DefaultValue:       "apiKeyValue",
					},
				},
			},
		}
		pluginConfigSchema = NewPluginConfigSchemaV1(services)
		Convey("When Marshal method is called", func() {
			marshalConfig, err := pluginConfigSchema.Marshal()
			Convey("Then the marshalConfig should contain the right marshal configuration (and the plugin_version property should not be present) and the error returned should be nil as configuration is correct", func() {
				So(err, ShouldBeNil)
				expectedConfig := fmt.Sprintf(`version: "1"
services:
  test:
    swagger-url: %s
    insecure_skip_verify: %t
    schema_configuration:
    - schema_property_name: apikey_auth
      default_value: apiKeyValue
`, expectedURL, expectedInsecureSkipVerify)
				So(string(marshalConfig), ShouldEqual, expectedConfig)
			})
		})
	})
}
