package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

type specAPIKeyHeaderBearerSecurityDefinition struct {
	Name string
}

// newAPIKeyHeaderBearerSecurityDefinition constructs a SpecSecurityDefinition of Header type using the Bearer authentication
// scheme. The secDefName value is the identifier of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyHeaderBearerSecurityDefinition(secDefName string) specAPIKeyHeaderBearerSecurityDefinition {
	return specAPIKeyHeaderBearerSecurityDefinition{secDefName}
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getName() string {
	return s.Name
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getType() securityDefinitionType {
	return securityDefinitionAPIKey
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getAPIKey() specAPIKey {
	return newAPIKeyHeader("Authorization")
}

func (s specAPIKeyHeaderBearerSecurityDefinition) buildValue(value string) string {
	return fmt.Sprintf("Bearer %s", value)
}
