package openapi

import (
	"io"
)

// DataSources defines the data sources and data source instances exposed by the Terraform provider
type DataSources struct {
	// ProviderName is the name of the provider
	ProviderName        string
	DataSources         []DataSource
	DataSourceInstances []DataSource
}

// Render renders into the input writer the DataSources documentation formatted in HTML
func (t DataSources) Render(w io.Writer, template string) error {
	return Render(w, "DataSources", template, t)
}

// DataSource defines the attributes to generate documentation for a Terraform provider data source
type DataSource struct {
	Name         string
	OtherExample string
	Properties   []Property
}
