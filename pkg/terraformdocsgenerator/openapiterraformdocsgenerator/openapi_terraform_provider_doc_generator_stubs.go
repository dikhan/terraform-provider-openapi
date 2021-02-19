package openapiterraformdocsgenerator

import (
	"github.com/dikhan/terraform-provider-openapi/v2/openapi"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi/terraformutils"
)

// specAnalyserStub is a stubbed spec analyser used for testing purposes that implements the SpecAnalyser interface
type specAnalyserStub struct {
	openapi.SpecAnalyser
	resources            func() ([]openapi.SpecResource, error)
	dataSources          func() []openapi.SpecResource
	security             *specSecurityStub
	headers              openapi.SpecHeaderParameters
	backendConfiguration func() (*specStubBackendConfiguration, error)
	error                error
}

func (s *specAnalyserStub) GetTerraformCompliantResources() ([]openapi.SpecResource, error) {
	if s.resources != nil {
		return s.resources()
	}
	return nil, nil
}

func (s *specAnalyserStub) GetTerraformCompliantDataSources() []openapi.SpecResource {
	if s.dataSources != nil {
		return s.dataSources()
	}
	return nil
}

func (s *specAnalyserStub) GetSecurity() openapi.SpecSecurity {
	if s.security != nil {
		return s.security
	}
	return nil
}

func (s *specAnalyserStub) GetAllHeaderParameters() openapi.SpecHeaderParameters {
	if s.headers != nil {
		return s.headers
	}
	return nil
}

func (s *specAnalyserStub) GetAPIBackendConfiguration() (openapi.SpecBackendConfiguration, error) {
	if s.backendConfiguration != nil {
		return s.backendConfiguration()
	}
	return nil, nil
}

// specSecurityStub
type specSecurityStub struct {
	openapi.SpecSecurity
	securityDefinitions   func() (*openapi.SpecSecurityDefinitions, error)
	globalSecuritySchemes func() (openapi.SpecSecuritySchemes, error)
	error                 error
}

func (s *specSecurityStub) GetAPIKeySecurityDefinitions() (*openapi.SpecSecurityDefinitions, error) {
	return s.securityDefinitions()
}

func (s *specSecurityStub) GetGlobalSecuritySchemes() (openapi.SpecSecuritySchemes, error) {
	return s.globalSecuritySchemes()
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

// specStubResource
type specStubResource struct {
	openapi.SpecResource
	name                string
	shouldIgnore        bool
	schemaDefinition    *openapi.SpecSchemaDefinition
	parentResourceNames []string
	error               error
}

func (s *specStubResource) ShouldIgnoreResource() bool { return s.shouldIgnore }

func (s *specStubResource) GetResourceSchema() (*openapi.SpecSchemaDefinition, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.schemaDefinition, nil
}

func (s *specStubResource) GetResourceName() string { return s.name }

func (s *specStubResource) GetParentResourceInfo() *openapi.ParentResourceInfo {
	if len(s.parentResourceNames) > 0 {
		subRes := openapi.ParentResourceInfo{}
		subRes.SetParentResourceNames(s.parentResourceNames)
		return &subRes
	}
	return nil
}

//specStubSecurityDefinition
type specStubSecurityDefinition struct {
	openapi.SpecSecurityDefinition
	name string
}

func (s specStubSecurityDefinition) GetTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.name)
}
