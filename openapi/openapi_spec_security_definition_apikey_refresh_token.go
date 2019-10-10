package openapi

import (
	"fmt"
	"strings"

	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
)

type specAPIKeyHeaderRefreshTokenSecurityDefinition struct {
	name            string
	refreshTokenURL string
}

// newAPIKeyHeaderRefreshTokenSecurityDefinition constructs a SpecSecurityDefinition of Header type using the Bearer authentication
// scheme. The secDefName value is the identifier of the security definition, and the refreshTokenURL is the URL that the openapi_spec_authenticator_refresh_token.go
func newAPIKeyHeaderRefreshTokenSecurityDefinition(secDefName string, refreshTokenURL string) specAPIKeyHeaderRefreshTokenSecurityDefinition {
	return specAPIKeyHeaderRefreshTokenSecurityDefinition{secDefName, refreshTokenURL}
}

func (s specAPIKeyHeaderRefreshTokenSecurityDefinition) getName() string {
	return s.name
}

func (s specAPIKeyHeaderRefreshTokenSecurityDefinition) getType() securityDefinitionType {
	return securityDefinitionAPIKeyRefreshToken
}

func (s specAPIKeyHeaderRefreshTokenSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(s.name)
}

func (s specAPIKeyHeaderRefreshTokenSecurityDefinition) getAPIKey() specAPIKey {
	apiKey := newAPIKeyHeader(authorizationHeader)
	apiKey.Metadata = map[apiKeyMetadataKey]interface{}{
		refreshTokenURLKey: s.refreshTokenURL,
	}
	return apiKey
}

func (s specAPIKeyHeaderRefreshTokenSecurityDefinition) buildValue(refreshToken string) string {
	if !strings.Contains(refreshToken, bearerScheme) {
		refreshToken = fmt.Sprintf("Bearer %s", refreshToken)
	}
	return refreshToken
}

func (s specAPIKeyHeaderRefreshTokenSecurityDefinition) validate() error {
	if s.name == "" {
		return fmt.Errorf("specAPIKeyHeaderRefreshTokenSecurityDefinition missing mandatory security definition name")
	}
	if s.refreshTokenURL == "" {
		return fmt.Errorf("specAPIKeyHeaderRefreshTokenSecurityDefinition missing mandatory refresh token URL")
	}
	isUrl := isUrl(s.refreshTokenURL)
	if !isUrl {
		return fmt.Errorf("refresh token URL must be a valid URL")
	}
	return nil
}
