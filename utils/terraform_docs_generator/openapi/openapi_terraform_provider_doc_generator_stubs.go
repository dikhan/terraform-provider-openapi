package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

// specAnalyserStub is a stubbed spec analyser used for testing purposes that implements the SpecAnalyser interface
type specAnalyserStub struct {
	openapi.SpecAnalyser
	//resources            []openapi.SpecResource
	//dataSources          []openapi.SpecResource
	security *specSecurityStub
	//headers              openapi.SpecHeaderParameters
	backendConfiguration *specStubBackendConfiguration
	error                error
}

//func (s *specAnalyserStub) GetTerraformCompliantResources() ([]openapi.SpecResource, error) {
//	if s.error != nil {
//		return nil, s.error
//	}
//	return s.resources, nil
//}
//
//func (s *specAnalyserStub) GetTerraformCompliantDataSources() []openapi.SpecResource {
//	return s.dataSources
//}

func (s *specAnalyserStub) GetSecurity() openapi.SpecSecurity {
	if s.security != nil {
		return s.security
	}
	return nil
}

//func (s *specAnalyserStub) GetAllHeaderParameters() (openapi.SpecHeaderParameters, error) {
//	if s.headers != nil {
//		return s.headers, nil
//	}
//	return nil, nil
//}
//
func (s *specAnalyserStub) GetAPIBackendConfiguration() (openapi.SpecBackendConfiguration, error) {
	if s.error != nil {
		return nil, s.error
	}
	if s.backendConfiguration != nil {
		return s.backendConfiguration, nil
	}
	return nil, nil
}

// specSecurityStub
type specSecurityStub struct {
	openapi.SpecSecurity
	securityDefinitions   *openapi.SpecSecurityDefinitions
	globalSecuritySchemes openapi.SpecSecuritySchemes
	error                 error
}

func (s *specSecurityStub) GetAPIKeySecurityDefinitions() (*openapi.SpecSecurityDefinitions, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.securityDefinitions, nil
}

func (s *specSecurityStub) GetGlobalSecuritySchemes() (openapi.SpecSecuritySchemes, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.globalSecuritySchemes, nil
}

//specStubBackendConfiguration
type specStubBackendConfiguration struct {
	openapi.SpecBackendConfiguration
	host    string
	regions []string
	err     error
}

func (s *specStubBackendConfiguration) IsMultiRegion() (bool, string, []string, error) {
	if s.err != nil {
		return false, "", nil, s.err
	}
	if len(s.regions) > 0 {
		return true, s.host, s.regions, nil
	}
	return false, "", nil, nil
}

type specStubResource struct {
	openapi.SpecResource
	name             string
	shouldIgnore     bool
	schemaDefinition *openapi.SpecSchemaDefinition
	error            error
}

func (s *specStubResource) ShouldIgnoreResource() bool { return s.shouldIgnore }

func (s *specStubResource) GetResourceSchema() (*openapi.SpecSchemaDefinition, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.schemaDefinition, nil
}

func (s *specStubResource) GetResourceName() string { return s.name }

//specStubSecurityDefinition
type specStubSecurityDefinition struct {
	openapi.SpecSecurityDefinition
	name string
}

func (s specStubSecurityDefinition) GetTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.name)
}
