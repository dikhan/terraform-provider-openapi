package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

// specAPIKeyHeaderSecurityDefinition defines a security definition. This struct serves as a translation between the OpenAPI document
// and the scheme that will be used by the OpenAPI Terraform provider when making API calls to the backend
type specAPIKeyQuerySecurityDefinition struct {
	Name   string
	apiKey specAPIKey
}

// newAPIKeyHeaderSecurityDefinition constructs a SpecSecurityDefinition of Query type. The secDefName value is the identifier
// of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyQuerySecurityDefinition(secDefName, apiKeyName string) specAPIKeyQuerySecurityDefinition {
	return specAPIKeyQuerySecurityDefinition{secDefName, newAPIKeyQuery(apiKeyName)}
}

func (s specAPIKeyQuerySecurityDefinition) getName() string {
	return s.Name
}

func (s specAPIKeyQuerySecurityDefinition) getType() securityDefinitionType {
	return securityDefinitionAPIKey
}

func (s specAPIKeyQuerySecurityDefinition) getAPIKey() specAPIKey {
	return s.apiKey
}

func (s specAPIKeyQuerySecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s specAPIKeyQuerySecurityDefinition) buildValue(value string) string {
	return value
}
