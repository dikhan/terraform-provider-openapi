package openapiterraformdocsgenerator

// DataSources defines the data sources and data source instances exposed by the Terraform provider
type DataSources struct {
	// ProviderName is the name of the provider
	ProviderName        string
	DataSources         []DataSource
	DataSourceInstances []DataSource
}

// DataSource defines the attributes to generate documentation for a Terraform provider data source
type DataSource struct {
	Name         string
	Description  string
	OtherExample string
	Properties   []Property
}
