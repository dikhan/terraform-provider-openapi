package openapi

// SpecBackendConfiguration defines the behaviour related to the OpenAPI doc backend configuration
type SpecBackendConfiguration interface {
	getHost() string
	getBasePath() string
	getHTTPSchemes() []string
}
