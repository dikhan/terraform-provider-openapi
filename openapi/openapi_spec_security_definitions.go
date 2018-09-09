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

type specAPIKey struct {
	In   string
	Name string
}

// SpecSecurityDefinition defines a security definition. This struct serves as a translation between the OpenAPI document
// and the scheme that will be used by the OpenAPI Terraform provider when making API calls to the backend
type SpecSecurityDefinition struct {
	Name   string
	Type   string // apiKey
	apiKey specAPIKey
}

func (o *SpecSecurityDefinition) getTerraformConfigurationName() string {
	return terraformutils.ConvertToTerraformCompliantName(o.Name)
}
