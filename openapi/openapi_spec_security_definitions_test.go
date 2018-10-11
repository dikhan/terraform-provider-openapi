package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAPIKeySecurityDefinition(t *testing.T) {
	Convey("Given a secDefName and an apiKey", t, func() {
		secDefName := "api_header_auth"
		specAPIKey := specAPIKey{
			Name: "apiKeyName",
			In: "somewhere",
		}
		Convey("When newAPIKeySecurityDefinition method is called", func() {
			secDef := newAPIKeySecurityDefinition(secDefName, specAPIKey)
			Convey("Then the sec def name should match", func() {
				So(secDef.Name, ShouldEqual, secDefName)
			})
			Convey("And the sec def type should be apiKey", func() {
				So(secDef.Type, ShouldEqual, "apiKey")
			})
			Convey("And the sec def apikey should contain the apikey name and the right IN inHeader value", func() {
				So(secDef.apiKey.Name, ShouldEqual, "apiKeyName")
				So(secDef.apiKey.In, ShouldEqual, "somewhere")
			})
		})
	})
}

func TestNewAPIKeyHeaderSecurityDefinition(t *testing.T) {
	Convey("Given a secDefName and an apiKeyName  ", t, func() {
		secDefName := "api_header_auth"
		apiKeyName := "Authorization"
		Convey("When newAPIKeyHeaderSecurityDefinition method is called", func() {
			secDef := newAPIKeyHeaderSecurityDefinition(secDefName, apiKeyName)
			Convey("Then the sec def name should match", func() {
				So(secDef.Name, ShouldEqual, secDefName)
			})
			Convey("And the sec def type should be apiKey", func() {
				So(secDef.Type, ShouldEqual, "apiKey")
			})
			Convey("And the sec def apikey should contain the apikey name and the right IN inHeader value", func() {
				So(secDef.apiKey.Name, ShouldEqual, apiKeyName)
				So(secDef.apiKey.In, ShouldEqual, inHeader)
			})
		})
	})
}

func TestNewAPIKeyQuerySecurityDefinition(t *testing.T) {
	Convey("Given a secDefName and an apiKeyName  ", t, func() {
		secDefName := "api_query_auth"
		apiKeyName := "Authorization"
		Convey("When newAPIKeyQuerySecurityDefinition method is called", func() {
			secDef := newAPIKeyQuerySecurityDefinition(secDefName, apiKeyName)
			Convey("Then the sec def name should match", func() {
				So(secDef.Name, ShouldEqual, secDefName)
			})
			Convey("And the sec def type should be apiKey", func() {
				So(secDef.Type, ShouldEqual, "apiKey")
			})
			Convey("And the sec def apikey should contain the apikey name and the right IN inQuery value", func() {
				So(secDef.apiKey.Name, ShouldEqual, apiKeyName)
				So(secDef.apiKey.In, ShouldEqual, inQuery)
			})
		})
	})
}

func TestGetTerraformConfigurationName(t *testing.T) {
	Convey("Given a SpecSecurityDefinition with a compliant name ", t, func() {
		s := SpecSecurityDefinition{
			Name: "name",
		}
		Convey("When getTerraformConfigurationName method is called", func() {
			tfName := s.getTerraformConfigurationName()
			Convey("Then the result should match the original name", func() {
				So(tfName, ShouldEqual, s.Name)
			})
		})
	})
	Convey("Given a SpecSecurityDefinition with a Non compliant name ", t, func() {
		s := SpecSecurityDefinition{
			Name: "nonCompliantName",
		}
		Convey("When getTerraformConfigurationName method is called", func() {
			tfName := s.getTerraformConfigurationName()
			Convey("Then the result should bethe compliant name (snake case)", func() {
				So(tfName, ShouldEqual, "non_compliant_name")
			})
		})
	})
}

func TestFindSecurityDefinitionFor(t *testing.T) {
	Convey("Given a SpecSecurityDefinitions", t, func() {
		expectedSecDefName := "secDefName"
		s := SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition(expectedSecDefName, "Authorization"),
		}
		Convey("When findSecurityDefinitionFor method is called with an existing sec def name", func() {
			secDef := s.findSecurityDefinitionFor(expectedSecDefName)
			Convey("Then the secDef result should not be nil", func() {
				So(secDef, ShouldNotBeNil)
			})
			Convey("And the secDef should match the expected one", func() {
				So(secDef.Name, ShouldEqual, expectedSecDefName)
			})
		})
		Convey("When findSecurityDefinitionFor method is called with a NON existing sec def name", func() {
			secDef := s.findSecurityDefinitionFor("nonExistingSecDefName")
			Convey("Then the secDef result should be nil", func() {
				So(secDef, ShouldBeNil)
			})
		})
	})

}


