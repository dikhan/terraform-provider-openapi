package openapi

import (
	"fmt"

	"net/http"

	"github.com/dikhan/http_goclient"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type providerFactory struct {
	name                string
	openAPISpecAnalyser OpenAPISpecAnalyser
}

type providerConfig struct {
	Headers                   map[string]string
	SecuritySchemaDefinitions map[string]authenticator
}

func NewProviderFactory(name string, openAPISpecAnalyser OpenAPISpecAnalyser) (*providerFactory, error) {
	if name == "" {
		return nil, fmt.Errorf("provider name not specified")
	}
	if openAPISpecAnalyser == nil {
		return nil, fmt.Errorf("provider missing an OpenAPI Spec Analyser")
	}
	return &providerFactory{
		name:                name,
		openAPISpecAnalyser: openAPISpecAnalyser,
	}, nil
}

func (p providerFactory) createProvider() (*schema.Provider, error) {
	// If host is not specified, it is assumed to be the same host where the API documentation is being served.
	//if apiSpecAnalyser.d.Spec().Host == "" {
	//	apiSpecAnalyser.d.Spec().Host = openapiutils.GetHostFromURL(p.discoveryAPIURL)
	//}
	provider, err := p.generateProviderFromAPISpec()
	if err != nil {
		return nil, fmt.Errorf("error occurred while creating schema provider. Error=%s", err)
	}
	return provider, nil
}

func (p providerFactory) generateProviderFromAPISpec() (*schema.Provider, error) {
	resourceMap := map[string]*schema.Resource{}
	openAPIResources, err := p.openAPISpecAnalyser.GetTerraformCompliantResources()
	//resourcesInfo, err := apiSpecAnalyser.getResourcesInfo()
	if err != nil {
		return nil, err
	}
	for _, openAPIResource := range openAPIResources {
		r := resourceFactory{
			http_goclient.HttpClient{HttpClient: &http.Client{}},
			openAPIResource,
			newAPIAuthenticator(p.openAPISpecAnalyser.GetSecurity().GetGlobalSecuritySchemes()),
		}
		resource, err := r.createTerraformResource()
		if err != nil {
			return nil, err
		}
		resourceName := p.getProviderResourceName(openAPIResource.getResourceName())
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
			s[terraformutils.ConvertToTerraformCompliantName(secDefName)] = &schema.Schema{
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
