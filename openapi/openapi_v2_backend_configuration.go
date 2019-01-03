package openapi

import (
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

func (o specV2BackendConfiguration) getDefaultRegion() (string, error) {
	isMultiRegion, _, regions, err := o.isMultiRegion()
	if !isMultiRegion {
		if err != nil {
			return "", fmt.Errorf("failed to get default region value: %s", err)
		}
		return "", fmt.Errorf("failed to get default region value: service is not multi-region")
	}
	if err != nil {
		return "", err
	}
	return regions[0], nil
}

func (o specV2BackendConfiguration) isMultiRegion() (bool, string, []string, error) {
	if host, exists := o.spec.Extensions.GetString(extTfProviderMultiRegionFQDN); exists {
		isMultiRegion, _ := openapiutils.IsMultiRegionHost(host)
		if !isMultiRegion {
			return false, "", nil, fmt.Errorf("'%s' extension value provided not matching multiregion host format", extTfProviderMultiRegionFQDN)
		}
		regionsExtensionValue, regionsExtensionExists := o.spec.Extensions.GetString(extTfProviderRegions)
		if !regionsExtensionExists || regionsExtensionValue == "" {
			return false, "", nil, fmt.Errorf("'%s' extension missing or empty value provided", extTfProviderRegions)
		}
		regions := strings.Split(strings.Replace(regionsExtensionValue, " ", "", -1), ",")
		return true, host, regions, nil
	}
	return false, "", nil, nil
}

func (o specV2BackendConfiguration) getBasePath() string {
	return o.spec.BasePath
}

func (o specV2BackendConfiguration) getHTTPSchemes() []string {
	return o.spec.Schemes
}
