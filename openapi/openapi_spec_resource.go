package openapi

import (
	"time"
)

// SpecResource defines the behaviour related to terraform compliant OpenAPI Resources.
type SpecResource interface {
	getResourceName() string
	getHost() (string, error)
	getResourcePath(ids []string) (string, error)
	getResourceSchema() (*specSchemaDefinition, error)
	shouldIgnoreResource() bool
	getResourceOperations() specResourceOperations
	getTimeouts() (*specTimeouts, error)

	// TODO: Expand type SpecResource interface and add a new method like: isSubresource() bool. This helper method will
	// TODO: be used in multiple places to facilitate the identification of subresources. Subsequently expand SpecV2Resource,
	// TODO: which is an implementation of SpecResource, to incorporate an implementation of the aforementioned method.
	// TODO: The o.Path can be used to detect if the path is subresource or not. That can be done with a regex inspecting whether the path is parametrised.
	// TODO: o.Path contains always the ROOT path for the resource, and in the case of subresources the path should be parametrised
	// isSubResource() bool
}

type specTimeouts struct {
	Post   *time.Duration
	Get    *time.Duration
	Put    *time.Duration
	Delete *time.Duration
}
