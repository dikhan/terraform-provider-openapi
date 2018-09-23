package openapi

// specAnalyserStub is a stubbed spec analyser used for testing purposes that implements the SpecAnalyser interface
type specAnalyserStub struct {
	resources            []SpecResource
	security             SpecSecurity
	headers              SpecHeaderParameters
	backendConfiguration SpecBackendConfiguration
	error                error
}

func (s *specAnalyserStub) GetTerraformCompliantResources() ([]SpecResource, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.resources, nil
}

func (s *specAnalyserStub) GetSecurity() SpecSecurity {
	return s.security
}

func (s *specAnalyserStub) GetAllHeaderParameters() SpecHeaderParameters {
	return s.headers
}

func (s *specAnalyserStub) GetOpenAPIBackendConfiguration() (SpecBackendConfiguration, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.backendConfiguration, nil
}
