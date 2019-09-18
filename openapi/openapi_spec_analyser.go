package openapi

import (
	"fmt"
)

// SpecAnalyser analyses the swagger doc and provides helper methods to retrieve all the end points that can
// be used as terraform resources. These endpoints have to meet certain criteria to be considered eligible resources
// as explained below:
// A resource is considered any end point that meets the following:
// 	- POST operation on the root path (e,g: api/users)
//	- GET operation on the instance path (e,g: api/users/{id}). Other operations like DELETE, PUT are optional
// In the example above, the resource name would be 'users'.
// Versioning is also supported, thus if the endpoint above had been api/v1/users the corresponding resouce name would
// have been 'users_v1'
type SpecAnalyser interface {
	// GetTerraformCompliantResources defines the method that is meant to discover the paths from the OpenAPI document
	// that are considered Terraform compliant, returning a list of SpecResource or an error otherwise.
	GetTerraformCompliantResources() ([]SpecResource, error)
	// GetTerraformCompliantDataSources is responsible for finding endpoints that are deemed terraform data source compatible
	// and returns a list of SpecResource configured as data sources
	GetTerraformCompliantDataSources() []SpecResource
	// GetSecurity returns a SpecSecurity based on the security defined in the OpenAPI document
	GetSecurity() SpecSecurity
	// GetAllHeaderParameters returns SpecHeaderParameters containing all the headers defined in the OpenAPI document. This
	// enabled the OpenAPI provider to expose the headers as configurable properties available in the OpenAPI Terraform
	// provider; so users can provide values for the headers that are meant to be sent along with the operations the headers
	// are defined in.
	GetAllHeaderParameters() (SpecHeaderParameters, error)
	// GetAPIBackendConfiguration encapsulates all the information related to the backend in the OpenAPI doc
	// (e,g: host, protocols, etc) which is then used in the ProviderClient to communicate with the API as specified in
	// the configuration.
	GetAPIBackendConfiguration() (SpecBackendConfiguration, error)
}

// SpecAnalyserVersion defines the type for versions supported in the SpecAnalyser
type SpecAnalyserVersion string

const (
	// specAnalyserV2 version that supports OpenAPI v2 (swagger)
	specAnalyserV2 SpecAnalyserVersion = "v2"
)

// CreateSpecAnalyser is a factory method that returns the appropriate implementation of SpecAnalyser
// depending upon the openApiSpecAnalyserVersion passed in.
// Currently only OpenAPI v2 version is supported but this constructor is ready to handle new implementations such as v3
// when the time comes
func CreateSpecAnalyser(specAnalyserVersion SpecAnalyserVersion, openAPIDocumentURL string) (SpecAnalyser, error) {
	var err error
	var specAnalyser SpecAnalyser
	switch specAnalyserVersion {
	case specAnalyserV2:
		specAnalyser, err = newSpecAnalyserV2(openAPIDocumentURL)
	default:
		return nil, fmt.Errorf("open api spec analyser version '%s' not supported, please choose a valid SpecAnalyser implementation [%s]", specAnalyserVersion, specAnalyserV2)
	}
	if err != nil {
		return nil, err
	}
	return specAnalyser, nil
}
