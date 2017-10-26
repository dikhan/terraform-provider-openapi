package terraform_provider_api

import "github.com/go-openapi/spec"

type CrudResourcesInfo map[string]ResourceInfo

type ResourceInfo struct {
	Name             string
	SchemaDefinition spec.Schema
	// Path info contains info about /resource
	CreatePathInfo   spec.PathItem
	// Path info contains info about /resource/{id}
	PathInfo         spec.PathItem
}
