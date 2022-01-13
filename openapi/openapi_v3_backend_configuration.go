package openapi

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/getkin/kin-openapi/openapi3"
)

type specV3BackendConfiguration struct {
	openAPIDocumentURL string
	spec               *openapi3.T
}

var _ SpecBackendConfiguration = (*specV3BackendConfiguration)(nil)

func newOpenAPIBackendConfigurationV3(spec *openapi3.T, openAPIDocumentURL string) (*specV3BackendConfiguration, error) {
	if !strings.HasPrefix(spec.OpenAPI, "3.0") {
		return nil, fmt.Errorf("swagger version '%s' not supported, specV3BackendConfiguration only supports 3.0.*", spec.OpenAPI)
	}
	if openAPIDocumentURL == "" {
		return nil, fmt.Errorf("missing mandatory parameter openAPIDocumentURL")
	}
	return &specV3BackendConfiguration{openAPIDocumentURL, spec}, nil
}

func (o *specV3BackendConfiguration) getHost() (string, error) {
	server, err := o.getPreferredServer()
	if err != nil {
		return "", err
	}
	return server.Host, nil
}

func (o *specV3BackendConfiguration) getBasePath() string {
	server, err := o.getPreferredServer()
	if err != nil {
		log.Printf("[DEBUG] unable to parse preferred server: %v", err)
		return ""
	}
	return server.Path
}

func (o *specV3BackendConfiguration) getHTTPScheme() (string, error) {
	server, err := o.getPreferredServer()
	if err != nil {
		return "", err
	}
	if server.Scheme != "" {
		return server.Scheme, nil
	}
	// TODO: should this be http or https
	return "https", nil
}

func (o *specV3BackendConfiguration) getHostByRegion(region string) (string, error) {
	panic("implement me - getHostByRegion")
}

func (o *specV3BackendConfiguration) IsMultiRegion() (bool, string, []string, error) {
	// TODO: add support for multi-region backends
	return false, "", []string{}, nil
}

func (o *specV3BackendConfiguration) GetDefaultRegion(i []string) (string, error) {
	panic("implement me - GetDefaultRegion")
}

func (o *specV3BackendConfiguration) getPreferredServer() (*url.URL, error) {
	if len(o.spec.Servers) == 0 {
		log.Printf("[WARN] servers field not specified in the openapi configuration, falling back to retrieving the host from where the OpenAPI document is served: '%s'", o.openAPIDocumentURL)
		hostFromURL := openapiutils.GetHostFromURL(o.openAPIDocumentURL)
		if hostFromURL == "" {
			return nil, fmt.Errorf("could not find valid host from URL provided: '%s'", o.openAPIDocumentURL)
		}
		hostURL, err := url.Parse(hostFromURL)
		if err != nil {
			return nil, fmt.Errorf("could not parse host from OpenAPI document URL '%s' - error: %v", hostFromURL, err)
		}
		return hostURL, nil
	}
	if len(o.spec.Servers) > 1 {
		log.Printf("[INFO] using the first entry from the servers field in the openapi configuration: '%s'", o.spec.Servers[0].URL)
	}
	// TODO: define more configurable mechanism for TF user to select the desired server URL
	serverURL, err := url.Parse(o.spec.Servers[0].URL)
	if err != nil {
		return nil, fmt.Errorf("could not parse server URL '%s' - error: %v", serverURL, err)
	}
	return serverURL, nil
}
