package openapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"

	"log"

	"github.com/dikhan/http_goclient"
	"github.com/hashicorp/terraform/helper/schema"
)

type providerFactory struct {
	name                 string
	specAnalyser         SpecAnalyser
	serviceConfiguration ServiceConfiguration
}

func newProviderFactory(name string, specAnalyser SpecAnalyser, serviceConfiguration ServiceConfiguration) (*providerFactory, error) {
	if name == "" {
		return nil, fmt.Errorf("provider name not specified")
	}
	if compliantName := terraformutils.ConvertToTerraformCompliantName(name); name != compliantName {
		return nil, fmt.Errorf("provider name '%s' not terraform name compliant, please consider renaming provider to '%s'", name, compliantName)
	}
	if specAnalyser == nil {
		return nil, fmt.Errorf("provider missing an OpenAPI Spec Analyser")
	}
	if serviceConfiguration == nil {
		return nil, fmt.Errorf("provider missing the service configuration")
	}
	return &providerFactory{
		name:                 name,
		specAnalyser:         specAnalyser,
		serviceConfiguration: serviceConfiguration,
	}, nil
}

func (p providerFactory) createProvider() (*schema.Provider, error) {
	var providerSchema map[string]*schema.Schema
	var resourceMap map[string]*schema.Resource
	var dataSources map[string]*schema.Resource
	var err error

	openAPIBackendConfiguration, err := p.specAnalyser.GetAPIBackendConfiguration()
	if err != nil {
		return nil, err
	}

	if providerSchema, err = p.createTerraformProviderSchema(openAPIBackendConfiguration); err != nil {
		return nil, err
	}
	if resourceMap, err = p.createTerraformProviderResourceMap(); err != nil {
		return nil, err
	}

	if dataSources, err = p.createTerraformProviderDataSourceMap(); err != nil {
		return nil, err
	}

	provider := &schema.Provider{
		Schema:         providerSchema,
		ResourcesMap:   resourceMap,
		DataSourcesMap: dataSources,
		ConfigureFunc:  p.configureProvider(openAPIBackendConfiguration),
	}
	return provider, nil
}

// createTerraformProviderSchema adds support for specific provider configuration such as:
// - api key auth which will be used as the authentication mechanism when making http requests to the service provider
// - specific headers used in operations
func (p providerFactory) createTerraformProviderSchema(openAPIBackendConfiguration SpecBackendConfiguration) (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}

	isMultiRegion, host, regions, err := openAPIBackendConfiguration.isMultiRegion()
	if err != nil {
		return nil, err
	}
	if isMultiRegion {
		log.Printf("[DEBUG] service provider is configured with multi-region. API calls will be made against %s and the region provided by the user (or the default value otherwise, being the first element of supported region list: %+v), unless overriden by specific resources", host, regions)
		if err := p.configureProviderProperty(s, providerPropertyRegion, regions[0], true, regions); err != nil {
			return nil, err
		}
	}

	// Override security definitions to required if they are global security schemes
	globalSecuritySchemes, err := p.specAnalyser.GetSecurity().GetGlobalSecuritySchemes()
	if err != nil {
		return nil, err
	}

	// Add all security definitions as optional properties
	securityDefinitions, err := p.specAnalyser.GetSecurity().GetAPIKeySecurityDefinitions()
	if err != nil {
		return nil, err
	}
	for _, securityDefinition := range *securityDefinitions {
		secDefName := securityDefinition.getTerraformConfigurationName()
		required := false
		if globalSecuritySchemes.securitySchemeExists(securityDefinition) {
			required = true
		}
		if err := p.configureProviderPropertyFromPluginConfig(s, secDefName, required); err != nil {
			return nil, err
		}
	}

	headers, err := p.specAnalyser.GetAllHeaderParameters()
	log.Printf("[DEBUG] all header parameters: %+v", headers)
	if err != nil {
		return nil, err
	}
	for _, headerParam := range headers {
		headerTerraformCompliantName := headerParam.GetHeaderTerraformConfigurationName()
		if err := p.configureProviderPropertyFromPluginConfig(s, headerTerraformCompliantName, false); err != nil {
			return nil, err
		}
	}

	providerConfigurationEndPoints, err := newProviderConfigurationEndPoints(p.specAnalyser)
	if err != nil {
		return nil, err
	}
	endpoints, err := providerConfigurationEndPoints.endpointsSchema()
	if err != nil {
		return nil, err
	}
	if endpoints != nil {
		s[providerPropertyEndPoints] = endpoints
	}
	return s, nil
}

