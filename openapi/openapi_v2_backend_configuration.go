package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/go-openapi/spec"
)

type specV2BackendConfiguration struct {
	openAPIDocumentURL string
	spec               *spec.Swagger
}

func newOpenAPIBackendConfigurationV2(spec *spec.Swagger, openAPIDocumentURL string) (specV2BackendConfiguration, error) {
	if spec.Swagger != "2.0" {
		return specV2BackendConfiguration{}, fmt.Errorf("swagger version '%s' not supported, specV2BackendConfiguration only supports 2.0 ", spec.Swagger)
	}
	return specV2BackendConfiguration{openAPIDocumentURL, spec}, nil
}

func (o specV2BackendConfiguration) getHost() string {
	if o.spec.Host == "" {
		return openapiutils.GetHostFromURL(o.openAPIDocumentURL)
	}
	return o.spec.Host
}

func (o specV2BackendConfiguration) getBasePath() string {
	return o.spec.BasePath
}

func (o specV2BackendConfiguration) getHTTPSchemes() []string {
	return o.spec.Schemes
}
