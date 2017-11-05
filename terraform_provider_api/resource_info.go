package terraform_provider_api

import "github.com/go-openapi/spec"

type CrudResourcesInfo map[string]ResourceInfo

type ResourceInfo struct {
	Name             string
	Path             string
	Host             string
	SchemaDefinition spec.Schema
	// CreatePathInfo contains info about /resource
	CreatePathInfo spec.PathItem
	// PathInfo contains info about /resource/{id}
	PathInfo spec.PathItem
}
