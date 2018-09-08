package openapi

import (
	"github.com/hashicorp/terraform/helper/schema"
)

type providerConfiguration struct {
	Headers                   map[string]string
	SecuritySchemaDefinitions map[string]authenticator
}

// createProviderConfig returns a providerConfiguration populated with the values provided by the user in the provider's terraform
// configuration mapped to the corresponding
func newProviderConfiguration(headers SpecHeaderParameters, securitySchemaDefinitions SpecSecurityDefinitions, data *schema.ResourceData) providerConfiguration {
	providerConfiguration := providerConfiguration{}
	providerConfiguration.Headers = map[string]string{}
	providerConfiguration.SecuritySchemaDefinitions = map[string]authenticator{}

	for _, secDef := range securitySchemaDefinitions {
		secDefTerraformCompliantName := secDef.getTerraformConfigurationName()
		if value, exists := data.GetOkExists(secDefTerraformCompliantName); exists {
			providerConfiguration.SecuritySchemaDefinitions[secDefTerraformCompliantName] = createAPIKeyAuthenticator(secDef.In, secDef.Name, value.(string))
		}
	}
	for _, headerParam := range headers {
		headerTerraformCompliantName := headerParam.GetHeaderTerraformConfigurationName()
		if value, exists := data.GetOkExists(headerTerraformCompliantName); exists {
			providerConfiguration.Headers[headerTerraformCompliantName] = value.(string)
		}
	}
	return providerConfiguration
}
