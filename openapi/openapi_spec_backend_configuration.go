package openapi

// SpecBackendConfiguration defines the behaviour related to the OpenAPI doc backend configuration
type SpecBackendConfiguration interface {
	getHost() (string, error)
	getBasePath() string
	getHTTPSchemes() []string
	getHTTPSchemes2() ([]string, error)
	getHostByRegion(region string) (string, error)
	isMultiRegion() (bool, string, []string, error)
	getDefaultRegion([]string) (string, error)
}
