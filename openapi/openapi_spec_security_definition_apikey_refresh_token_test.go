package openapi

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAPIKeyHeaderRefreshTokenSecurityDefinition(t *testing.T) {
	Convey("Given a name and an refresh token URL", t, func() {
		name := "apikey_auth"
		refreshTokenURL := "https://api.iam.com/token"
		Convey("When newAPIKeyHeaderRefreshTokenSecurityDefinition method is called", func() {
			apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition(name, refreshTokenURL)
			Convey("Then the apiKeyHeaderAuthenticator should comply with SpecSecurityDefinition interface", func() {
				var _ SpecSecurityDefinition = apiKeyHeaderRefreshTokenSecurityDefinition
			})
		})
	})
}

func TestAPIKeyHeaderRefreshTokenSecurityDefinitionGetType(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("apikey_auth", "https://api.iam.com/token")
		Convey("When getType method is called", func() {
			secDefType := apiKeyHeaderRefreshTokenSecurityDefinition.getType()
			Convey("Then the result should be securityDefinitionAPIKeyRefreshToken", func() {
				So(secDefType, ShouldEqual, securityDefinitionAPIKeyRefreshToken)
			})
		})
	})
}

func TestAPIKeyHeaderRefreshTokenSecurityDefinitionGetTerraformConfigurationName(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition with a compliant name", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("apikey_auth", "https://api.iam.com/token")
		Convey("When getTerraformConfigurationName method is called", func() {
			secDefTfName := apiKeyHeaderRefreshTokenSecurityDefinition.getTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "apikey_auth")
			})
		})
	})

	Convey("Given an APIKeyHeaderSecurityDefinition with a NON compliant name", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("nonCompliantName", "https://api.iam.com/token")
		Convey("When getTerraformConfigurationName method is called", func() {
			secDefTfName := apiKeyHeaderRefreshTokenSecurityDefinition.getTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "non_compliant_name")
			})
		})
	})
}

func TestAPIKeyHeaderRefreshTokenSecurityDefinitionGetAPIKey(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("apikey_auth", "https://api.iam.com/token")
		Convey("When getTerraformConfigurationName method is called", func() {
			apiKey := apiKeyHeaderRefreshTokenSecurityDefinition.getAPIKey()
			Convey("Then the result should contain the right apikey name and location", func() {
				So(apiKey.Name, ShouldEqual, "Authorization")
				So(apiKey.In, ShouldEqual, inHeader)
				So(apiKey.Metadata, ShouldNotBeEmpty)
				So(apiKey.Metadata[refreshTokenURLKey], ShouldEqual, "https://api.iam.com/token")
			})
		})
	})
}

func TestAPIKeyHeaderRefreshTokenSecurityDefinitionBuildValue(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("apikey_auth", "https://api.iam.com/token")
		Convey("When buildValue method is called with a token that does not contain the bearer scheme", func() {
			expectedValue := "jwtRefreshToken"
			returnedValue := apiKeyHeaderRefreshTokenSecurityDefinition.buildValue(expectedValue)
			Convey("Then the expectedValue should be the expected expectedValue with Bearer included", func() {
				So(returnedValue, ShouldEqual, fmt.Sprintf("Bearer %s", expectedValue))
			})
		})
		Convey("When buildValue method is called with a token that does contain the bearer scheme", func() {
			expectedValue := "Bearer jwtRefreshToken"
			returnedValue := apiKeyHeaderRefreshTokenSecurityDefinition.buildValue(expectedValue)
			Convey("Then the expectedValue should be the expected expectedValue with Bearer included", func() {
				So(returnedValue, ShouldEqual, expectedValue)
			})
		})
	})
}

func TestAPIKeyHeaderRefreshTokenSecurityDefinitionValidate(t *testing.T) {
	Convey("Given an APIKeyQueryBearerSecurityDefinition with a security definition name and a valid refresh token url", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("apikey_auth", "https://api.iam.com/token")
		Convey("When validate method is called", func() {
			err := apiKeyHeaderRefreshTokenSecurityDefinition.validate()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
	Convey("Given an APIKeyQueryBearerSecurityDefinition with an empty security definition name and a valid refresh token URL", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("", "https://api.iam.com/token")
		Convey("When validate method is called", func() {
			err := apiKeyHeaderRefreshTokenSecurityDefinition.validate()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message should match the expected", func() {
				So(err.Error(), ShouldEqual, "specAPIKeyHeaderRefreshTokenSecurityDefinition missing mandatory security definition name")
			})
		})
	})
	Convey("Given an APIKeyQueryBearerSecurityDefinition with a security definition name and an valid refresh token URL", t, func() {
		apiKeyHeaderRefreshTokenSecurityDefinition := newAPIKeyHeaderRefreshTokenSecurityDefinition("apikey_auth", "api.iam.com/token")
		Convey("When validate method is called", func() {
			err := apiKeyHeaderRefreshTokenSecurityDefinition.validate()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message should match the expected", func() {
				So(err.Error(), ShouldEqual, "refresh token URL must be a valid URL")
			})
		})
	})
}
