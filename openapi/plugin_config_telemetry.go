package openapi

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type TelemetryProviderConfiguration map[string]interface{}

// TelemetryProvider holds the behaviour expected to be implemented for the Telemetry Providers supported. At the moment
// only Graphite is supported.
type TelemetryProvider interface {
	// Validate performs a check to confirm that the telemetry configuration is valid
	Validate() error
	// IncOpenAPIPluginVersionTotalRunsCounter is the method responsible for submitting to the corresponding telemetry platform the counter increase for the OpenAPI plugin Version used
	IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string, telemetryProviderConfiguration TelemetryProviderConfiguration) error
	// IncServiceProviderTotalRunsCounter is the method responsible for submitting to the corresponding telemetry platform the counter increase for the service provider used
	IncServiceProviderTotalRunsCounter(providerName string, telemetryProviderConfiguration TelemetryProviderConfiguration) error
	// GetTelemetryProviderConfiguration is the method responsible for getting a specific telemetry provider config given the input data provided
	GetTelemetryProviderConfiguration(data *schema.ResourceData) TelemetryProviderConfiguration
}
