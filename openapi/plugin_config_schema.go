package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/version"
	"gopkg.in/yaml.v2"
	"log"
)

// ServiceConfigurations contains the map with all service configurations
type ServiceConfigurations map[string]ServiceConfiguration

// PluginConfigSchema defines the interface/expected behaviour for PluginConfigSchema implementations.
type PluginConfigSchema interface {
	// Validate performs a check to confirm that the schema content is correct
	Validate() error
	// GetServiceConfig returns the service configuration for a given provider name
	GetServiceConfig(providerName string) (ServiceConfiguration, error)
	// GetAllServiceConfigurations returns all the service configuration
	GetAllServiceConfigurations() (ServiceConfigurations, error)
	// GetVersion returns the plugin configuration version
	GetVersion() (string, error)
	// Marshal serializes the value provided into a YAML document
	Marshal() ([]byte, error)
	// GetTelemetryHandler returns a handler containing validated telemetry providers
	GetTelemetryHandler(providerName string) TelemetryHandler
}

// PluginConfigSchemaV1 defines PluginConfigSchema version 1
// Configuration example:
// version: '1'
// services:
//   monitor:
//     swagger-url: http://monitor-api.com/swagger.json
//     plugin_version: 0.14.0
//     insecure_skip_verify: true
//   cdn:
//     swagger-url: https://cdn-api.com/swagger.json
//   vm:
//     swagger-url: http://vm-api.com/swagger.json
type PluginConfigSchemaV1 struct {
	Version         string                      `yaml:"version"`
	TelemetryConfig *TelemetryConfig            `yaml:"telemetry,omitempty"`
	Services        map[string]*ServiceConfigV1 `yaml:"services"`
}

// TelemetryConfig contains the configuration for the telemetry
type TelemetryConfig struct {
	// Graphite defines the configuration needed to ship telemetry to Graphite
	Graphite *TelemetryProviderGraphite `yaml:"graphite,omitempty"`
	// HttpEndpoint defines the configuration needed to ship telemetry to an http endpoint
	HttpEndpoint *TelemetryProviderHttpEndpoint `yaml:"http_endpoint,omitempty"`
}

// NewPluginConfigSchemaV1 creates a new PluginConfigSchemaV1 that implements PluginConfigSchema interface
func NewPluginConfigSchemaV1(services map[string]*ServiceConfigV1, telemetryConfig *TelemetryConfig) *PluginConfigSchemaV1 {
	return &PluginConfigSchemaV1{
		Version:         "1",
		Services:        services,
		TelemetryConfig: telemetryConfig,
	}
}

// Validate makes sure that schema data is correct
func (p *PluginConfigSchemaV1) Validate() error {
	if p.Version != "1" {
		return fmt.Errorf("provider configuration version not matching current implementation, please use version '1' of provider configuration specification")
	}
	return nil
}

// GetServiceConfig returns the configuration for the given provider name
func (p *PluginConfigSchemaV1) GetServiceConfig(providerName string) (ServiceConfiguration, error) {
	if providerName == "" {
		return nil, fmt.Errorf("providerName not specified")
	}
	serviceConfig, exists := p.Services[providerName]
	if !exists {
		return nil, fmt.Errorf("'%s' not found in provider's services configuration", providerName)
	}
	return serviceConfig, nil
}

// GetVersion returns the plugin configuration version
func (p *PluginConfigSchemaV1) GetVersion() (string, error) {
	return p.Version, nil
}

// GetAllServiceConfigurations returns all the service configuration
func (p *PluginConfigSchemaV1) GetAllServiceConfigurations() (ServiceConfigurations, error) {
	serviceConfigurations := ServiceConfigurations{}
	for k, v := range p.Services {
		serviceConfigurations[k] = v
	}
	return serviceConfigurations, nil
}

// Marshal serializes the value provided into a YAML document
func (p *PluginConfigSchemaV1) Marshal() ([]byte, error) {
	out, err := yaml.Marshal(p)
	return out, err
}

// GetTelemetryHandler returns a handler containing validated telemetry providers
func (p *PluginConfigSchemaV1) GetTelemetryHandler(providerName string) TelemetryHandler {
	var telemetryProviders []TelemetryProvider
	if p.TelemetryConfig != nil {
		if p.TelemetryConfig.Graphite != nil {
			err := p.TelemetryConfig.Graphite.Validate()
			if err != nil {
				log.Printf("[WARN] ignoring graphite telemetry due to the following validation error: %s", err)
			} else {
				telemetryProviders = append(telemetryProviders, p.TelemetryConfig.Graphite)
				log.Printf("[DEBUG] graphite telemetry provider enabled")
			}
		} else {
			log.Printf("[DEBUG] graphite telemetry configuration not present")
		}

		if p.TelemetryConfig.HttpEndpoint != nil {
			err := p.TelemetryConfig.HttpEndpoint.Validate()
			if err != nil {
				log.Printf("[WARN] ignoring http endpoint telemetry due to the following validation error: %s", err)
			} else {
				telemetryProviders = append(telemetryProviders, p.TelemetryConfig.HttpEndpoint)
				log.Printf("[DEBUG] http endpoint telemetry provider enabled")
			}
		} else {
			log.Printf("[DEBUG] http endpoint telemetry configuration not present")
		}
	}

	if len(telemetryProviders) == 0 {
		log.Printf("[DEBUG] telemetry not configured")
		return nil
	}

	return telemetryHandlerTimeoutSupport{
		timeout:            telemetryTimeout,
		providerName:       providerName,
		openAPIVersion:     version.Version,
		telemetryProviders: telemetryProviders,
	}
}
