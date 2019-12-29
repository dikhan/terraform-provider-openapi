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
	// getParentResourceInfo returns a struct populated with relevant parentResourceInfo if the resource is considered
	// a subresource; nil otherwise.
	getParentResourceInfo() *parentResourceInfo
}

type specTimeouts struct {
	Post   *time.Duration
	Get    *time.Duration
	Put    *time.Duration
	Delete *time.Duration
}
