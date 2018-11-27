package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAPIKeyQuerySecurityDefinition(t *testing.T) {
	Convey("Given a name and an apikey name", t, func() {
		name := "name"
		apiKeyName := "apiKey_name"
		Convey("When newAPIKeyQuerySecurityDefinition method is called", func() {
			apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition(name, apiKeyName)
			Convey("Then the apiKeyHeaderAuthenticator should comply with specAPIKeyAuthenticator interface", func() {
				var _ SpecSecurityDefinition = apiKeyQuerySecurityDefinition
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionGetName(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		expectedName := "apikey_name"
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition(expectedName, "Authorization")
		Convey("When getTerraformConfigurationName method is called", func() {
			name := apiKeyQuerySecurityDefinition.getName()
			Convey("Then the result should match the original name", func() {
				So(name, ShouldEqual, expectedName)
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionGetType(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apikey_name", "Authorization")
		Convey("When getType method is called", func() {
			secDefType := apiKeyQuerySecurityDefinition.getType()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefType, ShouldEqual, securityDefinitionAPIKey)
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionGetTerraformConfigurationName(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition with a compliant name", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apikey_name", "Authorization")
		Convey("When getTerraformConfigurationName method is called", func() {
			secDefTfName := apiKeyQuerySecurityDefinition.getTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "apikey_name")
			})
		})
	})

	Convey("Given an APIKeyHeaderSecurityDefinition with a NON compliant name", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("nonCompliantName", "Authorization")
		Convey("When getTerraformConfigurationName method is called", func() {
			secDefTfName := apiKeyQuerySecurityDefinition.getTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "non_compliant_name")
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionGetAPIKey(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		expectedAPIKey := "Authorization"
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apiKeyName", expectedAPIKey)
		Convey("When getTerraformConfigurationName method is called", func() {
			apiKey := apiKeyQuerySecurityDefinition.getAPIKey()
			Convey("Then the result should contain the right apikey name and location", func() {
				So(apiKey.Name, ShouldEqual, expectedAPIKey)
				So(apiKey.In, ShouldEqual, inQuery)
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionBuildValue(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		expectedAPIKey := "Authorization"
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apiKeyName", expectedAPIKey)
		Convey("When getTerraformConfigurationName method is called", func() {
			expectedValue := "someValue"
			value := apiKeyQuerySecurityDefinition.buildValue("someValue")
			Convey("Then the value should be the expected value with no modifications", func() {
				So(value, ShouldEqual, expectedValue)
			})
		})
	})
}
