package openapiv2

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/go-openapi/spec"
)

type openAPIBackendConfigurationV2 struct {
	spec *spec.Swagger
}

func newOpenAPIBackendConfigurationV2(spec *spec.Swagger) (openAPIBackendConfigurationV2, error) {
	if spec.Swagger != "2.0" {
		return openAPIBackendConfigurationV2{}, fmt.Errorf("swagger version '%s' not supported, openAPIBackendConfigurationV2 only supports 2.0 ", spec.Swagger)
	}
	return openAPIBackendConfigurationV2{spec}, nil
}

func (o openAPIBackendConfigurationV2) getHost() string {
	return o.spec.Host
}

func (o openAPIBackendConfigurationV2) getBasePath() string {
	return o.spec.BasePath
}

func (o openAPIBackendConfigurationV2) getHTTPSchemes() []string {
	return o.spec.Schemes
}

func (o openAPIBackendConfigurationV2) getHeaderParamsForPathGetOperation(path string) []openapi.HeaderParam {
	return openapiutils.GetHeadersForPath(o.spec, path, HttpGet)
}

func (o openAPIBackendConfigurationV2) getHeaderParamsForPathPostOperation(path string) []openapi.HeaderParam {
	return openapiutils.GetHeadersForPath(o.spec, path, HttpPost)
}

func (o openAPIBackendConfigurationV2) getHeaderParamsForPathPutOperation(path string) []openapi.HeaderParam {
	return openapiutils.GetHeadersForPath(o.spec, path, HttpPut)
}

func (o openAPIBackendConfigurationV2) getHeaderParamsForPathDeleteOperation(path string) []openapi.HeaderParam {
	return openapiutils.GetHeadersForPath(o.spec, path, HttpDelete)
}

func (o openAPIBackendConfigurationV2) getSecurityForPathGetOperation(path string) []map[string][]string {
	return o.getSecurityForPath(path, HttpGet)
}

func (o openAPIBackendConfigurationV2) getSecurityForPathPostOperation(path string) []map[string][]string {
	return o.getSecurityForPath(path, HttpPost)
}

func (o openAPIBackendConfigurationV2) getSecurityForPathPutOperation(path string) []map[string][]string {
	return o.getSecurityForPath(path, HttpPut)
}

func (o openAPIBackendConfigurationV2) getSecurityForPathDeleteOperation(path string) []map[string][]string {
	return o.getSecurityForPath(path, HttpDelete)
}

func (o openAPIBackendConfigurationV2) getSecurityForPath(path string, method HttpMethodSupported) []map[string][]string {
	switch method {
	case HttpGet:
		return o.spec.Paths.Paths[path].Get.Security
	case HttpPost:
		return o.spec.Paths.Paths[path].Post.Security
	case HttpPut:
		return o.spec.Paths.Paths[path].Put.Security
	case HttpDelete:
		return o.spec.Paths.Paths[path].Delete.Security
	}
	return nil
}
