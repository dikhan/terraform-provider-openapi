package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi/version"
	"net/http"
	"strings"
	"time"

	"github.com/dikhan/terraform-provider-openapi/v2/openapi/terraformutils"

	"log"

	"github.com/dikhan/http_goclient"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
	var dataSourcesInstance map[string]*schema.Resource
	var err error

	openAPIBackendConfiguration, err := p.specAnalyser.GetAPIBackendConfiguration()
	if err != nil {
		return nil, err
	}

	if resourceMap, dataSourcesInstance, err = p.createTerraformProviderResourceMapAndDataSourceInstanceMap(); err != nil {
		return nil, err
	}

	resourceNames := p.getResourceNames(resourceMap)
	providerConfigurationEndPoints := &providerConfigurationEndPoints{resourceNames}

	if providerSchema, err = p.createTerraformProviderSchema(openAPIBackendConfiguration, providerConfigurationEndPoints); err != nil {
		return nil, err
	}
	if dataSources, err = p.createTerraformProviderDataSourceMap(); err != nil {
		return nil, err
	}

	for k, v := range dataSourcesInstance {
		dataSources[k] = v
	}

	provider := &schema.Provider{
		Schema:         providerSchema,
		ResourcesMap:   resourceMap,
		DataSourcesMap: dataSources,
		ConfigureFunc:  p.configureProvider(openAPIBackendConfiguration, providerConfigurationEndPoints),
	}
	return provider, nil
}

// createTerraformProviderSchema adds support for specific provider configuration such as:
// - api key auth which will be used as the authentication mechanism when making http requests to the service provider
// - specific headers used in operations
// - endpoints override in case the user wants to point the resource to a different API (e,g: staging environment endpoint)
func (p providerFactory) createTerraformProviderSchema(openAPIBackendConfiguration SpecBackendConfiguration, providerConfigurationEndPoints *providerConfigurationEndPoints) (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}

	isMultiRegion, host, regions, err := openAPIBackendConfiguration.IsMultiRegion()
	if err != nil {
		return nil, err
	}
	if isMultiRegion {
		log.Printf("[DEBUG] service provider is configured with multi-region. API calls will be made against %s and the region provided by the user (or the default value otherwise, being the first element of supported region list: %+v), unless overridden by specific resources", host, regions)
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
		secDefName := securityDefinition.GetTerraformConfigurationName()
		required := false
		if globalSecuritySchemes.securitySchemeExists(securityDefinition) {
			required = true
		}
		p.configureProviderPropertyFromPluginConfig(s, secDefName, required)
	}

	headers := p.specAnalyser.GetAllHeaderParameters()
	log.Printf("[DEBUG] all header parameters: %+v", headers)
	for _, headerParam := range headers {
		headerTerraformCompliantName := headerParam.GetHeaderTerraformConfigurationName()
		p.configureProviderPropertyFromPluginConfig(s, headerTerraformCompliantName, false)
	}

	if providerConfigurationEndPoints != nil {
		endpoints := providerConfigurationEndPoints.endpointsSchema()
		if endpoints != nil {
			s[providerPropertyEndPoints] = endpoints
		}
	}

	return s, nil
}

// getResourceNames returns the resources exposed by the provider. The list of resources names returned will then be
// used to create the provider's endpoint schema property as well as to configure the endpoints values with the data
// provided bu the user
func (p providerFactory) getResourceNames(resourceMap map[string]*schema.Resource) []string {
	var resourceNames []string
	for resourceName := range resourceMap {
		resourceNames = append(resourceNames, strings.Replace(resourceName, fmt.Sprintf("%s_", p.name), "", 1))
	}
	return resourceNames
}

