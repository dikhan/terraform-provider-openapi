package openapi

import "github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"

type HttpMethodSupported string

const (
	HttpGet    HttpMethodSupported = "GET"
	HttpPost   HttpMethodSupported = "POST"
	HttpPut    HttpMethodSupported = "PUT"
	HttpDelete HttpMethodSupported = "DELETE"
)

type HeaderParam struct {
	Name          string
	TerraformName string
}

func (h HeaderParam) GetHeaderTerraformConfigurationName() string {
	if h.TerraformName != "" {
		return openapiutils.ConvertToTerraformCompliantFieldName(h.TerraformName)
	}
	return openapiutils.ConvertToTerraformCompliantFieldName(h.Name)
}

type openAPIBackendConfiguration interface {
	getHost() string
	getBasePath() string
	getHTTPSchemes() []string

	getHeaderParamsForPathGetOperation(path string) []HeaderParam
	getHeaderParamsForPathPostOperation(path string) []HeaderParam
	getHeaderParamsForPathPutOperation(path string) []HeaderParam
	getHeaderParamsForPathDeleteOperation(path string) []HeaderParam

	getSecurityForPathGetOperation(path string) []map[string][]string
	getSecurityForPathPostOperation(path string) []map[string][]string
	getSecurityForPathPutOperation(path string) []map[string][]string
	getSecurityForPathDeleteOperation(path string) []map[string][]string
}
