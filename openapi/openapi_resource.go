package openapi

import "github.com/hashicorp/terraform/helper/schema"

type OpenApiResource interface {
	getResourceName() string
	getResourcePath() string
	createResourceSchema() (map[string]*schema.Schema, error)
}
