package openapi

import (
	"errors"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/go-openapi/spec"
	"log"
	"strings"
)

const extTfProviderMultiRegionFQDN = "x-terraform-provider-multiregion-fqdn"
const extTfProviderRegions = "x-terraform-provider-regions"

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

func (o specV2BackendConfiguration) getHostByRegion(region string) (string, error) {
	if region == "" {
		return "", fmt.Errorf("can't get host by region, missing region value")
	}
	isMultiRegion, host, allowedRegions, err := o.isMultiRegion()
	if err != nil {
		return "", err
	}
	if !isMultiRegion {
		return "", fmt.Errorf("missing '%s' extension or value provided not matching multiregion host format", extTfProviderMultiRegionFQDN)
	}
	if err := o.validateRegion(region, allowedRegions); err != nil {
		return "", err
	}
	overrideHost, err := openapiutils.GetMultiRegionHost(host, region)
	if err != nil {
		return "", err
	}
	return overrideHost, nil
}

func (o specV2BackendConfiguration) validateRegion(region string, allowedRegions []string) error {
	for _, r := range allowedRegions {
		if r == region {
			return nil
		}
	}
	return fmt.Errorf("region %s not matching allowed ones %+v", region, allowedRegions)
}

func (o specV2BackendConfiguration) getDefaultRegion(regions []string) (string, error) {
	if regions == nil || len(regions) == 0 {
		return "", fmt.Errorf("empty regions provided")
	}
	return regions[0], nil
}

func (o specV2BackendConfiguration) isMultiRegion() (bool, string, []string, error) {
	isHostMultiRegion, host, err := o.isHostMultiRegion()
	if err != nil {
		return false, "", nil, err
	}
	if isHostMultiRegion {
		regions, err := o.getProviderRegions()
		if err != nil {
			return false, "", nil, err
		}
		return true, host, regions, nil
	}
	return false, "", nil, nil
}

func (o specV2BackendConfiguration) isHostMultiRegion() (bool, string, error) {
	if host, exists := o.spec.Extensions.GetString(extTfProviderMultiRegionFQDN); exists {
		isMultiRegion, _ := openapiutils.IsMultiRegionHost(host)
		if !isMultiRegion {
			return false, "", fmt.Errorf("'%s' extension value provided not matching multiregion host format", extTfProviderMultiRegionFQDN)
		}
		return true, host, nil
	}
	return false, "", nil
}

func (o specV2BackendConfiguration) getProviderRegions() ([]string, error) {
	regionsExtensionValue, regionsExtensionExists := o.spec.Extensions.GetString(extTfProviderRegions)
	if !regionsExtensionExists {
		return nil, fmt.Errorf("mandatory multiregion '%s' extension missing", extTfProviderRegions)
	}
	if regionsExtensionValue == "" {
		return nil, fmt.Errorf("mandatory multiregion '%s' extension empty value provided", extTfProviderRegions)
	}
	regions := strings.Split(strings.Replace(regionsExtensionValue, " ", "", -1), ",")
	return regions, nil
}

func (o specV2BackendConfiguration) getBasePath() string {
	return o.spec.BasePath
}

func (o specV2BackendConfiguration) getHTTPSchemes() []string {
	return o.spec.Schemes
}

func (o specV2BackendConfiguration) getHTTPSchemes2() (string, error) {
	var defaultScheme string

	if len(o.spec.Schemes) == 0 {
		return "", errors.New("no schemes specified")
	}
	for _, s := range o.spec.Schemes {
		if s == "https" {
			return s, nil
		}
		if s == "http" {
			defaultScheme = s
		}
	}

	if defaultScheme == "" {
		return "", fmt.Errorf("specified schemes %s are not supported - must use http or https", o.spec.Schemes)
	}

	return defaultScheme, nil
}
