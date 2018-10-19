package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetProviderName(t *testing.T) {
	Convey("Given a valid pluginName with no version", t, func() {
		binaryName := "terraform-provider-goa"
		Convey("When getProviderName method is called", func() {
			providerName, err := getProviderName(binaryName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(providerName, ShouldEqual, "goa")
			})
		})
	})
	Convey("Given a valid pluginName with correct version", t, func() {
		binaryName := "terraform-provider-goa_v1.0.0"
		Convey("When getProviderName method is called", func() {
			providerName, err := getProviderName(binaryName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(providerName, ShouldEqual, "goa")
			})
		})
	})
	Convey("Given a valid pluginName with numbers in the name", t, func() {
		binaryName := "terraform-provider-goa1234"
		Convey("When getProviderName method is called", func() {
			providerName, err := getProviderName(binaryName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(providerName, ShouldEqual, "goa1234")
			})
		})
	})
	Convey("Given a valid pluginName with incorrect version format", t, func() {
		binaryName := "terraform-provider-goa_v1"
		Convey("When getProviderName method is called", func() {
			_, err := getProviderName(binaryName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(err.Error(), ShouldEqual, "provider binary name (terraform-provider-goa_v1) does not match terraform naming convention 'terraform-provider-{name}', please rename the provider binary")
			})
		})
	})
	Convey("Given a NON valid pluginName", t, func() {
		binaryName := "terraform-providerWrongName-goa"
		Convey("When getProviderName method is called", func() {
			_, err := getProviderName(binaryName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(err.Error(), ShouldEqual, "provider binary name (terraform-providerWrongName-goa) does not match terraform naming convention 'terraform-provider-{name}', please rename the provider binary")
			})
		})
	})

}
