package openapi

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type telemetryProviderStub struct {
	validationError              error
	terraformVersionReceived     string
	openAPIPluginVersionReceived string
	providerNameReceived         string
	resourceNameReceived         string
	tfOperationReceived          TelemetryResourceOperation
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

func (t *telemetryProviderStub) IncServiceProviderResourceTotalRunsCounter(providerName, resourceName string, tfOperation TelemetryResourceOperation, telemetryProviderConfiguration TelemetryProviderConfiguration) error {
	t.providerNameReceived = providerName
	t.resourceNameReceived = resourceName
	t.tfOperationReceived = tfOperation
	return nil
}

func (t *telemetryProviderStub) GetTelemetryProviderConfiguration(data *schema.ResourceData) TelemetryProviderConfiguration {
	return t.telemetryProviderConfig
}
