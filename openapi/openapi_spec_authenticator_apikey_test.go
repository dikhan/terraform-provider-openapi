package openapi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateAPIKeyAuthenticator(t *testing.T) {
	testCases := []struct {
		name                    string
		secDef                  SpecSecurityDefinition
		value                   string
		expectedAuthType        specAPIKeyAuthenticator
		expectedType            authType
		expectedValidationError error
	}{
		{
			name:                    "createAPIKeyAuthenticator is called with a valid specAPIKeyHeaderSecurityDefinition and a value",
			secDef:                  newAPIKeyHeaderSecurityDefinition("header_auth", authorizationHeader),
			value:                   "value",
			expectedAuthType:        apiKeyHeaderAuthenticator{},
			expectedType:            authTypeAPIKeyHeader,
			expectedValidationError: nil,
		},
		{
			name:                    "createAPIKeyAuthenticator is called with a valid specAPIKeyQuerySecurityDefinition and a value",
			secDef:                  newAPIKeyQuerySecurityDefinition("query_auth", authorizationHeader),
			value:                   "value",
			expectedAuthType:        apiKeyQueryAuthenticator{},
			expectedType:            authTypeAPIQuery,
			expectedValidationError: nil,
		},
		{
			name:                    "createAPIKeyAuthenticator is called with a valid specAPIKeyHeaderRefreshTokenSecurityDefinition and a value",
			secDef:                  newAPIKeyHeaderRefreshTokenSecurityDefinition("header_auth", authorizationHeader),
			value:                   "value",
			expectedAuthType:        apiRefreshTokenAuthenticator{},
			expectedType:            authTypeAPIKeyHeader,
			expectedValidationError: nil,
		},
	}

	for _, tc := range testCases {
		apiKeyAuthenticator := createAPIKeyAuthenticator(tc.secDef, tc.value)
		assert.IsType(t, tc.expectedAuthType, apiKeyAuthenticator, tc.name)
		assert.Equal(t, tc.expectedType, apiKeyAuthenticator.getType(), tc.name)
		assert.Equal(t, tc.expectedValidationError, apiKeyAuthenticator.validate(), tc.name)
	}
}
