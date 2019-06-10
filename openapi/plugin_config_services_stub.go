package openapi

type serviceConfigStub struct {
	swaggerURL          string
	pluginVersion       string
	insecureSkipVerify  bool
	schemaConfiguration []*serviceSchemaPropertyConfigurationStub
	err                 error
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

func (s *serviceConfigStub) GetPluginVersion() string {
	return s.pluginVersion
}

func (s *serviceConfigStub) IsInsecureSkipVerifyEnabled() bool {
	return s.insecureSkipVerify
}

func (s *serviceConfigStub) Validate(runningPluginVersion string) error {
	return s.err
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
