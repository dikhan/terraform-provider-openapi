package terraformutils

import (
	"github.com/hashicorp/terraform/helper/schema"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"strings"
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

func TestCreateSchema(t *testing.T) {
	Convey("Given an environment variable, schemaType of type string and a required property", t, func() {
		propertyName := "propertyName"
		envVariableValue := "someValue"
		os.Setenv(strings.ToUpper(propertyName), envVariableValue)
		schemaType := schema.TypeString
		required := true
		Convey("When createSchema method is called", func() {
			schema := createSchema(propertyName, schemaType, required)
			Convey("Then the schema returned should be of type string", func() {
				So(schema.Type, ShouldEqual, schemaType)
			})
			Convey("And the schema returned should be required", func() {
				So(schema.Required, ShouldEqual, required)
			})
			Convey("And the schema default function should return the value set for te environment variable", func() {
				value, err := schema.DefaultFunc()
				So(err, ShouldBeNil)
				So(value, ShouldEqual, envVariableValue)
			})
		})
		os.Unsetenv(strings.ToUpper(propertyName))
	})
	Convey("Given a schemaType of type bool and an optional property", t, func() {
		schemaType := schema.TypeBool
		required := false
		Convey("When createSchema method is called", func() {
			schema := createSchema("propertyName", schemaType, required)
			Convey("Then the schema returned should be of type bool", func() {
				So(schema.Type, ShouldEqual, schemaType)
			})
			Convey("And the schema returned should be optional", func() {
				So(schema.Optional, ShouldEqual, !required)
			})
			Convey("And the schema default function should return nil as there's no env variable matching the property name", func() {
				value, err := schema.DefaultFunc()
				So(err, ShouldBeNil)
				So(value, ShouldBeNil)
			})
		})
	})
}

func TestCreateStringSchema(t *testing.T) {
	Convey("Given a required property of type string", t, func() {
		required := true
		Convey("When CreateStringSchema method is called", func() {
			s := CreateStringSchema("propertyName", required)
			Convey("Then the schema returned should be of type string", func() {
				So(s.Type, ShouldEqual, schema.TypeString)
			})
			Convey("And the schema returned should be required", func() {
				So(s.Required, ShouldEqual, required)
			})
			Convey("And the schema default function should return nil as there's no env variable matching the property name", func() {
				value, err := s.DefaultFunc()
				So(err, ShouldBeNil)
				So(value, ShouldBeNil)
			})
		})
	})
}

func TestEnvDefaultFunc(t *testing.T) {
	Convey("Given a property name that has an environment variable set up and nil default value", t, func() {
		propertyName := "propertyName"
		envVariableValue := "someValue"
		os.Setenv(strings.ToUpper(propertyName), envVariableValue)
		Convey("When envDefaultFunc method is called", func() {
			defaultFunc := envDefaultFunc(propertyName, nil)
			Convey("And the returned defaultFunc is invoked the value returned should be the value of the environment variable", func() {
				value, err := defaultFunc()
				So(err, ShouldBeNil)
				So(value, ShouldEqual, envVariableValue)
			})
		})
		os.Unsetenv(strings.ToUpper(propertyName))
	})
	Convey("Given a property name that DOES NOT have an environment variable set up and a default value is configured", t, func() {
		defaultValue := "someDefaultValue"
		Convey("When envDefaultFunc method is called", func() {
			defaultFunc := envDefaultFunc("propertyName", defaultValue)
			Convey("And the returned defaultFunc is invoked the value returned should be the defaultValue", func() {
				value, err := defaultFunc()
				So(err, ShouldBeNil)
				So(value, ShouldEqual, defaultValue)
			})
		})
	})
}
