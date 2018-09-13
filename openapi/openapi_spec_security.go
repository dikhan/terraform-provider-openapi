package openapi

// SpecSecurity defines the behaviour related to OpenAPI security.
// This interface serves as a translation between the OpenAPI document and the security spec that will be used by the
// OpenAPI Terraform provider
type SpecSecurity interface {
	// GetAPIKeySecurityDefinitions returns all the OpenAPI security definitions from the OpenAPI document and translates those
	// into SpecSecurityDefinitions
	GetAPIKeySecurityDefinitions() SpecSecurityDefinitions
	// GetGlobalSecuritySchemes returns all the global security schemes from the OpenAPI document and translates those
	// into SpecSecuritySchemes
	GetGlobalSecuritySchemes() (SpecSecuritySchemes, error)
}
