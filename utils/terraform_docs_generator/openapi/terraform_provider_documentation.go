package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/templates"
	"io"
	"text/template"
)

// TerraformProviderDocumentation defines the attributes needed to generate Terraform provider documentation
type TerraformProviderDocumentation struct {
	ProviderName          string
	ProviderInstallation  ProviderInstallation
	ProviderConfiguration ProviderConfiguration
	ProviderResources     ProviderResources
	DataSources           DataSources
}

// ProviderInstallation includes details needed to install the Terraform provider plugin
type ProviderInstallation struct {
	// Example code/commands for installing the provider
	Example string
	// Other instructions to install/run the provider
	Other string
	// Other code/commands needed to install/run the provider
	OtherCommand string
}

// ProviderConfiguration defines the details needed to properly configure the Terraform provider
type ProviderConfiguration struct {
	Regions            []string
	ConfigProperties   []Property
	ExampleUsage       []ExampleUsage
	ArgumentsReference ArgumentsReference
}

// ProviderResources defines the resources exposed by the Terraform provider
type ProviderResources struct {
	Resources []Resource
}

// DataSources defines the data sources and data source instances exposed by the Terraform provider
type DataSources struct {
	DataSources         []DataSource
	DataSourceInstances []DataSource
}

// DataSource defines the attributes to generate documentation for a Terraform provider data source
type DataSource struct {
	Name         string
	OtherExample string
	Properties   []Property
}

// Resource defines the attributes to generate documentation for a Terraform provider resource
type Resource struct {
	Name               string
	Description        string
	Properties         []Property
	ExampleUsage       []ExampleUsage
	ArgumentsReference ArgumentsReference
}

// ExampleUsage defines a block of code/commands to include in the docs
type ExampleUsage struct {
	Example string
}

// ArgumentsReference defines any notes that need to be appended to a resource's arguments reference section (eg: known issues)
type ArgumentsReference struct {
	Notes []string
}

// AttributesReference defines the attributes needed for a resource's attributes reference section
type AttributesReference struct {
	Description string
	Properties  []Property
	Notes       []string
}

// Property defines the attributes for describing a given property for a resource
type Property struct {
	Name           string
	Type           string
	ArrayItemsType string
	Required       bool
	Computed       bool
	Description    string
	Schema         []Property // This is used to describe the schema for array of objects or object properties
}

// RenderZendeskHTML renders the documentation in HTML using the Zendesk template templates.ZendeskTmpl
func (t TerraformProviderDocumentation) RenderZendeskHTML(w io.Writer) error {
	tmpl, err := template.New("TerraformProviderDocumentation").Parse(templates.ZendeskTmpl)
	if err != nil {
		return err
	}
	err = tmpl.Execute(w, t)
	if err != nil {
		return err
	}
	return nil
}
