package openapi

type telemetryProviderStub struct {
	validationError              error
	terraformVersionReceived     string
	openAPIPluginVersionReceived string
	providerNameReceived         string
}

func (t *telemetryProviderStub) Validate() error {
	return t.validationError
}

func (t *telemetryProviderStub) IncTerraformVersionTotalRunsCounter(terraformVersion string) error {
	t.terraformVersionReceived = terraformVersion
	return nil
}

func (t *telemetryProviderStub) IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string) error {
	t.openAPIPluginVersionReceived = openAPIPluginVersion
	return nil
}

func (t *telemetryProviderStub) IncServiceProviderTotalRunsCounter(providerName string) error {
	t.providerNameReceived = providerName
	return nil
}
