package openapi

// SpecBackendConfiguration defines the behaviour related to the OpenAPI doc backend configuration
type SpecBackendConfiguration interface {
	getHost() (string, error)
	getBasePath() string
	// TODO: integrate getHTTPSchemes2 wherever getHTTPSchemes is used; rename getHTTPSchemes2 afterwards to getHTTPScheme
	getHTTPSchemes2() (string, error)
	getHostByRegion(region string) (string, error)
	isMultiRegion() (bool, string, []string, error)
	getDefaultRegion([]string) (string, error)
}
