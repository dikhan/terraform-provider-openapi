package openapi

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"os"
)

// ServiceConfiguration defines the interface/expected behaviour for ServiceConfiguration implementations.
type ServiceConfiguration interface {
	// GetSwaggerURL returns the URL where the service swagger doc is exposed
	GetSwaggerURL() string
	// GetSPluginVersion returns the OpenAPI Plugin version
	GetPluginVersion() string
	// IsInsecureSkipVerifyEnabled returns true if the given provider's service configuration has InsecureSkipVerify enabled; false
	// otherwise
	IsInsecureSkipVerifyEnabled() bool
	// GetSchemaPropertyConfiguration returns the schema configuration for the given schemaPropertyName
	GetSchemaPropertyConfiguration(schemaPropertyName string) ServiceSchemaPropertyConfiguration
	// Validate makes sure the configuration is valid
	Validate(runningPluginVersion string) error
}

// ServiceConfigV1 defines configuration for the service provider
type ServiceConfigV1 struct {
	// SwaggerURL defines where the swagger is located
	SwaggerURL string `yaml:"swagger-url"`
	// PluginVersion defines the version of the OpenAPI Terraform plugin installed when generating the plugin configuration
	PluginVersion string `yaml:"plugin_version,omitempty"`
	// InsecureSkipVerify defines whether the internal http client used to fetch the swagger file should verify the server cert
	// or not. This should only be used purposefully if the server is using a self-signed cert and only if the server is trusted
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
	// SchemaConfigurationV1 represents the list of schema property configurations
	SchemaConfigurationV1 []ServiceSchemaPropertyConfigurationV1 `yaml:"schema_configuration"`
}

// NewServiceConfigV1 creates a new instance of NewServiceConfigV1 struct with the values provided
func NewServiceConfigV1(swaggerURL string, insecureSkipVerifyEnabled bool) *ServiceConfigV1 {
	return &ServiceConfigV1{
		SwaggerURL:            swaggerURL,
		InsecureSkipVerify:    insecureSkipVerifyEnabled,
		SchemaConfigurationV1: []ServiceSchemaPropertyConfigurationV1{},
	}
}

// GetSwaggerURL returns the URL where the service swagger doc is exposed
func (s *ServiceConfigV1) GetSwaggerURL() string {
	return s.SwaggerURL
}

// GetPluginVersion returns the OpenAPI Plugin version
func (s *ServiceConfigV1) GetPluginVersion() string {
	return s.PluginVersion
}

// IsInsecureSkipVerifyEnabled returns true if the given provider's service configuration has InsecureSkipVerify enabled; false
// otherwise
func (s *ServiceConfigV1) IsInsecureSkipVerifyEnabled() bool {
	return s.InsecureSkipVerify
}

// GetSchemaPropertyConfiguration returns the external configuration for the given schema property name; nil is returned
// if no such property exists
func (s *ServiceConfigV1) GetSchemaPropertyConfiguration(schemaPropertyName string) ServiceSchemaPropertyConfiguration {
	for _, schemaPropertyConfig := range s.SchemaConfigurationV1 {
		if schemaPropertyConfig.SchemaPropertyName == schemaPropertyName {
			return schemaPropertyConfig
		}
	}
	return nil
}

// Validate makes sure the configuration is valid:
// - if the user has specified an OpenAPI plugin version, and if the plugin does not match the version then something is off
func (s *ServiceConfigV1) Validate(runningPluginVersion string) error {
	if !govalidator.IsURL(s.SwaggerURL) {
		// fall back to try to load the swagger file from disk in case the path provided is a path to a file on disk
		if _, err := os.Stat(s.SwaggerURL); os.IsNotExist(err) {
			return fmt.Errorf("service swagger URL configuration not valid ('%s'). URL must be either a valid formed URL or a path to an existing swagger file stored in the disk", s.SwaggerURL)
		}
	}
	if s.PluginVersion != "" {
		if s.PluginVersion != runningPluginVersion {
			return fmt.Errorf("plugin version '%s' in the plugin configuration file does not match the version of the OpenAPI plugin that is running '%s'", s.PluginVersion, runningPluginVersion)
		}
	}

	return nil
}