func (p providerFactory) configureProviderPropertyFromPluginConfig(providerSchema map[string]*schema.Schema, schemaPropertyName string, required bool) error {
	var defaultValue = ""
	var err error
	schemaPropertyConfiguration := p.serviceConfiguration.GetSchemaPropertyConfiguration(schemaPropertyName)
	if schemaPropertyConfiguration != nil {
		err = schemaPropertyConfiguration.ExecuteCommand()
		if err != nil {
			return err
		}
		defaultValue, err = schemaPropertyConfiguration.GetDefaultValue()
		if err != nil {
			return err
		}
	}
	providerSchema[schemaPropertyName] = terraformutils.CreateStringSchemaProperty(schemaPropertyName, required, defaultValue)
	log.Printf("[DEBUG] registered new property '%s' into provider schema", schemaPropertyName)
	return nil
}

func (p providerFactory) configureProviderProperty(providerSchema map[string]*schema.Schema, schemaPropertyName string, defaultValue string, required bool, allowedValues []string) error {
	providerSchema[schemaPropertyName] = terraformutils.CreateStringSchemaProperty(schemaPropertyName, required, defaultValue)
	providerSchema[schemaPropertyName].ValidateFunc = p.createValidateFunc(allowedValues)
	log.Printf("[DEBUG] registered new property '%s' into provider schema", schemaPropertyName)
	return nil
}

func (p providerFactory) createValidateFunc(allowedValues []string) func(val interface{}, key string) (warns []string, errs []error) {
	if len(allowedValues) > 0 {
		return func(value interface{}, key string) ([]string, []error) {
			userValue := value.(string)
			for _, allowedValue := range allowedValues {
				if userValue == allowedValue {
					return nil, nil
				}
			}
			return nil, []error{fmt.Errorf("property %s value %s is not valid, please make sure the value is one of %+v", key, userValue, allowedValues)}
		}
	}
	return nil
}

// TODO: add tests for this method
func (p providerFactory) createTerraformProviderDataSourceMap() (map[string]*schema.Resource, error) {
	dataSourceMap := map[string]*schema.Resource{}
	openAPIDataResources := p.specAnalyser.GetTerraformCompliantDataSources()
	for _, openAPIDataSource := range openAPIDataResources {
		dataSourceName, err := p.getProviderResourceName(openAPIDataSource.getResourceName())
		fmt.Println(dataSourceName) // TODO: remove this (added to fix compile issues)
		if err != nil {
			return nil, err
		}
		// TODO: create data resource d := newDataSourceFactory(openAPIDataSource)
		// TODO: build schema resource calling d.createTerraformDataSource()
		// TODO: add new data source to dataSourceMap
	}
	return dataSourceMap, nil
}

func (p providerFactory) createTerraformProviderResourceMap() (map[string]*schema.Resource, error) {
	resourceMap := map[string]*schema.Resource{}
	openAPIResources, err := p.specAnalyser.GetTerraformCompliantResources()
	if err != nil {
		return nil, err
	}
	for _, openAPIResource := range openAPIResources {
		resourceName, err := p.getProviderResourceName(openAPIResource.getResourceName())
		if err != nil {
			return nil, err
		}
		start := time.Now()
		if openAPIResource.shouldIgnoreResource() {
			log.Printf("[WARN] '%s' is marked to be ignored and therefore skipping resource registration into the provider", openAPIResource.getResourceName())
			continue
		}
		_, alreadyThere := resourceMap[resourceName]
		if alreadyThere {
			log.Printf("[WARN] '%s' is a duplicate resource name and is being removed from the provider", openAPIResource.getResourceName())
			delete(resourceMap, resourceName)
			continue
		}
		r := newResourceFactory(openAPIResource)
		resource, err := r.createTerraformResource()
		if err != nil {
			return nil, err
		}
		log.Printf("[INFO] resource '%s' successfully registered in the provider (time:%s)", resourceName, time.Since(start))
		resourceMap[resourceName] = resource
	}
	return resourceMap, nil
}

func (p providerFactory) configureProvider(openAPIBackendConfiguration SpecBackendConfiguration) schema.ConfigureFunc {
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
		openAPIClient := &ProviderClient{
			openAPIBackendConfiguration: openAPIBackendConfiguration,
			apiAuthenticator:            authenticator,
			httpClient:                  &http_goclient.HttpClient{HttpClient: &http.Client{}},
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
	providerConfiguration, err := newProviderConfiguration(p.specAnalyser, data)
	if err != nil {
		return nil, err
	}
	return providerConfiguration, nil
}

func (p providerFactory) getProviderResourceName(resourceName string) (string, error) {
	if resourceName == "" {
		return "", fmt.Errorf("resource name can not be empty")
	}
	fullResourceName := fmt.Sprintf("%s_%s", p.name, resourceName)
	return fullResourceName, nil
}
