package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/go-openapi/spec"
	"strings"
)

type specAPIKeyHeaderCustomAuthSecurityDefinition struct {
	name            string
	secDef          *spec.SecurityScheme
	refreshTokenURL string
}

// newAPIKeyHeaderBearerSecurityDefinition constructs a SpecSecurityDefinition of Header type using the Bearer authentication
// scheme. The secDefName value is the identifier of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyHeaderRefreshTokenSecurityDefinition(secDefName string, secDef *spec.SecurityScheme, refreshTokenURL string) specAPIKeyHeaderCustomAuthSecurityDefinition {
	return specAPIKeyHeaderCustomAuthSecurityDefinition{secDefName, secDef, refreshTokenURL}
}

func (s specAPIKeyHeaderCustomAuthSecurityDefinition) getName() string {
	return s.name
}

func (s specAPIKeyHeaderCustomAuthSecurityDefinition) getType() securityDefinitionType {
	return securityDefinitionAPIKeyRefreshToken
}

func (s specAPIKeyHeaderCustomAuthSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.name)
}

func (s specAPIKeyHeaderCustomAuthSecurityDefinition) getAPIKey() specAPIKey {
	apiKey := newAPIKeyHeader(authorization)
	apiKey.Metadata = map[apiKeyMetadataKey]interface{}{
		refreshTokenURLKey: s.refreshTokenURL,
	}
	return apiKey
}

func (s specAPIKeyHeaderCustomAuthSecurityDefinition) buildValue(refreshToken string) string {
	if !strings.Contains(refreshToken, bearerScheme) {
		refreshToken = fmt.Sprintf("Bearer %s", refreshToken)
	}
	return refreshToken
}

func (s specAPIKeyHeaderCustomAuthSecurityDefinition) validate() error {
	if s.name == "" {
		return fmt.Errorf("specAPIKeyHeaderBearerSecurityDefinition missing mandatory security definition name")
	}
	return nil
}
