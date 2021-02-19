package openapi

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const providerPropertyRegion = "region"
const providerPropertyEndPoints = "endpoints"

// providerConfiguration contains all the configuration related to the OpenAPI provider. The configuration at the moment
// supports:
// - ProviderSchemaProperties: The headers map contains the header names as well as the values provided by the user in the terraform configuration
// file. These headers may be sent as part of the HTTP calls if the resource requires them (as specified in the swagger doc)
// - Security Definitions: The security definitions map contains the security definition names as well as the values provided by the user in the terraform configuration
// file. These headers may be sent as part of the HTTP calls if the resource requires them (as specified in the swagger doc)
// - Endpoints contains the endpoints configured by the user, which effectively will override the default host set in the swagger file
// - Region contains the region if user provided value for it (only supported for multi-region providers)
type providerConfiguration struct {
	Headers                   map[string]string
	SecuritySchemaDefinitions map[string]specAPIKeyAuthenticator
	Endpoints                 map[string]string
	Region                    string
}

// createProviderConfig returns a providerConfiguration populated with the values provided by the user in the provider's terraform
// configuration mapped to the corresponding
func newProviderConfiguration(specAnalyser SpecAnalyser, data *schema.ResourceData, providerConfigurationEndPoints *providerConfigurationEndPoints) (*providerConfiguration, error) {
	providerConfiguration := &providerConfiguration{}
	providerConfiguration.Headers = map[string]string{}
	providerConfiguration.Endpoints = map[string]string{}
	providerConfiguration.SecuritySchemaDefinitions = map[string]specAPIKeyAuthenticator{}

	securitySchemaDefinitions, err := specAnalyser.GetSecurity().GetAPIKeySecurityDefinitions()
	if err != nil {
		return nil, err
	}
	if securitySchemaDefinitions != nil {
		for _, secDef := range *securitySchemaDefinitions {
			secDefTerraformCompliantName := secDef.GetTerraformConfigurationName()
			if value, exists := data.GetOkExists(secDefTerraformCompliantName); exists {
				providerConfiguration.SecuritySchemaDefinitions[secDefTerraformCompliantName] = createAPIKeyAuthenticator(secDef, value.(string))
			} else {
				// Initialise the api authenticator with an empty value since the user did not provide one
				providerConfiguration.SecuritySchemaDefinitions[secDefTerraformCompliantName] = createAPIKeyAuthenticator(secDef, "")
			}
		}
	}

	headers := specAnalyser.GetAllHeaderParameters()
	if headers != nil {
		for _, headerParam := range headers {
			headerTerraformCompliantName := headerParam.GetHeaderTerraformConfigurationName()
			if value, exists := data.GetOkExists(headerTerraformCompliantName); exists {
				providerConfiguration.Headers[headerTerraformCompliantName] = value.(string)
			}
		}
	}

	region := data.Get(providerPropertyRegion)
	if region != nil {
		providerConfiguration.Region = region.(string)
	}

	if providerConfigurationEndPoints != nil {
		providerConfiguration.Endpoints = providerConfigurationEndPoints.configureEndpoints(data)
	}

	return providerConfiguration, nil
}

func (p *providerConfiguration) getAuthenticatorFor(s SpecSecurityScheme) specAPIKeyAuthenticator {
	securitySchemeConfigName := s.GetTerraformConfigurationName()
	return p.SecuritySchemaDefinitions[securitySchemeConfigName]
}

func (p *providerConfiguration) getHeaderValueFor(s SpecHeaderParam) string {
	headerConfigName := s.GetHeaderTerraformConfigurationName()
	return p.Headers[headerConfigName]
}

// getRegion returns the region value provided by the user in the configuration for the provider
func (p *providerConfiguration) getRegion() string {
	return p.Region
}

// getEndPoint resolves the endpoint value for a given resource name
func (p *providerConfiguration) getEndPoint(resourceName string) string {
	if p.Endpoints != nil {
		if endpoint, ok := p.Endpoints[resourceName]; ok {
			return endpoint
		}
	}
	return ""
}
