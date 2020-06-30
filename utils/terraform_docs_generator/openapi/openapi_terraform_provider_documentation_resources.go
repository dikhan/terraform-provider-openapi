package openapi

import (
	"io"
)

// ProviderResources defines the resources exposed by the Terraform provider
type ProviderResources struct {
	// ProviderName is the name of the provider
	ProviderName string
	Resources    []Resource
}

// Render renders into the input writer the ProviderResources documentation formatted in HTML
func (t ProviderResources) Render(w io.Writer, template string) error {
	return Render(w, "ProviderResources", template, t)
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
