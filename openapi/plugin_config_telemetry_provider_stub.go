package openapi

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type telemetryProviderStub struct {
	validationError              error
	terraformVersionReceived     string
	openAPIPluginVersionReceived string
	providerNameReceived         string
	telemetryProviderConfig      TelemetryProviderConfiguration
}

func (t *telemetryProviderStub) Validate() error {
	return t.validationError
}

func (t *telemetryProviderStub) IncTerraformVersionTotalRunsCounter(terraformVersion string) error {
	t.terraformVersionReceived = terraformVersion
	return nil
}

func (t *telemetryProviderStub) IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	t.openAPIPluginVersionReceived = openAPIPluginVersion
	return nil
}

func (t *telemetryProviderStub) IncServiceProviderTotalRunsCounter(providerName string, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	t.providerNameReceived = providerName
	return nil
}

func (t *telemetryProviderStub) GetTelemetryProviderConfiguration(data *schema.ResourceData) TelemetryProviderConfiguration {
	return t.telemetryProviderConfig
}
