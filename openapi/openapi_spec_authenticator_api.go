package openapi

import (
	"fmt"
	"log"
)

// apiAuth is an implementation of specAuthenticator encapsulating the general settings to be applied in case
// an operation does not contain a security policy; otherwise the operation's security policies will be applied instead.
type apiAuth struct {
	globalSecuritySchemes *SpecSecuritySchemes
}

// newAPIAuthenticator allows for the creation of a new authenticator
func newAPIAuthenticator(globalSecuritySchemes *SpecSecuritySchemes) specAuthenticator {
	return apiAuth{
		globalSecuritySchemes: globalSecuritySchemes,
	}
}

// Check if the operation contains any security policy. In the case where the operation contains multiple security
// policies, the first one found in the list will be the one returned.
// For more information about multiple api keys refer to https://swagger.io/docs/specification/authentication/api-keys/#multiple
func (oa apiAuth) authRequired(url string, operationSecuritySchemes SpecSecuritySchemes) (bool, SpecSecuritySchemes) {
	// TODO: check in the OpenAPI spec whether operation overrides global schemes or can complement global configuration?
	if len(operationSecuritySchemes) != 0 {
		log.Printf("operation security policies found for '%s' (overriding global security config if applicable). Selected the following based on order of appearance in the list %+v", url, operationSecuritySchemes)
		return true, operationSecuritySchemes
	}
	log.Printf("operation security schemes missing, falling back to global security schemes (if there's any)")
	if oa.globalSecuritySchemes != nil && len(*oa.globalSecuritySchemes) != 0 {
		log.Printf("the global configuration contains security schemes, selected the following based on order of appearance in the list %+v", oa.globalSecuritySchemes)
		return true, *oa.globalSecuritySchemes
	}
	return false, nil
}

// Validate security policies. This function will perform the following checks:
// 1. Verify that the operation security schemes are defined as security definitions in the provider config
func (oa apiAuth) fetchRequiredAuthenticators(operationSecuritySchemes SpecSecuritySchemes, providerConfig providerConfiguration) ([]specAPIKeyAuthenticator, error) {
	var authenticators []specAPIKeyAuthenticator
	for _, operationSecurityScheme := range operationSecuritySchemes {
		authenticator := providerConfig.getAuthenticatorFor(operationSecurityScheme)
		if authenticator == nil {
			return nil, fmt.Errorf("operation's security policy '%s' is not defined, please make sure the swagger file contains a security definition named '%s' under the securityDefinitions section", operationSecurityScheme, operationSecurityScheme)
		}
		authenticators = append(authenticators, authenticator)
	}
	return authenticators, nil
}

func (oa apiAuth) prepareAuth(url string, operationSecuritySchemes SpecSecuritySchemes, providerConfig providerConfiguration) (*authContext, error) {
	authContext := &authContext{
		headers: map[string]string{},
		url:     url,
	}
	if required, requiredSecuritySchemes := oa.authRequired(url, operationSecuritySchemes); required {
		authenticators, err := oa.fetchRequiredAuthenticators(requiredSecuritySchemes, providerConfig)
		if err != nil {
			return authContext, err
		}
		for _, authenticator := range authenticators {
			if err := authenticator.prepareAuth(authContext); err != nil {
				return authContext, err
			}
		}
	}
	return authContext, nil
}
