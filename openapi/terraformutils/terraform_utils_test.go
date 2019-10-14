package terraformutils

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTerraformUtilsGetTerraformPluginsVendorDir(t *testing.T) {
	Convey("Given an TerraformUtils set up with a homeDir and darwin platform", t, func() {
		t := TerraformUtils{
			HomeDir:  "/Users/username",
			Platform: "darwin",
		}
		Convey("When GetTerraformPluginsVendorDir is called", func() {
			vendorDir, err := t.GetTerraformPluginsVendorDir()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And vendor dir should be the default one as no os was specified`", func() {
				So(vendorDir, ShouldEqual, "/Users/username/.terraform.d/plugins")
			})
		})
	})
	Convey("Given an TerraformUtils set up with a homeDir and linux platform", t, func() {
		t := TerraformUtils{
			HomeDir:  "/Users/username",
			Platform: "linux",
		}
		Convey("When GetTerraformPluginsVendorDir is called", func() {
			vendorDir, err := t.GetTerraformPluginsVendorDir()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And vendor dir should be the default one as no os was specified`", func() {
				So(vendorDir, ShouldEqual, "/Users/username/.terraform.d/plugins")
			})
		})
	})
	Convey("Given an TerraformUtils set up with a homeDir and windows platform", t, func() {
		homeDir := "C:\\Users\\username\\"
		t := TerraformUtils{
			Platform: "windows",
			HomeDir:  homeDir,
		}
		Convey("When GetTerraformPluginsVendorDir is called", func() {
			vendorDir, err := t.GetTerraformPluginsVendorDir()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And vendor dir should be the windows specific path`", func() {
				So(vendorDir, ShouldEqual, fmt.Sprintf("%s\\AppData\\terraform.d\\plugins", homeDir))
			})
		})
	})
	Convey("Given an TerraformUtils missing the homeDir configuration", t, func() {
		t := TerraformUtils{
			HomeDir:  "",
			Platform: "darwin",
		}
		Convey("When GetTerraformPluginsVendorDir is called", func() {
			_, err := t.GetTerraformPluginsVendorDir()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And error message should match the expected one`", func() {
				So(err.Error(), ShouldEqual, "mandatory HomeDir value missing")
			})
		})
	})
	Convey("Given an TerraformUtils missing the platform configuration", t, func() {
		t := TerraformUtils{
			HomeDir:  "/Users/username",
			Platform: "",
		}
		Convey("When GetTerraformPluginsVendorDir is called", func() {
			_, err := t.GetTerraformPluginsVendorDir()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And error message should match the expected one`", func() {
				So(err.Error(), ShouldEqual, "mandatory platform information is missing")
			})
		})
	})
}

func TestConvertToTerraformCompliantFieldName(t *testing.T) {
	testCases := []struct {
		name                 string
		inputPropertyName    string
		expectedPropertyName string
		expectedError        string
	}{
		{name: "property name that is terraform name compliant", inputPropertyName: "some_prop_name_that_is_terraform_field_name_compliant", expectedPropertyName: "some_prop_name_that_is_terraform_field_name_compliant"},
		{name: "property name that is NOT terraform name compliant", inputPropertyName: "thisIsACamelCaseNameWhichIsNotTerraformNameCompliant", expectedPropertyName: "this_is_a_camel_case_name_which_is_not_terraform_name_compliant"},
		{name: "property name that is terraform name compliant but with numbers with no _ between number and next word", inputPropertyName: "cdns_v1id", expectedPropertyName: "cdns_v1_id"},
		{name: "property name that is terraform name compliant with one number in the middle", inputPropertyName: "cdns_v1_id", expectedPropertyName: "cdns_v1_id"},
		{name: "property name that is terraform name compliant with multiple numbers", inputPropertyName: "cdns_v1_firewall_v2_id", expectedPropertyName: "cdns_v1_firewall_v2_id"},
		{name: "property name that is terraform name compliant with one number at the end", inputPropertyName: "cdns_v1", expectedPropertyName: "cdns_v1"},
		{name: "property name that ends with double underscore ( __ )", inputPropertyName: "cdns__", expectedPropertyName: "cdns__"},
		{name: "property name that has v_1 on purpose", inputPropertyName: "cdns_v_1", expectedPropertyName: "cdns_v_1"},
		{name: "property name that ends with __1", inputPropertyName: "cdns__1", expectedPropertyName: "cdns__1"},
		{name: "property name that ends with v1_", inputPropertyName: "cdns_v1_", expectedPropertyName: "cdns_v1_"},
		{name: "property name that ends with _1", inputPropertyName: "cdns_1", expectedPropertyName: "cdns_1"},
		{name: "property name with underscore at the end", inputPropertyName: "cdns_", expectedPropertyName: "cdns_"},
		{name: "property name with leading and trailing underscores", inputPropertyName: "_cdns_", expectedPropertyName: "_cdns_"},
		{name: "property name 1", inputPropertyName: "1", expectedPropertyName: "1"},
		{name: "property name with a number and an underscore at the end", inputPropertyName: "cdns_1_", expectedPropertyName: "cdns_1_"},
	}

	for _, tc := range testCases {
		Convey("Given a "+tc.name, t, func() {
			Convey("When ConvertToTerraformCompliantName method is called", func() {
				fieldName := ConvertToTerraformCompliantName(tc.inputPropertyName)
				Convey("The string returned should be the expected one", func() {
					So(fieldName, ShouldEqual, tc.expectedPropertyName)
				})
			})
		})
	}
}

func TestCreateSchema(t *testing.T) {
	Convey("Given an environment variable, schemaType of type string, required property and an empty default value", t, func() {
		propertyName := "propertyName"
		envVariableValue := "someValue"
		defaultValue := ""
		os.Setenv(strings.ToUpper(propertyName), envVariableValue)
		schemaType := schema.TypeString
		required := true
		Convey("When createSchema method is called", func() {
			schema := createSchema(propertyName, schemaType, required, defaultValue)
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
	Convey("Given a schemaType of type bool, an optional property and an empty default value", t, func() {
		schemaType := schema.TypeBool
		required := false
		defaultValue := ""
		Convey("When createSchema method is called", func() {
			schema := createSchema("propertyName", schemaType, required, defaultValue)
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

	Convey("Given a schemaType of type string, required property and a NON empty default value", t, func() {
		propertyName := "propertyName"
		defaultValue := "defaultValue"
		schemaType := schema.TypeString
		required := true
		Convey("When createSchema method is called", func() {
			schema := createSchema(propertyName, schemaType, required, defaultValue)
			Convey("Then the schema returned should be of type string", func() {
				So(schema.Type, ShouldEqual, schemaType)
			})
			Convey("And the schema returned should be required", func() {
				So(schema.Required, ShouldEqual, required)
			})
			Convey("And the schema default function should return the value set for te environment variable", func() {
				value, err := schema.DefaultFunc()
				So(err, ShouldBeNil)
				So(value, ShouldEqual, defaultValue)
			})
		})
		os.Unsetenv(strings.ToUpper(propertyName))
	})
}

func TestCreateStringSchema(t *testing.T) {
	Convey("Given a required property of type string", t, func() {
		required := true
		Convey("When CreateStringSchemaProperty method is called", func() {
			s := CreateStringSchemaProperty("propertyName", required, "")
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
