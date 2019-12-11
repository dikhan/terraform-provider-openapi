package openapi

// specAPIKeyAuthenticator defines the behaviour for api key type authenticators (e,g: header/query)
type specAPIKeyAuthenticator interface {
	getContext() interface{}
	prepareAuth(*authContext) error
	getType() authType
	validate() error
}

func createAPIKeyAuthenticator(secDef SpecSecurityDefinition, value string) specAPIKeyAuthenticator {
	switch secDef.getAPIKey().In {
	case inHeader:
		if secDef.getType() == securityDefinitionAPIKeyRefreshToken {
			return newAPIRefreshTokenAuthenticator(secDef.getAPIKey().Name, secDef.buildValue(value), secDef.getAPIKey().Metadata[refreshTokenURLKey].(string), secDef.getTerraformConfigurationName())
		}
		return newAPIKeyHeaderAuthenticator(secDef.getAPIKey().Name, secDef.buildValue(value), secDef.getTerraformConfigurationName())
	case inQuery:
		return newAPIKeyQueryAuthenticator(secDef.getAPIKey().Name, secDef.buildValue(value), secDef.getTerraformConfigurationName())
	}
	return nil
}

type apiKey struct {
	name  string
	value string
}
