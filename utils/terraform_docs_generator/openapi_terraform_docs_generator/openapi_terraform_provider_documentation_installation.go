package openapi_terraform_docs_generator

// ProviderInstallation includes details needed to install the Terraform provider plugin
type ProviderInstallation struct {
	// ProviderName is the name of the provider
	ProviderName string
	// Example code/commands for installing the provider
	Example string
	// Other instructions to install/run the provider
	Other string
	// Other code/commands needed to install/run the provider
	OtherCommand string
}
