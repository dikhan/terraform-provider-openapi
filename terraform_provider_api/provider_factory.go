package main

import (
	"errors"
	"fmt"
	"log"

	"net/http"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

const API_KEY_HEADER_NAME = "api_key_header"
const API_KEY_QUERY_NAME = "api_key_query"

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

func (p ProviderFactory) createProvider() *schema.Provider {
	apiSpecAnalyser, err := p.createApiSpecAnalyser()
	if err != nil {
		log.Fatalf("error occurred while retrieving api specification. Error=%s", err)
	}
	provider, err := p.generateProviderFromApiSpec(apiSpecAnalyser)
	if err != nil {
		log.Fatalf("error occurred while creating schema provider. Error=%s", err)
	}
	return provider
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
			&http.Client{},
			resourceInfo,
		}
		resource := r.createSchemaResource()
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
	for _, secDef := range securityDefinitions {
		if secDef.Type == "apiKey" {
			var key string
			switch secDef.In {
			case "header":
				key = API_KEY_HEADER_NAME
			case "query":
				key = API_KEY_QUERY_NAME
			}
			s[key] = &schema.Schema{
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
					securitySchemaDefinition.ApiKeyHeader = ApiKey{secDef.Name, data.Get(API_KEY_HEADER_NAME).(string)}
				case "query":
					securitySchemaDefinition.ApiKeyQuery = ApiKey{secDef.Name, data.Get(API_KEY_QUERY_NAME).(string)}
				}
				config.SecuritySchemaDefinitions[secDefName] = securitySchemaDefinition
			}
		}
		PrettyPrint(config.SecuritySchemaDefinitions)
		return config, nil
	}
}

func (p ProviderFactory) getProviderResourceName(resourceName string) string {
	fullResourceName := fmt.Sprintf("%s_%s", p.Name, resourceName)
	return fullResourceName
}
