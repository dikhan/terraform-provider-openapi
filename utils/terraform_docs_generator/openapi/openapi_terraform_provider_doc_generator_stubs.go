package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

// specAnalyserStub is a stubbed spec analyser used for testing purposes that implements the SpecAnalyser interface
//type specAnalyserStub struct {
//	openapi.SpecAnalyser
//	resources            []openapi.SpecResource
//	dataSources          []openapi.SpecResource
//	security             *specSecurityStub
//	headers              openapi.SpecHeaderParameters
//	backendConfiguration openapi.SpecBackendConfiguration
//	error                error
//}
//
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
//
//func (s *specAnalyserStub) GetSecurity() openapi.SpecSecurity {
//	if s.security != nil {
//		return s.security
//	}
//	return nil
//}
//
//func (s *specAnalyserStub) GetAllHeaderParameters() (openapi.SpecHeaderParameters, error) {
//	if s.headers != nil {
//		return s.headers, nil
//	}
//	return nil, nil
//}
//
//func (s *specAnalyserStub) GetAPIBackendConfiguration() (openapi.SpecBackendConfiguration, error) {
//	if s.error != nil {
//		return nil, s.error
//	}
//	if s.backendConfiguration != nil {
//		return s.backendConfiguration, nil
//	}
//	return nil, nil
//}

// specStubResource is a stub implementation of SpecResource interface which is used for testing purposes
//type specStubResource struct {
//	openapi.SpecResource
//	name                    string
//	host                    string
//	path                    string
//	shouldIgnore            bool
//	schemaDefinition        *SpecSchemaDefinition
//	resourceGetOperation    *specResourceOperation
//	resourcePostOperation   *specResourceOperation
//	resourceListOperation   *specResourceOperation
//	resourcePutOperation    *specResourceOperation
//	resourceDeleteOperation *specResourceOperation
//	timeouts                *specTimeouts
//
//	parentResourceNames    []string
//	parentPropertyNames    []string
//	fullParentResourceName string
//
//	funcGetResourcePath   func(parentIDs []string) (string, error)
//	funcGetResourceSchema func() (*SpecSchemaDefinition, error)
//	error                 error
//}

// specSecurityStub
//type specSecurityStub struct {
//	openapi.SpecSecurity
//	securityDefinitions   *openapi.SpecSecurityDefinitions
//	globalSecuritySchemes openapi.SpecSecuritySchemes
//	error                 error
//}
//
//func (s *specSecurityStub) GetAPIKeySecurityDefinitions() (*openapi.SpecSecurityDefinitions, error) {
//	if s.error != nil {
//		return nil, s.error
//	}
//	return s.securityDefinitions, nil
//}
//
//func (s *specSecurityStub) GetGlobalSecuritySchemes() (openapi.SpecSecuritySchemes, error) {
//	if s.error != nil {
//		return nil, s.error
//	}
//	return s.globalSecuritySchemes, nil
//}

//specStubSecurityDefinition
type specStubSecurityDefinition struct {
	openapi.SpecSecurityDefinition
	name string
}

func (s specStubSecurityDefinition) GetTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.name)
}

//specStubBackendConfiguration
//type specStubBackendConfiguration struct {
//	openapi.SpecBackendConfiguration
//	host             string
//	regions          []string
//	err              error
//}
//
//func (s *specStubBackendConfiguration) IsMultiRegion() (bool, string, []string, error) {
//	if s.err != nil {
//		return false, "", nil, s.err
//	}
//	if len(s.regions) > 0 {
//		return true, s.host, s.regions, nil
//	}
//	return false, "", nil, nil
//}
