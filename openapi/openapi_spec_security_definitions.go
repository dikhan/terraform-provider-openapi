package openapi

// SpecSecurityDefinitions groups a list of SpecSecurityDefinition
type SpecSecurityDefinitions []SpecSecurityDefinition

func (s SpecSecurityDefinitions) findSecurityDefinitionFor(securitySchemeName string) SpecSecurityDefinition {
	for _, securityDefinition := range s {
		if securityDefinition.getName() == securitySchemeName {
			return securityDefinition
		}
	}
	return nil
}

type securityDefinitionType string

const (
	securityDefinitionAPIKey securityDefinitionType = "apiKey"
)

// SpecSecurityDefinition defines the behaviour expected for security definition implementations. This interface creates
// an abstraction between the swagger security definitions and the openapi provider removing dependencies in external
// libraries
type SpecSecurityDefinition interface {
	// getName returns the name of the security scheme as defined in the swagger file
	getName() string
	// getType returns the security definition type, e,g: apiKey
	getType() securityDefinitionType
	// getTerraformConfigurationName returns the name converted terraform compliant name (snake_case) if needed
	getTerraformConfigurationName() string
	// getAPIKey returns the actual apiKey info containing the location of the key (e,g: header/query param) and the
	// name of the parameter used, in the case of a header the header name and in the case of a query parameter the query
	// parameter name
	getAPIKey() specAPIKey
	// buildValue accepts a value that then can be used to join with other values (e,g: auth schemes such as bearer)
	// to form the final value returned
	buildValue(value string) string
	// validate performs a check on the security definition to verify that it's well formed nad has the right mandatory configuration
	// including security definition name and any extra validation on the specAPIKey
	validate() error
}
