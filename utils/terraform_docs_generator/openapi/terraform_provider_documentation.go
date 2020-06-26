package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/templates/zendesk"
	"io"
)

// TerraformProviderDocumentation defines the attributes needed to generate Terraform provider documentation
type TerraformProviderDocumentation struct {
	ProviderName                string
	ProviderInstallation        ProviderInstallation
	ProviderConfiguration       ProviderConfiguration
	ProviderResources           ProviderResources
	DataSources                 DataSources
	ShowSpecialTermsDefinitions bool
}

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

// RenderZendesk renders into the input writer the ProviderInstallation documentation formatted in HTML
func (t ProviderInstallation) RenderZendesk(w io.Writer) error {
	return Render(w, "ProviderInstallation", zendesk.ProviderInstallationTmpl, t)
}

// ProviderConfiguration defines the details needed to properly configure the Terraform provider
type ProviderConfiguration struct {
	// ProviderName is the name of the provider
	ProviderName       string
	Regions            []string
	ConfigProperties   []Property
	ExampleUsage       []ExampleUsage
	ArgumentsReference ArgumentsReference
}

// RenderZendesk renders into the input writer the ProviderInstallation documentation formatted in HTML
func (t ProviderConfiguration) RenderZendesk(w io.Writer) error {
	return Render(w, "ProviderConfiguration", zendesk.ProviderConfigurationTmpl, t)
}

// ProviderResources defines the resources exposed by the Terraform provider
type ProviderResources struct {
	// ProviderName is the name of the provider
	ProviderName string
	Resources    []Resource
}

// RenderZendesk renders into the input writer the ProviderResources documentation formatted in HTML
func (t ProviderResources) RenderZendesk(w io.Writer) error {
	return Render(w, "ProviderResources", zendesk.ProviderResourcesTmpl, t)
}

func (r ProviderResources) ContainsResourcesWithSecretProperties() bool {
	for _, resource := range r.Resources {
		for _, prop := range resource.Properties {
			if prop.IsSensitive {
				return true
			}
		}
	}
	return false
}

// DataSources defines the data sources and data source instances exposed by the Terraform provider
type DataSources struct {
	// ProviderName is the name of the provider
	ProviderName        string
	DataSources         []DataSource
	DataSourceInstances []DataSource
}

// RenderZendesk renders into the input writer the DataSources documentation formatted in HTML
func (t DataSources) RenderZendesk(w io.Writer) error {
	return Render(w, "DataSources", zendesk.DataSourcesTmpl, t)
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
	ParentProperties   []string
	ExampleUsage       []ExampleUsage
	ArgumentsReference ArgumentsReference
}

func (r Resource) BuildImportIDsExample() string {
	if r.ParentProperties == nil {
		return "id"
	}
	idExamples := ""
	for _, prop := range r.ParentProperties {
		idExamples += prop + "/"
	}
	// Append the actual resource instance id
	if idExamples != "" {
		idExamples += r.Name + "_id"
	}
	return idExamples
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
	Name               string
	Type               string
	ArrayItemsType     string
	Required           bool
	Computed           bool
	IsOptionalComputed bool
	IsSensitive        bool
	Description        string
	Schema             []Property // This is used to describe the schema for array of objects or object properties
}

// ContainsComputedSubProperties checks if a schema contains properties that are computed recursively
func (p Property) ContainsComputedSubProperties() bool {
	for _, s := range p.Schema {
		if s.Computed || s.ContainsComputedSubProperties() {
			return true
		}
	}
	return false
}

// RenderZendeskHTML renders the documentation in HTML
func (t TerraformProviderDocumentation) RenderZendeskHTML(w io.Writer) error {
	err := Render(w, "TerraformProviderDocTableOfContents", zendesk.TableOfContentsTmpl, t)
	if err != nil {
		return err
	}
	t.ProviderInstallation.ProviderName = t.ProviderName
	err = t.ProviderInstallation.RenderZendesk(w)
	if err != nil {
		return err
	}
	t.ProviderConfiguration.ProviderName = t.ProviderName
	err = t.ProviderConfiguration.RenderZendesk(w)
	if err != nil {
		return err
	}
	t.ProviderResources.ProviderName = t.ProviderName
	err = t.ProviderResources.RenderZendesk(w)
	if err != nil {
		return err
	}
	t.DataSources.ProviderName = t.ProviderName
	err = t.DataSources.RenderZendesk(w)
	if err != nil {
		return err
	}
	err = Render(w, "TerraformProviderDocSpecialTermsDefinitions", zendesk.SpecialTermsTmpl, t)
	if err != nil {
		return err
	}
	return nil
}
