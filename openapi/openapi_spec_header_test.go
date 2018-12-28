package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetHeaderTerraformConfigurationName(t *testing.T) {
	Convey("Given a SpecHeaderParam that has a compliant name and not terraform name", t, func() {
		specHeaderParam := SpecHeaderParam{
			Name:          "some_name",
			TerraformName: "",
		}
		Convey("When GetHeaderTerraformConfigurationName method is called", func() {
			terraformConfigurationName := specHeaderParam.GetHeaderTerraformConfigurationName()
			Convey("And the the terraformConfigurationName returned be the same", func() {
				So(terraformConfigurationName, ShouldEqual, "some_name")
			})
		})
	})

	Convey("Given a SpecHeaderParam that has a NON compliant name and not terraform name", t, func() {
		specHeaderParam := SpecHeaderParam{
			Name:          "someName",
			TerraformName: "",
		}
		Convey("When GetHeaderTerraformConfigurationName method is called", func() {
			terraformConfigurationName := specHeaderParam.GetHeaderTerraformConfigurationName()
			Convey("And the the terraformConfigurationName returned be the compliant name", func() {
				So(terraformConfigurationName, ShouldEqual, "some_name")
			})
		})
	})

	Convey("Given a SpecHeaderParam that has a name and a terraform name", t, func() {
		specHeaderParam := SpecHeaderParam{
			Name:          "someName",
			TerraformName: "terraform_name",
		}
		Convey("When GetHeaderTerraformConfigurationName method is called", func() {
			terraformConfigurationName := specHeaderParam.GetHeaderTerraformConfigurationName()
			Convey("And the the terraformConfigurationName returned be terraform preferred name", func() {
				So(terraformConfigurationName, ShouldEqual, "terraform_name")
			})
		})
	})

	Convey("Given a SpecHeaderParam that has a name and a terraform name which is not terraform compliant name", t, func() {
		specHeaderParam := SpecHeaderParam{
			Name:          "someName",
			TerraformName: "terraformName",
		}
		Convey("When GetHeaderTerraformConfigurationName method is called", func() {
			terraformConfigurationName := specHeaderParam.GetHeaderTerraformConfigurationName()
			Convey("And the the terraformConfigurationName returned be terraform compliant preferred name", func() {
				So(terraformConfigurationName, ShouldEqual, "terraform_name")
			})
		})
	})
}

func TestSpecHeaderExists(t *testing.T) {
	Convey("Given a list of SpecHeaderParameters", t, func() {
		alreadyExistingHeaderName := "alreadyExistingHeaderName"
		specHeaderParameters := SpecHeaderParameters{
			SpecHeaderParam{
				Name:          alreadyExistingHeaderName,
				TerraformName: "",
			},
		}
		Convey("When specHeaderExists method is called with an existing header param", func() {
			header := SpecHeaderParam{Name: alreadyExistingHeaderName}
			exists := specHeaderParameters.specHeaderExists(header)
			Convey("Then the specSecuritySchemes should be true", func() {
				So(exists, ShouldBeTrue)
			})
		})
	})

	Convey("Given a list of SpecHeaderParameters", t, func() {
		specHeaderParameters := SpecHeaderParameters{
			SpecHeaderParam{
				Name:          "alreadyExistingHeaderName",
				TerraformName: "",
			},
		}
		Convey("When securitySchemeExists method is called with a NON existing header param", func() {
			header := SpecHeaderParam{Name: "nonExistingHeaderName"}
			exists := specHeaderParameters.specHeaderExists(header)
			Convey("Then the specSecuritySchemes should be false", func() {
				So(exists, ShouldBeFalse)
			})
		})
	})
}
