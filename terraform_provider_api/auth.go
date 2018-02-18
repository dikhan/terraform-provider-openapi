package main

import (
	"fmt"
	"github.com/go-openapi/spec"
	"log"
)

// authType is an enum defining the different types of authentication supported
type authType byte

const ( // iota is reset to 0
	authTypeAPIKeyHeader authType = iota
	authTypeAPIQuery
)

// OperationAuthenticator encapsulates both the operation for which the authenticator works as well
// as the authContext which will keep the state of the authorization (e,g" new headers added, url changed with
// query parameter containg a token, etc)
type OperationAuthenticator struct {
	authContext *authContext
	operation   *spec.Operation
}

type authContext struct {
	headers map[string]string
	url     string
}

// NewOperationAuthenticator allows for the creation of a new authenticator for a given operation
func NewOperationAuthenticator(op *spec.Operation, url string) OperationAuthenticator {
	return OperationAuthenticator{
		authContext: &authContext{
			headers: map[string]string{},
			url:     url,
		},
		operation: op,
	}
}

// Check if the operation contains any security policy. In the case where the operation contains multiple security
// policies, the first one found in the list will be the one returned.
// For more information about multiple api keys refer to https://swagger.io/docs/specification/authentication/api-keys/#multiple
func (oa OperationAuthenticator) authRequired() (bool, map[string][]string) {
	if len(oa.operation.Security) != 0 {
		log.Printf("operation %s contains security policies, selected the following based on order of appearance in the list %+v", oa.operation.ID, oa.operation.Security[0])
		return true, oa.operation.Security[0]
	}
	return false, nil
}

// Validate security policies. This function will perform two checks:
// 1. Verify that the operation security schemes are defined as security definitions in the provider config
func (oa OperationAuthenticator) confirmOperationSecurityPoliciesAreDefined(operationSecurityPolicies map[string][]string, providerConfig providerConfig) error {
	for operationSecurityDefName := range operationSecurityPolicies {
		securityDefinition := providerConfig.SecuritySchemaDefinitions[operationSecurityDefName]
		if securityDefinition == nil {
			return fmt.Errorf("operation's security policy %s is not defined, please make sure the swagger file contains a security definition named %s under the securityDefinitions section", operationSecurityDefName, operationSecurityDefName)
		}
	}
	return nil
}

func (oa OperationAuthenticator) prepareAuth(providerConfig providerConfig) (*authContext, error) {
	if required, operationSecurityPolicies := oa.authRequired(); required {
		if err := oa.confirmOperationSecurityPoliciesAreDefined(operationSecurityPolicies, providerConfig); err != nil {
			return oa.authContext, err
		}
		for securitySchemaDefinitionName := range operationSecurityPolicies {
			securitySchemaDefinition := providerConfig.SecuritySchemaDefinitions[securitySchemaDefinitionName]
			if err := securitySchemaDefinition.prepareAuth(oa.authContext); err != nil {
				return oa.authContext, err
			}
		}
	}
	return oa.authContext, nil
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
