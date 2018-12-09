package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

type specAPIKeyQueryBearerSecurityDefinition struct {
	Name string
}

// newAPIKeyHeaderSecurityDefinition constructs a SpecSecurityDefinition of Query type. The secDefName value is the identifier
// of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyQueryBearerSecurityDefinition(secDefName string) specAPIKeyQueryBearerSecurityDefinition {
	return specAPIKeyQueryBearerSecurityDefinition{secDefName}
}

func (s specAPIKeyQueryBearerSecurityDefinition) getName() string {
	return s.Name
}

func (s specAPIKeyQueryBearerSecurityDefinition) getType() securityDefinitionType {
	return securityDefinitionAPIKey
}

func (s specAPIKeyQueryBearerSecurityDefinition) getAPIKey() specAPIKey {
	return newAPIKeyQuery("access_token")
}

func (s specAPIKeyQueryBearerSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.Name)
}

func (s specAPIKeyQueryBearerSecurityDefinition) buildValue(value string) string {
	return value
}

func (s specAPIKeyQueryBearerSecurityDefinition) validate() error {
	if s.Name == "" {
		return fmt.Errorf("specAPIKeyQueryBearerSecurityDefinition missing mandatory security definition name")
	}
	return nil
}
