package openapi

type serviceConfigStub struct {
	swaggerURL          string
	insecureSkipVerify  bool
	schemaConfiguration []*serviceSchemaPropertyConfigurationStub
}

type serviceSchemaPropertyConfigurationStub struct {
	schemaPropertyName   string
	defaultValue         string
	err                  error
	executeCommandCalled bool
}

func (s *serviceConfigStub) GetSwaggerURL() string {
	return s.swaggerURL
}

func (s *serviceConfigStub) IsInsecureSkipVerifyEnabled() bool {
	return s.insecureSkipVerify
}

func (s serviceConfigStub) GetSchemaPropertyConfiguration(schemaPropertyName string) ServiceSchemaPropertyConfiguration {
	for _, p := range s.schemaConfiguration {
		if p.schemaPropertyName == schemaPropertyName {
			return p
		}
	}
	return nil
}

func (s *serviceSchemaPropertyConfigurationStub) GetDefaultValue() (string, error) {
	return s.defaultValue, nil
}

func (s *serviceSchemaPropertyConfigurationStub) ExecuteCommand() error {
	s.executeCommandCalled = true
	return s.err
}
