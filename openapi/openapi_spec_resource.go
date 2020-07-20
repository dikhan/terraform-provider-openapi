package openapi

import (
	"time"
)

// SpecResource defines the behaviour related to terraform compliant OpenAPI Resources.
type SpecResource interface {
	GetResourceName() string
	getHost() (string, error)
	getResourcePath(parentIDs []string) (string, error)
	GetResourceSchema() (*SpecSchemaDefinition, error)
	ShouldIgnoreResource() bool
	getResourceOperations() specResourceOperations
	getTimeouts() (*specTimeouts, error)
	// GetParentResourceInfo returns a struct populated with relevant ParentResourceInfo if the resource is considered
	// a sub-resource; nil otherwise.
	GetParentResourceInfo() *ParentResourceInfo
}

type specTimeouts struct {
	Post   *time.Duration
	Get    *time.Duration
	Put    *time.Duration
	Delete *time.Duration
}
