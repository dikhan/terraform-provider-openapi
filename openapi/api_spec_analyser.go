package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiv2"
	"github.com/go-openapi/loads"
)

type OpenAPISecurity interface {
	GetApiKeySecurityDefinitions()
	GetGlobalSecuritySchemes() []map[string][]string
}

// apiSpecAnalyser analyses the swagger doc and provides helper methods to retrieve all the end points that can
// be used as terraform resources. These endpoints have to meet certain criteria to be considered eligible resources
// as explained below:
// A resource is considered any end point that meets the following:
// 	- POST operation on the root path (e,g: api/users)
//	- GET operation on the instance path (e,g: api/users/{id}). Other operations like DELETE, PUT are optional
// In the example above, the resource name would be 'users'.
// Versioning is also supported, thus if the endpoint above had been api/v1/users the corresponding resouce name would
// have been 'users_v1'
type OpenAPISpecAnalyser interface {
	GetHost() string
	GetTerraformCompliantResources() ([]OpenApiResource, error)
	GetSecurity() OpenAPISecurity
}

type OpenAPISpecAnalyserVersion string

const (
	OpenAPIv2SpecAnalyser OpenAPISpecAnalyserVersion = "v2"
)

func CreateOpenAPISpecAnalyser(openApiSpecAnalyserVersion OpenAPISpecAnalyserVersion, openAPIDocumentURL string) (OpenAPISpecAnalyser, error) {
	var err error
	var openApiSpecAnalyser OpenAPISpecAnalyser
	switch openApiSpecAnalyserVersion {
	case OpenAPIv2SpecAnalyser:
		openApiSpecAnalyser, err = openapiv2.NewOpenApiV2SpecAnalyser(openAPIDocumentURL)
	default:
		return nil, fmt.Errorf("open api spec analyser version '%s' not supported, please choose a valid OpenAPISpecAnalyser implementation [%s]", openApiSpecAnalyserVersion, OpenAPIv2SpecAnalyser)
	}
	if err != nil {
		return nil, err
	}
	return openApiSpecAnalyser, nil
}
