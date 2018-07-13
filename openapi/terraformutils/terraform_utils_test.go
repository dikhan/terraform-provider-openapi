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
