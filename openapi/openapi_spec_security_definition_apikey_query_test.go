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
	Convey("Given an APIKeyQuerySecurityDefinition", t, func() {
		expectedName := "apikey_name"
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition(expectedName, authorizationHeader)
		Convey("When GetTerraformConfigurationName method is called", func() {
			name := apiKeyQuerySecurityDefinition.getName()
			Convey("Then the result should match the original name", func() {
				So(name, ShouldEqual, expectedName)
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionGetType(t *testing.T) {
	Convey("Given an APIKeyQuerySecurityDefinition", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apikey_name", authorizationHeader)
		Convey("When getType method is called", func() {
			secDefType := apiKeyQuerySecurityDefinition.getType()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefType, ShouldEqual, securityDefinitionAPIKey)
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionGetTerraformConfigurationName(t *testing.T) {
	Convey("Given an APIKeyQuerySecurityDefinition with a compliant name", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apikey_name", authorizationHeader)
		Convey("When GetTerraformConfigurationName method is called", func() {
			secDefTfName := apiKeyQuerySecurityDefinition.GetTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "apikey_name")
			})
		})
	})

	Convey("Given an APIKeyQuerySecurityDefinition with a NON compliant name", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("nonCompliantName", authorizationHeader)
		Convey("When GetTerraformConfigurationName method is called", func() {
			secDefTfName := apiKeyQuerySecurityDefinition.GetTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "non_compliant_name")
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionGetAPIKey(t *testing.T) {
	Convey("Given an APIKeyQuerySecurityDefinition", t, func() {
		expectedAPIKey := authorizationHeader
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apiKeyName", expectedAPIKey)
		Convey("When GetTerraformConfigurationName method is called", func() {
			apiKey := apiKeyQuerySecurityDefinition.getAPIKey()
			Convey("Then the result should contain the right apikey name and location", func() {
				So(apiKey.Name, ShouldEqual, expectedAPIKey)
				So(apiKey.In, ShouldEqual, inQuery)
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionBuildValue(t *testing.T) {
	Convey("Given an APIKeyQuerySecurityDefinition", t, func() {
		expectedAPIKey := authorizationHeader
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apiKeyName", expectedAPIKey)
		Convey("When GetTerraformConfigurationName method is called", func() {
			expectedValue := "someValue"
			value := apiKeyQuerySecurityDefinition.buildValue("someValue")
			Convey("Then the value should be the expected value with no modifications", func() {
				So(value, ShouldEqual, expectedValue)
			})
		})
	})
}

func TestAPIKeyQuerySecurityDefinitionValidate(t *testing.T) {
	Convey("Given an APIKeyQuerySecurityDefinition with a security definition name and an apiKeyName", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apiKeyName", authorizationHeader)
		Convey("When validate method is called", func() {
			err := apiKeyQuerySecurityDefinition.validate()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
	Convey("Given an APIKeyQuerySecurityDefinition with an empty security definition name and an apiKeyName", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("", authorizationHeader)
		Convey("When validate method is called", func() {
			err := apiKeyQuerySecurityDefinition.validate()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "specAPIKeyQuerySecurityDefinition missing mandatory security definition name")
			})
		})
	})
	Convey("Given an APIKeyQuerySecurityDefinition with a security definition name and an empty apiKeyName", t, func() {
		apiKeyQuerySecurityDefinition := newAPIKeyQuerySecurityDefinition("apiKeyName", "")
		Convey("When validate method is called", func() {
			err := apiKeyQuerySecurityDefinition.validate()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "specAPIKeyQuerySecurityDefinition missing mandatory apiKey name")
			})
		})
	})
}
