package openapi

import "github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"

// SpecSecurityDefinitions groups a list of SpecSecurityDefinition
type SpecSecurityDefinitions []SpecSecurityDefinition

func (s SpecSecurityDefinitions) findSecurityDefinitionFor(securitySchemeName string) *SpecSecurityDefinition {
	for _, securityDefinition := range s {
		if securityDefinition.Name == securitySchemeName {
			return &securityDefinition
		}
	}
	return nil
}

type apiKeyIn string

const (
	inHeader apiKeyIn = "header"
	inQuery  apiKeyIn = "query"
)

type specAPIKey struct {
	In   apiKeyIn
	Name string
}

func newAPIKeyHeader(name string) specAPIKey {
	return newAPIKey(name, inHeader)
}

func newAPIKeyQuery(name string) specAPIKey {
	return newAPIKey(name, inQuery)
}

func newAPIKey(name string, in apiKeyIn) specAPIKey {
	return specAPIKey{
		Name: name,
		In:   in,
	}
}

// SpecSecurityDefinition defines a security definition. This struct serves as a translation between the OpenAPI document
// and the scheme that will be used by the OpenAPI Terraform provider when making API calls to the backend
type SpecSecurityDefinition struct {
	Name   string
	Type   string // apiKey
	apiKey specAPIKey
}

// newAPIKeyHeaderSecurityDefinition constructs a SpecSecurityDefinition of Header type. The secDefName value is the identifier
// of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyHeaderSecurityDefinition(secDefName, apiKeyName string) SpecSecurityDefinition {
	return newAPIKeySecurityDefinition(secDefName, newAPIKeyHeader(apiKeyName))
}

// newAPIKeyHeaderSecurityDefinition constructs a SpecSecurityDefinition of Query type. The secDefName value is the identifier
// of the security definition, and the apiKeyName is the actual value of the header/query that will be user in the HTTP request.
func newAPIKeyQuerySecurityDefinition(secDefName, apiKeyName string) SpecSecurityDefinition {
	return newAPIKeySecurityDefinition(secDefName, newAPIKeyQuery(apiKeyName))
}

func newAPIKeySecurityDefinition(name string, apiKey specAPIKey) SpecSecurityDefinition {
	return SpecSecurityDefinition{
		Name:   name,
		Type:   "apiKey",
		apiKey: apiKey,
	}
}

func (o *SpecSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(o.Name)
}
