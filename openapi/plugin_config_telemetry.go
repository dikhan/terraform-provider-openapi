package openapi

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

// TelemetryProviderConfiguration defines the struct type that specific telemetry providers can configure based on the
// resource data received in GetTelemetryProviderConfiguration. The struct serves as a way to document in the metric
// methods signature (eg: IncOpenAPIPluginVersionTotalRunsCounter) that a specific telemetry provider configuration struct
// can be passed in if needed
type TelemetryProviderConfiguration interface{}

// TelemetryResourceOperation defines the resource operation (CRUD) to be used in the telemetry metric
type TelemetryResourceOperation string

const (
	// TelemetryResourceOperationCreate represents the create operation invocation
	TelemetryResourceOperationCreate TelemetryResourceOperation = "create"
	// TelemetryResourceOperationRead represents the read operation invocation
	TelemetryResourceOperationRead TelemetryResourceOperation = "read"
	// TelemetryResourceOperationUpdate represents the update operation invocation
	TelemetryResourceOperationUpdate TelemetryResourceOperation = "update"
	// TelemetryResourceOperationDelete represents the delete operation invocation
	TelemetryResourceOperationDelete TelemetryResourceOperation = "delete"
	// TelemetryResourceOperationImport represents the import operation invocation
	TelemetryResourceOperationImport TelemetryResourceOperation = "import"
)

// TelemetryProvider holds the behaviour expected to be implemented for the Telemetry Providers supported. At the moment
// only Graphite is supported.
type TelemetryProvider interface {
	// Validate performs a check to confirm that the telemetry configuration is valid
	Validate() error
	// IncOpenAPIPluginVersionTotalRunsCounter is the method responsible for submitting to the corresponding telemetry platform the counter increase for the OpenAPI plugin Version used
	IncOpenAPIPluginVersionTotalRunsCounter(openAPIPluginVersion string, telemetryProviderConfiguration TelemetryProviderConfiguration) error
	// IncServiceProviderResourceTotalRunsCounter is the method responsible for submitting to the corresponding telemetry platform the counter increase for service provider used along
	// with tags for provider name, resource name, and Terraform operation
	IncServiceProviderResourceTotalRunsCounter(providerName, resourceName string, tfOperation TelemetryResourceOperation, telemetryProviderConfiguration TelemetryProviderConfiguration) error
	// GetTelemetryProviderConfiguration is the method responsible for getting a specific telemetry provider config given the input data provided
	GetTelemetryProviderConfiguration(data *schema.ResourceData) TelemetryProviderConfiguration
}
