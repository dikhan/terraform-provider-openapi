package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

// specAPIKeyHeaderSecurityDefinition defines a security definition. This struct serves as a translation between the OpenAPI document
// and the scheme that will be used by the OpenAPI Terraform provider when making API calls to the backend
type specAPIKeyHeaderSecurityDefinition struct {
	Name   string
	apiKey specAPIKey
}

// newAPIKeyHeaderSecurityDefinition constructs a SpecSecurityDefinition of Header type. The secDefName value is the identifier
// of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyHeaderSecurityDefinition(secDefName, apiKeyName string) specAPIKeyHeaderSecurityDefinition {
	return specAPIKeyHeaderSecurityDefinition{secDefName, newAPIKeyHeader(apiKeyName)}
}

func (s specAPIKeyHeaderSecurityDefinition) getName() string {
	return s.Name
}

func (s specAPIKeyHeaderSecurityDefinition) getType() securityDefinitionType {
	return securityDefinitionAPIKey
}

func (s specAPIKeyHeaderSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s specAPIKeyHeaderSecurityDefinition) getAPIKey() specAPIKey {
	return s.apiKey
}

func (s specAPIKeyHeaderSecurityDefinition) buildValue(value string) string {
	return value
}
