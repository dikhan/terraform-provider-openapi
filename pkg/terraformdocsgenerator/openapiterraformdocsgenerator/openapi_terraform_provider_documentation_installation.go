package openapiterraformdocsgenerator

// ProviderInstallation includes details needed to install the Terraform provider plugin
type ProviderInstallation struct {
	// ProviderName is the name of the provider
	ProviderName string
	// Hostname the Terraform registry that distributes the provider as documented in https://www.terraform.io/docs/language/providers/requirements.html#source-addresses
	// For in-house providers that you intend to distribute from a local filesystem directory, you can use an arbitrary hostname in a domain your organization controls. For example, if your corporate domain were example.com then you might choose
	// to use terraform.example.com as your placeholder hostname, even if that hostname doesn't actually resolve in DNS.
	Hostname string
	// Namespace An organizational namespace within the specified registry to be used for configuration purposes as documented in https://www.terraform.io/docs/language/providers/requirements.html#source-addresses
	Namespace string
	// PluginVersionConstraint should contain the OpenAPI plugin version constraint eg: "~> 2.1.0". If not populated the renderer
	// will default to ">= 2.1.0" OpenAPI provider version
	PluginVersionConstraint string
	// Example code/commands for installing the provider
	Example string
	// Other instructions to install/run the provider
	Other string
	// Other code/commands needed to install/run the provider
	OtherCommand string
}
