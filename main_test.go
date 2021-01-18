package main

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestInitProvider(t *testing.T) {
	Convey("Given a valid binary name and the OTF_VAR pointing at a valid OpenAPI document", t, func() {
		binaryName := "terraform-provider-openapi"
		file, err := ioutil.TempFile("", "openapi.yaml")
		if err != nil {
			log.Fatal(err)
		}
		file.Write([]byte(`swagger: "2.0"
paths:
  /v1/cdns:`))
		os.Setenv("OTF_VAR_openapi_SWAGGER_URL", file.Name())
		Convey("When initProvider method is called", func() {
			providerName, err := initProvider(binaryName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then the provider returned should not be nil", func() {
				So(providerName, ShouldNotBeNil)
			})
		})
		os.Unsetenv("OTF_VAR_openapi_SWAGGER_URL")
	})
	Convey("Given an invalid binary name", t, func() {
		binaryName := "some-invalid-binary-name"
		Convey("When initProvider method is called", func() {
			providerName, err := initProvider(binaryName)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "error getting the provider's name from the binary 'some-invalid-binary-name': provider binary name (some-invalid-binary-name) does not match terraform naming convention 'terraform-provider-{name}', please rename the provider binary")
			})
			Convey("And the plugin returned should be nil", func() {
				So(providerName, ShouldBeNil)
			})
		})
	})
	Convey("Given an invalid OpenAPI document", t, func() {
		binaryName := "terraform-provider-openapi"
		file, err := ioutil.TempFile("", "invalid_openapi_document.yaml")
		if err != nil {
			log.Fatal(err)
		}
		file.Write([]byte(`some non valid open api document`))
		os.Setenv("OTF_VAR_openapi_SWAGGER_URL", file.Name())
		Convey("When initProvider method is called", func() {
			providerName, err := initProvider(binaryName)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldContainSubstring, "error initialising the terraform provider: plugin OpenAPI spec analyser error: failed to retrieve the OpenAPI document")
				So(err.Error(), ShouldContainSubstring, "error = analyzed: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `some no...` into map[interface {}]interface {}")
			})
			Convey("And the plugin returned should be nil", func() {
				So(providerName, ShouldBeNil)
			})
		})
	})
}

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
	Convey("Given a valid pluginName with correct version using multiple digits", t, func() {
		binaryName := "terraform-provider-goa_v10.34.1"
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

	Convey("Given a valid binary path containing multiple matches for terraform-provider-", t, func() {
		binaryName := "/Users/user/dev/src/github.com/terraform-provider-openapi/examples/swaggercodegen/.terraform/providers/terraform.example.com/examplecorp/swaggercodegen/1.0.0/darwin_amd64/terraform-provider-swaggercodegen"
		Convey("When getProviderName method is called", func() {
			providerName, err := getProviderName(binaryName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(providerName, ShouldEqual, "swaggercodegen")
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
