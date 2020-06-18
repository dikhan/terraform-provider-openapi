package openapi

import (
	"io"
	"io/ioutil"
	"path/filepath"
	"text/template"
)

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
	ConfigProperties   []Property
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

func (t TerraformProviderDocumentation) RenderZendeskHTML(w io.Writer) error {
	absPath, _ := filepath.Abs("./utils/terraform_docs_generator/openapi/templates/zendesk_template.html")
	b, _ := ioutil.ReadFile(absPath)

	tmpl, err := template.New("TerraformProviderDocumentation").Parse(string(b))
	if err != nil {
		return err
	}
	err = tmpl.Execute(w, t)
	if err != nil {
		return err
	}
	return nil
}
