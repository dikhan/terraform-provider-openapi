package openapi

// ServiceConfigStub implements the ServiceConfiguration interface and can be used to simplify the creation of the ProviderOpenAPI
// provider by calling the CreateSchemaProviderWithConfiguration function passing in the stub wit the swagger URL populated
// with the URL where the openapi doc is hosted.
type ServiceConfigStub struct {
	SwaggerURL          string
	PluginVersion       string
	InsecureSkipVerify  bool
	SchemaConfiguration []*ServiceSchemaPropertyConfigurationStub
	Err                 error
}

// ServiceSchemaPropertyConfigurationStub implements the ServiceSchemaPropertyConfiguration and can be used to simplify
// tests that require ServiceSchemaPropertyConfiguration implemenations
type ServiceSchemaPropertyConfigurationStub struct {
	SchemaPropertyName   string
	DefaultValue         string
	Err                  error
	ExecuteCommandCalled bool
}

// GetSwaggerURL returns the swagger URL value configured in the ServiceConfigStub.SwaggerURL field
func (s *ServiceConfigStub) GetSwaggerURL() string {
	return s.SwaggerURL
}

// GetPluginVersion returns the plugin version value configured in the ServiceConfigStub.PluginVersion field
func (s *ServiceConfigStub) GetPluginVersion() string {
	return s.PluginVersion
}

// IsInsecureSkipVerifyEnabled returns the bool configured in the ServiceConfigStub.InsecureSkipVerify field
func (s *ServiceConfigStub) IsInsecureSkipVerifyEnabled() bool {
	return s.InsecureSkipVerify
}

// Validate returns an error if the ServiceConfigStub.Err field is set with an error
func (s *ServiceConfigStub) Validate(runningPluginVersion string) error {
	return s.Err
}

// GetSchemaPropertyConfiguration returns the service schema configuration set in the ServiceConfigStub.SchemaConfiguration field
func (s ServiceConfigStub) GetSchemaPropertyConfiguration(schemaPropertyName string) ServiceSchemaPropertyConfiguration {
	for _, p := range s.SchemaConfiguration {
		if p.SchemaPropertyName == schemaPropertyName {
			return p
		}
	}
	return nil
}

// GetDefaultValue returns the dafult value configured in the ServiceSchemaPropertyConfigurationStub.defaultValue field
func (s *ServiceSchemaPropertyConfigurationStub) GetDefaultValue() (string, error) {
	return s.DefaultValue, nil
}

// ExecuteCommand keeps track if the execute command method has been called and returns the configured err
// ServiceSchemaPropertyConfigurationStub.ServiceSchemaPropertyConfigurationStub if set
func (s *ServiceSchemaPropertyConfigurationStub) ExecuteCommand() error {
	s.ExecuteCommandCalled = true
	return s.Err
}
