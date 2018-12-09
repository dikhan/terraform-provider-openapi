package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAPIKeyQueryBearerSecurityDefinition(t *testing.T) {
	Convey("Given a security definition name", t, func() {
		name := "sec_def_name"
		Convey("When newAPIKeyQueryBearerSecurityDefinition method is called", func() {
			specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition(name)
			Convey("Then the apiKeyHeaderAuthenticator should comply with specAPIKeyAuthenticator interface", func() {
				var _ SpecSecurityDefinition = specAPIKeyQueryBearerSecurityDefinition
			})
		})
	})
}

func TestAPIKeyQueryBearerSecurityDefinitionGetName(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		expectedName := "apikey_name"
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition(expectedName)
		Convey("When getTerraformConfigurationName method is called", func() {
			name := specAPIKeyQueryBearerSecurityDefinition.getName()
			Convey("Then the result should match the original name", func() {
				So(name, ShouldEqual, expectedName)
			})
		})
	})
}

func TestAPIKeyQueryBearerSecurityDefinitionGetType(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition("apikey_name")
		Convey("When getType method is called", func() {
			secDefType := specAPIKeyQueryBearerSecurityDefinition.getType()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefType, ShouldEqual, securityDefinitionAPIKey)
			})
		})
	})
}

func TestAPIKeyQueryBearerSecurityDefinitionGetTerraformConfigurationName(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition with a compliant name", t, func() {
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition("apikey_name")
		Convey("When getTerraformConfigurationName method is called", func() {
			secDefTfName := specAPIKeyQueryBearerSecurityDefinition.getTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "apikey_name")
			})
		})
	})

	Convey("Given an APIKeyHeaderSecurityDefinition with a NON compliant name", t, func() {
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition("nonCompliantName")
		Convey("When getTerraformConfigurationName method is called", func() {
			secDefTfName := specAPIKeyQueryBearerSecurityDefinition.getTerraformConfigurationName()
			Convey("Then the result should be securityDefinitionAPIKey", func() {
				So(secDefTfName, ShouldEqual, "non_compliant_name")
			})
		})
	})
}

func TestAPIKeyQueryBearerSecurityDefinitionGetAPIKey(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition("apikey_name")
		Convey("When getTerraformConfigurationName method is called", func() {
			apiKey := specAPIKeyQueryBearerSecurityDefinition.getAPIKey()
			Convey("Then the result should contain the right apikey name and location", func() {
				So(apiKey.Name, ShouldEqual, "access_token")
				So(apiKey.In, ShouldEqual, inQuery)
			})
		})
	})
}

func TestAPIKeyQueryBearerSecurityDefinitionBuildValue(t *testing.T) {
	Convey("Given an APIKeyHeaderSecurityDefinition", t, func() {
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition("apikey_name")
		Convey("When getTerraformConfigurationName method is called", func() {
			expectedValue := "jwtToken"
			returnedValue := specAPIKeyQueryBearerSecurityDefinition.buildValue(expectedValue)
			Convey("Then the expectedValue should be the expected expectedValue with Bearer included", func() {
				So(returnedValue, ShouldEqual, expectedValue)
			})
		})
	})
}

func TestAPIKeyQueryBearerSecurityDefinitionValidate(t *testing.T) {
	Convey("Given an APIKeyQueryBearerSecurityDefinition with a security definition name", t, func() {
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition("apikey_name")
		Convey("When validate method is called", func() {
			err := specAPIKeyQueryBearerSecurityDefinition.validate()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
	Convey("Given an APIKeyQueryBearerSecurityDefinition with a security definition name", t, func() {
		specAPIKeyQueryBearerSecurityDefinition := newAPIKeyQueryBearerSecurityDefinition("")
		Convey("When validate method is called", func() {
			err := specAPIKeyQueryBearerSecurityDefinition.validate()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message should match the expected", func() {
				So(err.Error(), ShouldEqual, "specAPIKeyQueryBearerSecurityDefinition missing mandatory security definition name")
			})
		})
	})
}
