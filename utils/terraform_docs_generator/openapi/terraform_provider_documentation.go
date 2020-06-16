package openapi

type TerraformProviderDocumentation struct {
	ProviderName          string
	ProviderInstallation  ProviderInstallation
	ProviderConfiguration ProviderConfiguration
	ProviderResources     ProviderResources
	DataSources           DataSources
}

type ProviderInstallation struct {
	Example      string
	Other        string
	OtherCommand string
}

type ProviderConfiguration struct {
	Regions            []string
	ExampleUsage       []ExampleUsage
	ArgumentsReference ArgumentsReference
}

type ProviderResources struct {
	Resources []Resource
}

type DataSources struct {
	DataSources         []DataSource
	DataSourceInstances []DataSource
}

type DataSource struct {
	Name        string
	Description string
	Properties  []Property
}

type Resource struct {
	Name               string
	Description        string
	Properties         []Property
	Import             Import
	ArgumentsReference ArgumentsReference
}

type ExampleUsage struct {
	Example string
}

type ArgumentsReference struct {
	Notes []string
}

type AttributesReference struct {
	Description string
	Properties  []Property
	Notes       []string
}

type Import struct {
	Notes []string
}

type Property struct {
	Name           string
	Type           string
	ArrayItemsType string
	Required       bool
	Computed       bool
	Description    string
	Schema         []Property // This is used to describe the schema for array of objects or object properties
}

func (t TerraformProviderDocumentation) renderMarkup() {

}
