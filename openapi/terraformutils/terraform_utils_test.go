package terraformutils

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTerraformUtilsGetTerraformPluginsVendorDir(t *testing.T) {
	Convey("Given a TerraformUtils init with linux runtime", t, func() {
		t := TerraformUtils{Runtime: "linux"}
		Convey("When GetTerraformPluginsVendorDir method is called", func() {
			vendorDir, err := t.GetTerraformPluginsVendorDir()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the vendorDir should contain", func() {
				So(vendorDir, ShouldContainSubstring, "/.terraform.d/plugins")
			})
		})
	})

	Convey("Given a TerraformUtils init with windows runtime", t, func() {
		t := TerraformUtils{Runtime: "windows"}
		Convey("When GetTerraformPluginsVendorDir method is called", func() {
			vendorDir, err := t.GetTerraformPluginsVendorDir()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the vendorDir should contain", func() {
				So(vendorDir, ShouldContainSubstring, "/terraform.d/plugins")
			})
		})
	})
}

func TestConvertToTerraformCompliantFieldName(t *testing.T) {
	Convey("Given a name that is terraform name compliant", t, func() {
		propertyName := "some_prop_name_that_is_terraform_field_name_compliant"
		Convey("When ConvertToTerraformCompliantName method is called", func() {
			fieldName := ConvertToTerraformCompliantName(propertyName)
			Convey("And string return is terraform field name compliant, ", func() {
				So(fieldName, ShouldEqual, propertyName)
			})
		})
	})

	Convey("Given a name that is NOT terraform name compliant", t, func() {
		propertyName := "thisIsACamelCaseNameWhichIsNotTerraformNameCompliant"
		Convey("When ConvertToTerraformCompliantName method is called", func() {
			fieldName := ConvertToTerraformCompliantName(propertyName)
			Convey("And string return is terraform field name compliant, ", func() {
				So(fieldName, ShouldEqual, "this_is_a_camel_case_name_which_is_not_terraform_name_compliant")
			})
		})
	})
}
