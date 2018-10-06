package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateTerraformPropertySchema(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		//r := resourceFactory{}
		Convey("When createTerraformPropertySchema is called with a schema definition property that is required, force new, sensitive and has a default value", func() {
			s := newStringSchemaDefinitionProperty("propertyName", "", true, false, true, true, false, false, false, "defaultValue")
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema returned should be configured as required", func() {
				So(terraformPropertySchema.Required, ShouldBeTrue)
			})
			Convey("And the schema returned should be configured as NOT computed", func() {
				So(terraformPropertySchema.Computed, ShouldBeFalse)
			})
			Convey("And the schema returned should be configured as force new", func() {
				So(terraformPropertySchema.ForceNew, ShouldBeTrue)
			})
			Convey("And the schema returned should be configured as sensitive", func() {
				So(terraformPropertySchema.Sensitive, ShouldBeTrue)
			})
			Convey("And the schema returned should be configured with default value", func() {
				So(terraformPropertySchema.Default, ShouldEqual, s.Default)
			})
			Convey("And the schema returned should be configured with a validate function", func() {
				So(terraformPropertySchema.ValidateFunc, ShouldNotBeNil)
			})
		})

		Convey("When createTerraformPropertySchema is called with a schema definition property that is readonly", func() {
			s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, false, "")
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema returned should be configured as computed", func() {
				So(terraformPropertySchema.Computed, ShouldBeTrue)
			})
			Convey("And the schema returned should be configured with a validate function", func() {
				So(terraformPropertySchema.ValidateFunc, ShouldNotBeNil)
			})
		})

		Convey("When createTerraformPropertySchema is called with a schema definition property that validation fails due to read only field having a default value", func() {
			s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, false, "defaultValue")
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema returned should be configured as computed", func() {
				So(terraformPropertySchema.Computed, ShouldBeTrue)
			})
			Convey("And the schema returned should be configured with a validate function", func() {
				So(terraformPropertySchema.ValidateFunc, ShouldNotBeNil)
			})
			Convey("And the schema validate function should return an error ", func() {
				_, err := terraformPropertySchema.ValidateFunc(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "'propertyName.' is configured as 'readOnly' and can not have a default expectedValue.")
			})
		})

		Convey("When createTerraformPropertySchema is called with a schema definition property that validation fails due to immutable and forceNew set", func() {
			s := newStringSchemaDefinitionProperty("propertyName", "", false, false, true, false, true, false, false, "")
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema returned should be configured with a validate function", func() {
				So(terraformPropertySchema.ValidateFunc, ShouldNotBeNil)
			})
			Convey("And the schema validate function should return an error ", func() {
				_, err := terraformPropertySchema.ValidateFunc(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "property 'propertyName' is configured as immutable and can not be configured with forceNew too")
			})
		})

		Convey("When createTerraformPropertySchema is called with a schema definition property that validation fails due to required and computed set", func() {
			s := newStringSchemaDefinitionProperty("propertyName", "", true, true, false, false, false, false, false, nil)
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema returned should be configured as required", func() {
				So(terraformPropertySchema.Required, ShouldBeTrue)
			})
			Convey("And the schema returned should be configured as computed", func() {
				So(terraformPropertySchema.Computed, ShouldBeTrue)
			})
			Convey("And the schema returned should be configured with a validate function", func() {
				So(terraformPropertySchema.ValidateFunc, ShouldNotBeNil)
			})
			Convey("And the schema validate function should return an error ", func() {
				_, err := terraformPropertySchema.ValidateFunc(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "property 'propertyName' is configured as required and can not be configured as computed too")
			})
		})
	})
}
