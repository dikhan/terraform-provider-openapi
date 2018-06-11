package openapi

import (
	"errors"
	"fmt"

	"net/http"

	"github.com/dikhan/http_goclient"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type providerFactory struct {
	name            string
	discoveryAPIURL string
}

type providerConfig struct {
	SecuritySchemaDefinitions map[string]authenticator
}

func (p providerFactory) createProvider() (*schema.Provider, error) {
	apiSpecAnalyser, err := p.createAPISpecAnalyser()
	if err != nil {
		return nil, fmt.Errorf("error occurred while retrieving api specification. Error=%s", err)
	}
	provider, err := p.generateProviderFromAPISpec(apiSpecAnalyser)
	if err != nil {
		return nil, fmt.Errorf("error occurred while creating schema provider. Error=%s", err)
	}
	return provider, nil
}

func (p providerFactory) createAPISpecAnalyser() (*apiSpecAnalyser, error) {
	if p.discoveryAPIURL == "" {
		return nil, errors.New("required param 'apiUrl' missing")
	}
	apiSpec, err := loads.JSONSpec(p.discoveryAPIURL)
	if err != nil {
		return nil, fmt.Errorf("error occurred when retrieving api spec from %s. Error=%s", p.discoveryAPIURL, err)
	}
	apiSpecAnalyser := &apiSpecAnalyser{apiSpec}
	return apiSpecAnalyser, nil
}

func (p providerFactory) generateProviderFromAPISpec(apiSpecAnalyser *apiSpecAnalyser) (*schema.Provider, error) {
	resourceMap := map[string]*schema.Resource{}
	resourcesInfo, err := apiSpecAnalyser.getResourcesInfo()
	if err != nil {
		return nil, err
	}
	for resourceName, resourceInfo := range resourcesInfo {
		r := resourceFactory{
			http_goclient.HttpClient{HttpClient: &http.Client{}},
			resourceInfo,
			newAPIAuthenticator(apiSpecAnalyser.d.Spec().Security),
		}
		resource, err := r.createSchemaResource()
		if err != nil {
			return nil, err
		}
		resourceName := p.getProviderResourceName(resourceName)
		resourceMap[resourceName] = resource
	}
	provider := &schema.Provider{
		Schema:        p.createTerraformProviderSchema(apiSpecAnalyser.d.Spec().SecurityDefinitions),
		ResourcesMap:  resourceMap,
		ConfigureFunc: p.configureProvider(apiSpecAnalyser.d.Spec().SecurityDefinitions),
	}
	return provider, nil
}

// createTerraformProviderSchema adds support for specific provider configuration such as api key which will
// be used as the authentication mechanism when making http requests to the service provider
func (p providerFactory) createTerraformProviderSchema(securityDefinitions spec.SecurityDefinitions) map[string]*schema.Schema {
	s := map[string]*schema.Schema{}
	for secDefName, secDef := range securityDefinitions {
		if secDef.Type == "apiKey" {
			s[secDefName] = &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			}
		}
	}
	return s
}

func (p providerFactory) configureProvider(securityDefinitions spec.SecurityDefinitions) schema.ConfigureFunc {
	return func(data *schema.ResourceData) (interface{}, error) {
		config := providerConfig{}
		config.SecuritySchemaDefinitions = map[string]authenticator{}
		for secDefName, secDef := range securityDefinitions {
			if secDef.Type == "apiKey" {
				config.SecuritySchemaDefinitions[secDefName] = createAPIKeyAuthenticator(secDef.In, secDef.Name, data.Get(secDefName).(string))
			}
		}
		return config, nil
	}
}

func (p providerFactory) getProviderResourceName(resourceName string) string {
	fullResourceName := fmt.Sprintf("%s_%s", p.name, resourceName)
	return fullResourceName
}
