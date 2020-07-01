package openapi_terraform_docs_generator

// ProviderConfiguration defines the details needed to properly configure the Terraform provider
type ProviderConfiguration struct {
	// ProviderName is the name of the provider
	ProviderName       string
	Regions            []string
	ConfigProperties   []Property
	ExampleUsage       []ExampleUsage
	ArgumentsReference ArgumentsReference
}
