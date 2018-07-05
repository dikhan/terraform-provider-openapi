package openapi

import (
	"fmt"
	"github.com/asaskevich/govalidator"
)

// PluginConfigSchema defines the interface/expected behaviour for PluginConfigSchema implementations.
type PluginConfigSchema interface {
	Validate() error
	GetProviderURL(providerName string) (string, error)
}

// PluginConfigSchemaV1 defines PluginConfigSchema version 1
// Configuration example:
// version: '1'
// services:
//   monitor: http://monitor-api.com/swagger.json
//   cdn: https://cdn-api.com/swagger.json
//   vm: http://vm-api.com/swagger.json
type PluginConfigSchemaV1 struct {
	Version  string
	Services map[string]string
}

// NewPluginConfigSchemaV1 creates a new PluginConfigSchemaV1 that implements PluginConfigSchema interface
func NewPluginConfigSchemaV1(version string, services map[string]string) PluginConfigSchemaV1 {
	return PluginConfigSchemaV1{
		Version:  version,
		Services: services,
	}
}

// Validate makes sure that schema data is correct
func (p PluginConfigSchemaV1) Validate() error {
	if p.Version != "1" {
		return fmt.Errorf("provider configuration version not matching current implementation, please use version '1' of provider configuration specification")
	}
	for k, v := range p.Services {
		if !govalidator.IsURL(v) {
			return fmt.Errorf("service '%s' found in the provider configuration does not contain a valid URL value '%s'", k, v)
		}
	}
	return nil
}

// GetProviderURL returns the swagger URL for the given provider name
func (p PluginConfigSchemaV1) GetProviderURL(providerName string) (string, error) {
	if providerName == "" {
		return "", fmt.Errorf("providerName not specified")
	}
	providerURL := p.Services[providerName]
	if providerURL == "" {
		return "", fmt.Errorf("'%s' not found in provider's services configuration", providerName)
	}
	return providerURL, nil
}
