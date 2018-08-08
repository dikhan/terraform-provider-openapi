package openapi

import (
	"errors"
	"fmt"

	"net/http"

	"github.com/dikhan/http_goclient"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type providerFactory struct {
	name            string
	discoveryAPIURL string
}

type providerConfig struct {
	Headers                   map[string]string
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
		Schema:        p.createTerraformProviderSchema(apiSpecAnalyser.d.Spec()),
		ResourcesMap:  resourceMap,
		ConfigureFunc: p.configureProvider(apiSpecAnalyser.d.Spec()),
	}
	return provider, nil
}

// createTerraformProviderSchema adds support for specific provider configuration such as api key which will
// be used as the authentication mechanism when making http requests to the service provider
func (p providerFactory) createTerraformProviderSchema(spec *spec.Swagger) map[string]*schema.Schema {
	s := map[string]*schema.Schema{}
	for secDefName, secDef := range spec.SecurityDefinitions {
		if secDef.Type == "apiKey" {
			s[secDefName] = &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			}
		}
	}
	headerConfigProps := openapiutils.GetAllHeaderParameters(spec)
	for headerConfigProp := range headerConfigProps {
		s[headerConfigProp] = &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		}
	}
	return s
}

func (p providerFactory) configureProvider(spec *spec.Swagger) schema.ConfigureFunc {
	return func(data *schema.ResourceData) (interface{}, error) {
		config := providerConfig{}
		config.SecuritySchemaDefinitions = map[string]authenticator{}
		for secDefName, secDef := range spec.SecurityDefinitions {
			if secDef.Type == "apiKey" {
				config.SecuritySchemaDefinitions[secDefName] = createAPIKeyAuthenticator(secDef.In, secDef.Name, data.Get(secDefName).(string))
			}
		}
		config.Headers = map[string]string{}
		headerConfigProps := openapiutils.GetAllHeaderParameters(spec)
		// Here we only save the value of the header with its corresponding identifier, which is defined in the keys
		// saved in the map returned headerConfigProps
		for headerConfigProp := range headerConfigProps {
			config.Headers[headerConfigProp] = data.Get(headerConfigProp).(string)
		}
		return config, nil
	}
}

func (p providerFactory) getProviderResourceName(resourceName string) string {
	fullResourceName := fmt.Sprintf("%s_%s", p.name, resourceName)
	return fullResourceName
}
