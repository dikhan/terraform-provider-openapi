package openapi

// ServiceConfigStub implements the ServiceConfiguration interface and can be used to simplify the creation of the ProviderOpenAPI
// provider by calling the CreateSchemaProviderWithConfiguration function passing in the stub wit the swagger URL populated
// with the URL where the openapi doc is hosted.
type ServiceConfigStub struct {
	SwaggerURL          string
	PluginVersion       string
	InsecureSkipVerify  bool
	Telemetry           TelemetryProvider
	SchemaConfiguration []*ServiceSchemaPropertyConfigurationStub
	Err                 error
}

// ServiceSchemaPropertyConfigurationStub implements the ServiceSchemaPropertyConfiguration and can be used to simplify
// tests that require ServiceSchemaPropertyConfiguration implemenations
type ServiceSchemaPropertyConfigurationStub struct {
	SchemaPropertyName   string
	DefaultValue         string
	Err                  error
	GetDefaultValueFunc  func() (string, error)
	ExecuteCommandCalled bool
}

// GetSwaggerURL returns the swagger URL value configured in the ServiceConfigStub.SwaggerURL field
func (s *ServiceConfigStub) GetSwaggerURL() string {
	return s.SwaggerURL
}

// IsInsecureSkipVerifyEnabled returns the bool configured in the ServiceConfigStub.InsecureSkipVerify field
func (s *ServiceConfigStub) IsInsecureSkipVerifyEnabled() bool {
	return s.InsecureSkipVerify
}

// Validate returns an error if the ServiceConfigStub.Err field is set with an error
func (s *ServiceConfigStub) Validate() error {
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

// GetTelemetryConfiguration returns the TelemetryProvider configured in the ServiceConfigStub
func (s ServiceConfigStub) GetTelemetryConfiguration() TelemetryProvider {
	return s.Telemetry
}

// GetDefaultValue returns the default value configured in the ServiceSchemaPropertyConfigurationStub.defaultValue field
func (s *ServiceSchemaPropertyConfigurationStub) GetDefaultValue() (string, error) {
	if s.GetDefaultValueFunc != nil {
		return s.GetDefaultValueFunc()
	}
	return s.DefaultValue, nil
}

// ExecuteCommand keeps track if the execute command method has been called and returns the configured err
// ServiceSchemaPropertyConfigurationStub.ServiceSchemaPropertyConfigurationStub if set
func (s *ServiceSchemaPropertyConfigurationStub) ExecuteCommand() error {
	s.ExecuteCommandCalled = true
	return s.Err
}
