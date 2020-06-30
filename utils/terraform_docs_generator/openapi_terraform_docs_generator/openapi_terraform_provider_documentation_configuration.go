package openapi_terraform_docs_generator

import (
	"io"
)

// ProviderConfiguration defines the details needed to properly configure the Terraform provider
type ProviderConfiguration struct {
	// ProviderName is the name of the provider
	ProviderName       string
	Regions            []string
	ConfigProperties   []Property
	ExampleUsage       []ExampleUsage
	ArgumentsReference ArgumentsReference
}

// Render renders into the input writer the ProviderInstallation documentation formatted in HTML
func (t ProviderConfiguration) Render(w io.Writer, template string) error {
	return Render(w, "ProviderConfiguration", template, t)
}
