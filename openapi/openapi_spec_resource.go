package openapi

import (
	"time"
)

// SpecResource defines the behaviour related to terraform compliant OpenAPI Resources.
type SpecResource interface {
	getResourceName() string
	getHost() (string, error)
	getResourcePath(parentIDs []string) (string, error)
	getResourceSchema() (*specSchemaDefinition, error)
	shouldIgnoreResource() bool
	getResourceOperations() specResourceOperations
	getTimeouts() (*specTimeouts, error)
	// isSubResource returns true if the resource path is a subresource. Additionally, it will return the list of parent
	// resource names and the resource parent names merged in one to facilitate parent names processing.
	isSubResource() (bool, []string, string)
	// getParentPropertiesNames is responsible to building the parent properties names for a resource that is a subresource
	getParentPropertiesNames() ([]string, error)
}

type specTimeouts struct {
	Post   *time.Duration
	Get    *time.Duration
	Put    *time.Duration
	Delete *time.Duration
}
