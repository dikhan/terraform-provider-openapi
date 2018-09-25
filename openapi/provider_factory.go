package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"

	"net/http"

	"github.com/dikhan/http_goclient"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
)

type providerFactory struct {
	name         string
	specAnalyser SpecAnalyser
}

func newProviderFactory(name string, specAnalyser SpecAnalyser) (*providerFactory, error) {
	if name == "" {
		return nil, fmt.Errorf("provider name not specified")
	}
	if compliantName := terraformutils.ConvertToTerraformCompliantName(name); name != compliantName {
		return nil, fmt.Errorf("provider name '%s' not terraform name compliant, please consider renaming provider to '%s'", name, compliantName)
	}
	if specAnalyser == nil {
		return nil, fmt.Errorf("provider missing an OpenAPI Spec Analyser")
	}
	return &providerFactory{
		name:         name,
		specAnalyser: specAnalyser,
	}, nil
}

func (p providerFactory) createProvider() (*schema.Provider, error) {
	var providerSchema map[string]*schema.Schema
	var resourceMap map[string]*schema.Resource
	var err error

	if providerSchema, err = p.createTerraformProviderSchema(); err != nil {
		return nil, err
	}
	if resourceMap, err = p.createTerraformProviderResourceMap(); err != nil {
		return nil, err
	}
	provider := &schema.Provider{
		Schema:        providerSchema,
		ResourcesMap:  resourceMap,
		ConfigureFunc: p.configureProvider(),
	}
	return provider, nil
}

// createTerraformProviderSchema adds support for specific provider configuration such as:
// - api key auth which will be used as the authentication mechanism when making http requests to the service provider
// - specific headers used in operations
func (p providerFactory) createTerraformProviderSchema() (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}
	globalSecuritySchemes, err := p.specAnalyser.GetSecurity().GetGlobalSecuritySchemes()
	if err != nil {
		return nil, err
	}
	for _, securityScheme := range globalSecuritySchemes {
		s[securityScheme.getTerraformConfigurationName()] = &schema.Schema{
			Type:     schema.TypeString,
			Required: true,
		}
	}
	headers, err := p.specAnalyser.GetAllHeaderParameters()
	if err != nil {
		return nil, err
	}
	for _, headerParam := range headers {
		headerTerraformCompliantName := headerParam.GetHeaderTerraformConfigurationName()
		s[headerTerraformCompliantName] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	return s, nil
}

func (p providerFactory) createTerraformProviderResourceMap() (map[string]*schema.Resource, error) {
	resourceMap := map[string]*schema.Resource{}
	openAPIResources, err := p.specAnalyser.GetTerraformCompliantResources()
	if err != nil {
		return nil, err
	}
	for _, openAPIResource := range openAPIResources {
		if openAPIResource.shouldIgnoreResource() {
			log.Printf("[WARN] '%s' is marked as to be ignored and therefore skipping resource registration into the provider", openAPIResource.getResourceName())
			continue
		}
		r := resourceFactory{
			openAPIResource,
		}
		resource, err := r.createTerraformResource()
		if err != nil {
			return nil, err
		}
		resourceName, err := p.getProviderResourceName(openAPIResource.getResourceName())
		if err != nil {
			return nil, err
		}
		log.Printf("[INFO] resource '%s' successfully registered in the provider", resourceName)
		resourceMap[resourceName] = resource
	}
	return resourceMap, nil
}

func (p providerFactory) configureProvider() schema.ConfigureFunc {
	return func(data *schema.ResourceData) (interface{}, error) {
		globalSecuritySchemes, err := p.specAnalyser.GetSecurity().GetGlobalSecuritySchemes()
		if err != nil {
			return nil, err
		}
		authenticator := newAPIAuthenticator(&globalSecuritySchemes)
		config, err := p.createProviderConfig(data)
		if err != nil {
			return nil, err
		}
		openAPIBackendConfiguration, err := p.specAnalyser.GetAPIBackendConfiguration()
		if err != nil {
			return nil, err
		}
		openAPIClient := &ProviderClient{
			openAPIBackendConfiguration: openAPIBackendConfiguration,
			apiAuthenticator:            authenticator,
			httpClient:                  http_goclient.HttpClient{HttpClient: &http.Client{}},
			providerConfiguration:       *config,
		}
		return openAPIClient, nil
	}
}

// createProviderConfig returns a providerConfiguration populated with:
// - Header values that might be required by API operations
// - Security definition values that might be required by API operations (or globally)
// configuration mapped to the corresponding
func (p providerFactory) createProviderConfig(data *schema.ResourceData) (*providerConfiguration, error) {
	securityDefinitions, err := p.specAnalyser.GetSecurity().GetAPIKeySecurityDefinitions()
	if err != nil {
		return nil, err
	}
	headers, err := p.specAnalyser.GetAllHeaderParameters()
	if err != nil {
		return nil, err
	}
	providerConfiguration := newProviderConfiguration(headers, securityDefinitions, data)
	return providerConfiguration, nil
}

func (p providerFactory) getProviderResourceName(resourceName string) (string, error) {
	if resourceName == "" {
		return "", fmt.Errorf("resource name can not be empty")
	}
	fullResourceName := fmt.Sprintf("%s_%s", p.name, resourceName)
	return fullResourceName, nil
}
