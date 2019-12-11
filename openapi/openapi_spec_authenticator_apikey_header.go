package openapi

import "fmt"

// Api Key Header Auth
type apiKeyHeaderAuthenticator struct {
	terraformConfigurationName string
	apiKey
}

func newAPIKeyHeaderAuthenticator(name, value, terraformConfigurationName string) apiKeyHeaderAuthenticator {
	return apiKeyHeaderAuthenticator{
		terraformConfigurationName: terraformConfigurationName,
		apiKey: apiKey{
			name:  name,
			value: value,
		},
	}
}

func (a apiKeyHeaderAuthenticator) getContext() interface{} {
	return a.apiKey
}

func (a apiKeyHeaderAuthenticator) getType() authType {
	return authTypeAPIKeyHeader
}

// prepareAPIKeyAuthentication adds to the map the auth header required for apikey header authentication. The url
// remains the same
func (a apiKeyHeaderAuthenticator) prepareAuth(authContext *authContext) error {
	apiKey := a.getContext().(apiKey)
	authContext.headers[apiKey.name] = apiKey.value
	return nil
}

func (a apiKeyHeaderAuthenticator) validate() error {
	if a.value == "" {
		return fmt.Errorf("required security definition '%s' is missing the value. Please make sure the property '%s' is configured with a value in the provider's terraform configuration", a.terraformConfigurationName, a.terraformConfigurationName)
	}
	return nil
}
