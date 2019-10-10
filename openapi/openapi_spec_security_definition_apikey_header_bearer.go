package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"strings"
)

const bearerScheme = "Bearer"

type specAPIKeyHeaderBearerSecurityDefinition struct {
	name string
}

// newAPIKeyHeaderBearerSecurityDefinition constructs a SpecSecurityDefinition of Header type using the Bearer authentication
// scheme. The secDefName value is the identifier of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyHeaderBearerSecurityDefinition(secDefName string) specAPIKeyHeaderBearerSecurityDefinition {
	return specAPIKeyHeaderBearerSecurityDefinition{secDefName}
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getName() string {
	return s.name
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getType() securityDefinitionType {
	return securityDefinitionAPIKey
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.name)
}

func (s specAPIKeyHeaderBearerSecurityDefinition) getAPIKey() specAPIKey {
	return newAPIKeyHeader(authorizationHeader)
}

func (s specAPIKeyHeaderBearerSecurityDefinition) buildValue(value string) string {
	if strings.Contains(value, bearerScheme) {
		return value
	}
	return fmt.Sprintf("Bearer %s", value)
}

func (s specAPIKeyHeaderBearerSecurityDefinition) validate() error {
	if s.name == "" {
		return fmt.Errorf("specAPIKeyHeaderBearerSecurityDefinition missing mandatory security definition name")
	}
	return nil
}
