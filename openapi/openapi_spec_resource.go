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
	isSubResource() (bool, error)
}

type specTimeouts struct {
	Post   *time.Duration
	Get    *time.Duration
	Put    *time.Duration
	Delete *time.Duration
}
