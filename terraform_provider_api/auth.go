package main

import (
	"fmt"
	"log"
)

// authType is an enum defining the different types of authentication supported
type authType byte

const ( // iota is reset to 0
	authTypeAPIKeyHeader authType = iota
	authTypeAPIQuery
)

type apiAuthenticator interface {
	// prepareAuth generates an auth context with all the information regarding the authentication, including
	// any metadata that should be passed in to the request when making the http call to get a resource (e,g: new headers
	// with authentication details like access tokens, url with a query token, etc).
	// The following parameters describe the operationId for which the authentication is being prepared, the url of
	// the resource, the operation security schemes and the provider config containing the actual values like tokens,
	// special headers, etc for each security schemes
	prepareAuth(operationID, url string, operationSecuritySchemes []map[string][]string, providerConfig providerConfig) (*authContext, error)
}

// apiAuth is an implementation of apiAuthenticator encapsulating the general settings to be applied in case
// an operation does not contain a security policy; otherwise the operation's security policies will be applied instead.
type apiAuth struct {
	globalSecuritySchemes []map[string][]string
}

type authContext struct {
	headers map[string]string
	url     string
}

// newAPIAuthenticator allows for the creation of a new authenticator for a given operation
func newAPIAuthenticator(globalSecuritySchemes []map[string][]string) apiAuthenticator {
	return apiAuth{
		globalSecuritySchemes: globalSecuritySchemes,
	}
}

// Check if the operation contains any security policy. In the case where the operation contains multiple security
// policies, the first one found in the list will be the one returned.
// For more information about multiple api keys refer to https://swagger.io/docs/specification/authentication/api-keys/#multiple
func (oa apiAuth) authRequired(operationID, url string, operationSecuritySchemes []map[string][]string) (bool, map[string][]string) {
	if len(operationSecuritySchemes) != 0 {
		log.Printf("operation %s [%s] contains security policies (overriding global security config if applicable), selected the following based on order of appearance in the list %+v", operationID, url, operationSecuritySchemes[0])
		return true, operationSecuritySchemes[0]
	}
	log.Printf("operation %s [%s] is missing specific security scheme, falling back to global security schemes (if there's any)", operationID, url)
	if len(oa.globalSecuritySchemes) != 0 {
		log.Printf("the global configuration contains security schemes, selected the following based on order of appearance in the list %+v", oa.globalSecuritySchemes[0])
		return true, oa.globalSecuritySchemes[0]
	}
	return false, nil
}

// Validate security policies. This function will perform the following checks:
// 1. Verify that the operation security schemes are defined as security definitions in the provider config
func (oa apiAuth) confirmOperationSecurityPoliciesAreDefined(operationSecurityPolicies map[string][]string, providerConfig providerConfig) error {
	for operationSecurityDefName := range operationSecurityPolicies {
		securityDefinition := providerConfig.SecuritySchemaDefinitions[operationSecurityDefName]
		if securityDefinition == nil {
			return fmt.Errorf("operation's security policy %s is not defined, please make sure the swagger file contains a security definition named %s under the securityDefinitions section", operationSecurityDefName, operationSecurityDefName)
		}
	}
	return nil
}

func (oa apiAuth) prepareAuth(operationID, url string, operationSecuritySchemes []map[string][]string, providerConfig providerConfig) (*authContext, error) {
	authContext := &authContext{
		headers: map[string]string{},
		url:     url,
	}
	if required, operationSecurityPolicies := oa.authRequired(operationID, url, operationSecuritySchemes); required {
		if err := oa.confirmOperationSecurityPoliciesAreDefined(operationSecurityPolicies, providerConfig); err != nil {
			return authContext, err
		}
		for securitySchemaDefinitionName := range operationSecurityPolicies {
			securitySchemaDefinition := providerConfig.SecuritySchemaDefinitions[securitySchemaDefinitionName]
			if err := securitySchemaDefinition.prepareAuth(authContext); err != nil {
				return authContext, err
			}
		}
	}
	return authContext, nil
}

type authenticator interface {
	getContext() interface{}
	prepareAuth(*authContext) error
	getType() authType
}

type apiKey struct {
	name  string
	value string
}

// Api Key Header Auth
type apiKeyHeader struct {
	apiKey
}

func (a apiKeyHeader) getContext() interface{} {
	return a.apiKey
}

func (a apiKeyHeader) getType() authType {
	return authTypeAPIKeyHeader
}

// prepareAPIKeyAuthentication adds to the map the auth header required for apikey header authentication. The url
// remains the same
func (a apiKeyHeader) prepareAuth(authContext *authContext) error {
	apiKey := a.getContext().(apiKey)
	authContext.headers[apiKey.name] = apiKey.value
	return nil
}

// Api Key Query Auth
type apiKeyQuery struct {
	apiKey
}

func (a apiKeyQuery) getContext() interface{} {
	return a.apiKey
}

func (a apiKeyQuery) getType() authType {
	return authTypeAPIQuery
}

// prepareAPIKeyAuthentication updates the url to insert the query api auth values. The map returned is not
// populated in this case as the auth is done via query parameters. However, having the ability to return the map
// provides the opportunity to inject some headers if needed.
func (a apiKeyQuery) prepareAuth(authContext *authContext) error {
	apiKey := a.getContext().(apiKey)
	authContext.url = fmt.Sprintf("%s?%s=%s", authContext.url, apiKey.name, apiKey.value)
	return nil
}

func createAPIKeyAuthenticator(apiKeyAuthType, name, value string) authenticator {
	switch apiKeyAuthType {
	case "header":
		return apiKeyHeader{apiKey{name, value}}
	case "query":
		return apiKeyQuery{apiKey{name, value}}
	}
	return nil
}
