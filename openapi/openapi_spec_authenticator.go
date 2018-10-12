package openapi

// authType is an enum defining the different types of authentication supported
type authType byte

const ( // iota is reset to 0
	authTypeAPIKeyHeader authType = iota
	authTypeAPIQuery
)

type specAuthenticator interface {
	// prepareAuth generates an auth context with all the information regarding the authentication, including
	// any metadata that should be passed in to the request when making the http call to get a resource (e,g: new headers
	// with authentication details like access tokens, url with a query token, etc).
	// The following parameters describe the operationId for which the authentication is being prepared, the url of
	// the resource, the operation security schemes and the provider config containing the actual values like tokens,
	// special headers, etc for each security schemes
	prepareAuth(url string, operationSecuritySchemes SpecSecuritySchemes, providerConfig providerConfiguration) (*authContext, error)
}

type authContext struct {
	headers map[string]string
	url     string
}