func (p providerFactory) configureProviderPropertyFromPluginConfig(providerSchema map[string]*schema.Schema, schemaPropertyName string, required bool) {
	var defaultValue = ""
	var err error
	schemaPropertyConfiguration := p.serviceConfiguration.GetSchemaPropertyConfiguration(schemaPropertyName)
	if schemaPropertyConfiguration != nil {
		err = schemaPropertyConfiguration.ExecuteCommand()
		if err != nil {
			log.Printf("[ERROR] %s", err)
		}
		defaultValue, err = schemaPropertyConfiguration.GetDefaultValue()
		if err != nil {
			log.Printf("[ERROR] %s", err)
		}
	}
	providerSchema[schemaPropertyName] = terraformutils.CreateStringSchemaProperty(schemaPropertyName, required, defaultValue)
	log.Printf("[DEBUG] registered new property '%s' (required=%t) into provider schema", schemaPropertyName, required)
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

func (p providerFactory) createTerraformProviderDataSourceMap() (map[string]*schema.Resource, error) {
	dataSourceMap := map[string]*schema.Resource{}
	openAPIDataResources := p.specAnalyser.GetTerraformCompliantDataSources()
	for _, openAPIDataSource := range openAPIDataResources {
		dataSourceName, err := p.getProviderResourceName(openAPIDataSource.GetResourceName())
		if err != nil {
			return nil, err
		}
		start := time.Now()
		d := newDataSourceFactory(openAPIDataSource)
		dataSourceTFSchema, err := d.createTerraformDataSource()
		if err != nil {
			return nil, err
		}
		log.Printf("[INFO] data source '%s' successfully registered in the provider (time:%s)", dataSourceName, time.Since(start))
		dataSourceMap[dataSourceName] = dataSourceTFSchema
	}
	return dataSourceMap, nil
}

// createTerraformProviderResourceMapAndDataSourceInstanceMap is responsible for building the following:
// - a map containing the resources that are terraform compatible
// - a map containing the data sources from the resources that are terraform compatible. This data sources enable data
//  source configuration on the resource instance GET operation.
func (p providerFactory) createTerraformProviderResourceMapAndDataSourceInstanceMap() (resourceMap, dataSourceInstanceMap map[string]*schema.Resource, err error) {
	resourceMap = map[string]*schema.Resource{}
	dataSourceInstanceMap = map[string]*schema.Resource{}
	openAPIResources, err := p.specAnalyser.GetTerraformCompliantResources()
	if err != nil {
		return nil, nil, err
	}
	for _, openAPIResource := range openAPIResources {
		start := time.Now()

		resourceName, err := p.getProviderResourceName(openAPIResource.GetResourceName())
		if err != nil {
			return nil, nil, err
		}

		if openAPIResource.ShouldIgnoreResource() {
			log.Printf("[WARN] '%s' is marked to be ignored and therefore skipping resource registration into the provider", openAPIResource.GetResourceName())
			continue
		}

		r := newResourceFactory(openAPIResource)
		d := newDataSourceInstanceFactory(openAPIResource)
		fullDataSourceInstanceName, _ := p.getProviderResourceName(d.getDataSourceInstanceName())

		if _, alreadyThere := resourceMap[resourceName]; alreadyThere {
			log.Printf("[WARN] '%s' is a duplicate resource name and is being removed from the provider", openAPIResource.GetResourceName())
			delete(resourceMap, resourceName)
			delete(dataSourceInstanceMap, fullDataSourceInstanceName)
			continue
		}

		// Register resource
		resource, err := r.createTerraformResource()
		if err != nil {
			return nil, nil, err
		}
		log.Printf("[INFO] resource '%s' successfully registered in the provider (time:%s)", resourceName, time.Since(start))
		resourceMap[resourceName] = resource

		// Register data source instance
		dataSourceInstance, _ := d.createTerraformInstanceDataSource() // if createTerraformResource did not throw an error, it's assumed that the data source instance would work too considering it's subset of the resource
		log.Printf("[INFO] data source instance '%s' successfully registered in the provider (time:%s)", fullDataSourceInstanceName, time.Since(start))
		dataSourceInstanceMap[fullDataSourceInstanceName] = dataSourceInstance
	}
	return resourceMap, dataSourceInstanceMap, nil
}

func (p providerFactory) configureProvider(openAPIBackendConfiguration SpecBackendConfiguration, providerConfigurationEndPoints *providerConfigurationEndPoints) schema.ConfigureFunc {
	return func(data *schema.ResourceData) (interface{}, error) {
		globalSecuritySchemes, err := p.specAnalyser.GetSecurity().GetGlobalSecuritySchemes()
		if err != nil {
			return nil, err
		}
		authenticator := newAPIAuthenticator(&globalSecuritySchemes)
		config, err := p.createProviderConfig(data, providerConfigurationEndPoints)
		if err != nil {
			return nil, err
		}
		telemetryHandler := p.GetTelemetryHandler(data)
		if telemetryHandler != nil {
			telemetryHandler.SubmitPluginExecutionMetrics()
		}
		openAPIClient := &ProviderClient{
			openAPIBackendConfiguration: openAPIBackendConfiguration,
			apiAuthenticator:            authenticator,
			httpClient:                  &http_goclient.HttpClient{HttpClient: &http.Client{}},
			providerConfiguration:       *config,
			telemetryHandler:            telemetryHandler,
		}
		return openAPIClient, nil
	}
}

// GetTelemetryHandler returns a handler containing validated telemetry providers
func (p providerFactory) GetTelemetryHandler(data *schema.ResourceData) TelemetryHandler {
	telemetryProvider := p.serviceConfiguration.GetTelemetryConfiguration()
	if telemetryProvider != nil {
		err := telemetryProvider.Validate()
		if err != nil {
			log.Printf("[WARN] telemetry validation failed: %s, ignoring telemetry", err)
			return nil
		}
	}
	return telemetryHandlerTimeoutSupport{
		timeout:           telemetryTimeout,
		providerName:      p.name,
		openAPIVersion:    version.Version,
		telemetryProvider: telemetryProvider,
		data:              data,
	}
}

// createProviderConfig returns a providerConfiguration populated with:
// - Header values that might be required by API operations
// - Security definition values that might be required by API operations (or globally)
// configuration mapped to the corresponding
func (p providerFactory) createProviderConfig(data *schema.ResourceData, providerConfigurationEndPoints *providerConfigurationEndPoints) (*providerConfiguration, error) {
	providerConfiguration, err := newProviderConfiguration(p.specAnalyser, data, providerConfigurationEndPoints)
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
