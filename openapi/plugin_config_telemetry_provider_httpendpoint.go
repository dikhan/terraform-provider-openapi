package openapi

// TelemetryProviderHttpEndpoint defines the configuration for HttpEndpoint. This struct also implements the TelemetryProvider interface
// and ships metrics to the following namespace by default <prefix>.terraform.* where '<prefix>' can be configured.
type TelemetryProviderHttpEndpoint struct {
	// URL describes the HTTP endpoint to send the metric to
	URL string `yaml:"url"`
	// Prefix enables to append a prefix to the metrics pushed to graphite
	Prefix string `yaml:"prefix,omitempty"`
}

// Validate checks whether the provider is configured correctly. This validation is performed upon telemetry provider registration. If this
// method returns an error the error will be logged but the telemetry will be disabled. Otherwise, the telemetry will be enabled
// and the corresponding metrics will be shipped to Graphite
func (g TelemetryProviderHttpEndpoint) Validate() error {

	return nil
}

// IncOpenAPIPluginVersionTotalRunsCounter will submit an increment to 1 the metric type counter '<prefix>.terraform.openapi_plugin_version.%s.total_runs'. The
// %s will be replaced by the OpenAPI plugin version used at runtime
func (g TelemetryProviderHttpEndpoint) IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string) error {

	return nil
}

// IncServiceProviderTotalRunsCounter will submit an increment to 1 the metric type counter '<prefix>.terraform.providers.%s.total_runs'. The
// %s will be replaced by the provider name used at runtime
func (g TelemetryProviderHttpEndpoint) IncServiceProviderTotalRunsCounter(providerName string) error {

	return nil
}
