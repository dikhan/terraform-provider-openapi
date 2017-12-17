package main

import (
	"errors"
	"fmt"

	"net/http"

	"github.com/dikhan/http_goclient"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

type ProviderFactory struct {
	Name            string
	DiscoveryApiUrl string
}

type ApiKey struct {
	Name  string
	Value string
}

type SecuritySchemaDefinition struct {
	ApiKeyHeader ApiKey
	ApiKeyQuery  ApiKey
}

type ProviderConfig struct {
	SecuritySchemaDefinitions map[string]SecuritySchemaDefinition
}

func (p ProviderFactory) createProvider() (*schema.Provider, error) {
	apiSpecAnalyser, err := p.createApiSpecAnalyser()
	if err != nil {
		return nil, fmt.Errorf("error occurred while retrieving api specification. Error=%s", err)
	}
	provider, err := p.generateProviderFromApiSpec(apiSpecAnalyser)
	if err != nil {
		return nil, fmt.Errorf("error occurred while creating schema provider. Error=%s", err)
	}
	return provider, nil
}

func (p ProviderFactory) createApiSpecAnalyser() (*ApiSpecAnalyser, error) {
	if p.DiscoveryApiUrl == "" {
		return nil, errors.New("required param 'apiUrl' missing")
	}
	apiSpec, err := loads.JSONSpec(p.DiscoveryApiUrl)
	if err != nil {
		return nil, fmt.Errorf("error occurred when retrieving api spec from %s. Error=%s", p.DiscoveryApiUrl, err)
	}
	apiSpecAnalyser := &ApiSpecAnalyser{apiSpec}
	return apiSpecAnalyser, nil
}

func (p ProviderFactory) generateProviderFromApiSpec(apiSpecAnalyser *ApiSpecAnalyser) (*schema.Provider, error) {
	resourceMap := map[string]*schema.Resource{}
	for resourceName, resourceInfo := range apiSpecAnalyser.getCrudResources() {
		r := ResourceFactory{
			http_goclient.HttpClient{HttpClient: &http.Client{}},
			resourceInfo,
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
func (p ProviderFactory) createTerraformProviderSchema(securityDefinitions spec.SecurityDefinitions) map[string]*schema.Schema {
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

func (p ProviderFactory) configureProvider(securityDefinitions spec.SecurityDefinitions) schema.ConfigureFunc {
	return func(data *schema.ResourceData) (interface{}, error) {
		config := ProviderConfig{}
		config.SecuritySchemaDefinitions = map[string]SecuritySchemaDefinition{}
		for secDefName, secDef := range securityDefinitions {
			if secDef.Type == "apiKey" {
				securitySchemaDefinition := SecuritySchemaDefinition{}
				switch secDef.In {
				case "header":
					securitySchemaDefinition.ApiKeyHeader = ApiKey{secDef.Name, data.Get(secDefName).(string)}
				case "query":
					securitySchemaDefinition.ApiKeyQuery = ApiKey{secDef.Name, data.Get(secDefName).(string)}
				}
				config.SecuritySchemaDefinitions[secDefName] = securitySchemaDefinition
			}
		}
		return config, nil
	}
}

func (p ProviderFactory) getProviderResourceName(resourceName string) string {
	fullResourceName := fmt.Sprintf("%s_%s", p.Name, resourceName)
	return fullResourceName
}
