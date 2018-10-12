package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/go-openapi/spec"
	"log"
)

type specV2BackendConfiguration struct {
	openAPIDocumentURL string
	spec               *spec.Swagger
}

func newOpenAPIBackendConfigurationV2(spec *spec.Swagger, openAPIDocumentURL string) (*specV2BackendConfiguration, error) {
	if spec.Swagger != "2.0" {
		return nil, fmt.Errorf("swagger version '%s' not supported, specV2BackendConfiguration only supports 2.0", spec.Swagger)
	}
	if openAPIDocumentURL == "" {
		return nil, fmt.Errorf("missing mandatory parameter openAPIDocumentURL")
	}
	return &specV2BackendConfiguration{openAPIDocumentURL, spec}, nil
}

func (o specV2BackendConfiguration) getHost() (string, error) {
	if o.spec.Host == "" {
		log.Printf("[WARN] host field not specified in the swagger configuration, falling back to retrieving the host from where the OpenAPI document is served: '%s'", o.openAPIDocumentURL)
		hostFromURL := openapiutils.GetHostFromURL(o.openAPIDocumentURL)
		if hostFromURL == "" {
			return "", fmt.Errorf("could not find valid host from URL provided: '%s'", o.openAPIDocumentURL)
		}
		return hostFromURL, nil
	}
	return o.spec.Host, nil
}

func (o specV2BackendConfiguration) getBasePath() string {
	return o.spec.BasePath
}

func (o specV2BackendConfiguration) getHTTPSchemes() []string {
	return o.spec.Schemes
}
