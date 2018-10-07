package openapi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTerraformCompliantPropertyName(t *testing.T) {
	Convey("Given a specSchemaDefinitionProperty that has a name and not preferred name and name is compliant already", t, func() {
		s := &specSchemaDefinitionProperty{
			Name: "compliant_prop_name",
			Type: typeString,
		}
		Convey("When getTerraformCompliantPropertyName method is called", func() {
			compliantName := s.getTerraformCompliantPropertyName()
			Convey("Then the resulted bool should be true", func() {
				So(compliantName, ShouldEqual, "compliant_prop_name")
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that has a name and not preferred name and name is NOT compliant", t, func() {
		s := &specSchemaDefinitionProperty{
			Name: "nonCompliantName",
			Type: typeString,
		}
		Convey("When getTerraformCompliantPropertyName method is called", func() {
			compliantName := s.getTerraformCompliantPropertyName()
			Convey("Then the resulted bool should be true", func() {
				So(compliantName, ShouldEqual, "non_compliant_name")
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that has a name AND a preferred name and name is compliant", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:          "compliant_prop_name",
			PreferredName: "preferred_compliant_name",
			Type:          typeString,
		}
		Convey("When getTerraformCompliantPropertyName method is called", func() {
			compliantName := s.getTerraformCompliantPropertyName()
			Convey("Then the resulted bool should be true", func() {
				So(compliantName, ShouldEqual, "preferred_compliant_name")
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that has a name AND a preferred name and name is NOT compliant", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:          "compliant_prop_name",
			PreferredName: "preferredNonCompliantName",
			Type:          typeString,
		}
		Convey("When getTerraformCompliantPropertyName method is called", func() {
			compliantName := s.getTerraformCompliantPropertyName()
			Convey("Then the resulted bool should be true", func() {
				So(compliantName, ShouldEqual, "preferred_non_compliant_name")
			})
		})
	})
}

func TestIsPropertyNamedID(t *testing.T) {
	Convey("Given a specSchemaDefinitionProperty that is PropertyNamedID", t, func() {
		s := &specSchemaDefinitionProperty{
			Name: idDefaultPropertyName,
			Type: typeString,
		}
		Convey("When isPropertyNamedID method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedID()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is PropertyNamedID with no compliant name", t, func() {
		s := &specSchemaDefinitionProperty{
			Name: "ID",
			Type: typeString,
		}
		Convey("When isPropertyNamedID method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedID()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is NOT PropertyNamedID", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "some_other_property_name",
			Type:     typeString,
			Required: false,
		}
		Convey("When isPropertyNamedID method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedID()
			Convey("Then the resulted bool should be false", func() {
				So(isPropertyNamedStatus, ShouldBeFalse)
			})
		})
	})
}

func TestIsPropertyNamedStatus(t *testing.T) {
	Convey("Given a specSchemaDefinitionProperty that is PropertyNamedStatus", t, func() {
		s := &specSchemaDefinitionProperty{
			Name: statusDefaultPropertyName,
			Type: typeString,
		}
		Convey("When isPropertyNamedStatus method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedStatus()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is PropertyNamedStatus with no compliant name", t, func() {
		s := &specSchemaDefinitionProperty{
			Name: "Status",
			Type: typeString,
		}
		Convey("When isPropertyNamedStatus method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedStatus()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is NOT PropertyNamedStatus", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "some_other_property_name",
			Type:     typeString,
			Required: false,
		}
		Convey("When isPropertyNamedStatus method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedStatus()
			Convey("Then the resulted bool should be false", func() {
				So(isPropertyNamedStatus, ShouldBeFalse)
			})
		})
	})
}

func TestIsObjectProperty(t *testing.T) {
	Convey("Given a specSchemaDefinitionProperty that is ObjectProperty", t, func() {
		s := &specSchemaDefinitionProperty{
			Name: "object_prop",
			Type: typeObject,
		}
		Convey("When isObjectProperty method is called", func() {
			isArrayProperty := s.isObjectProperty()
			Convey("Then the resulted bool should be true", func() {
				So(isArrayProperty, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is NOT ObjectProperty", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeString,
			Required: false,
		}
		Convey("When isObjectProperty method is called", func() {
			isArrayProperty := s.isObjectProperty()
			Convey("Then the resulted bool should be false", func() {
				So(isArrayProperty, ShouldBeFalse)
			})
		})
	})
}

func TestIsArrayProperty(t *testing.T) {
	Convey("Given a specSchemaDefinitionProperty that is ArrayProperty", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeList,
			Required: true,
		}
		Convey("When isArrayProperty method is called", func() {
			isArrayProperty := s.isArrayProperty()
			Convey("Then the resulted bool should be true", func() {
				So(isArrayProperty, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is NOT ArrayProperty", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeString,
			Required: false,
		}
		Convey("When isArrayProperty method is called", func() {
			isArrayProperty := s.isArrayProperty()
			Convey("Then the resulted bool should be false", func() {
				So(isArrayProperty, ShouldBeFalse)
			})
		})
	})
}

func TestIsRequired(t *testing.T) {
	Convey("Given a specSchemaDefinitionProperty that is Required", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeString,
			Required: true,
		}
		Convey("When isRequired method is called", func() {
			isRequired := s.isRequired()
			Convey("Then the resulted bool should be true", func() {
				So(isRequired, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is NOT Required", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeString,
			Required: false,
		}
		Convey("When isRequired method is called", func() {
			isRequired := s.isRequired()
			Convey("Then the resulted bool should be false", func() {
				So(isRequired, ShouldBeFalse)
			})
		})
	})
}

func TestIsReadOnly(t *testing.T) {
	Convey("Given a specSchemaDefinitionProperty that is readonly", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeString,
			ReadOnly: true,
		}
		Convey("When isReadOnly method is called", func() {
			isReadOnly := s.isReadOnly()
			Convey("Then the resulted bool should be true", func() {
				So(isReadOnly, ShouldBeTrue)
			})
		})
	})

	Convey("Given a specSchemaDefinitionProperty that is NOT readonly", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeString,
			ReadOnly: false,
		}
		Convey("When isReadOnly method is called", func() {
			isReadOnly := s.isReadOnly()
			Convey("Then the resulted bool should be false", func() {
				So(isReadOnly, ShouldBeFalse)
			})
		})
	})
}

func TestTerraformSchema(t *testing.T) {
	Convey("Given a swagger schema definition that has a property of type 'string' which is required", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     typeString,
			ReadOnly: false,
			Required: true,
		}
		Convey("When createResourceSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted tfPropSchema should be of type string too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeString)
				So(tfPropSchema.Required, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'integer'", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "int_prop",
			Type:     typeInt,
			ReadOnly: false,
			Required: true,
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be of type int too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeInt)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'number'", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "number_prop",
			Type:     typeFloat,
			ReadOnly: false,
			Required: true,
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be of type float too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeFloat)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'boolean'", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "boolean_prop",
			Type:     typeBool,
			ReadOnly: false,
			Required: true,
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be of type int too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeBool)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array'", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "array_prop",
			Type:     typeList,
			ReadOnly: false,
			Required: true,
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be of type array too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
			})
			Convey("And the array elements are of the default type string (only supported type for now)", func() {
				So(reflect.TypeOf(tfPropSchema.Elem).Elem(), ShouldEqual, reflect.TypeOf(schema.Schema{}))
				So(tfPropSchema.Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type object and a ref pointing to the schema", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "object_prop",
			Type:     typeObject,
			ReadOnly: false,
			Required: true,
			SpecSchemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					&specSchemaDefinitionProperty{
						Name: "message",
						Type: typeString,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the tf resource schema returned should not be nil", func() {
				So(tfPropSchema, ShouldNotBeNil)
			})
			Convey("And the tf resource schema returned should contained the sub property - 'message'", func() {
				So(tfPropSchema.Elem.(*schema.Resource).Schema, ShouldContainKey, "message")
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type object that has nested schema and property named id", t, func() {
		s := &specSchemaDefinitionProperty{
			Name:     "object_prop",
			Type:     typeObject,
			ReadOnly: false,
			Required: true,
			SpecSchemaDefinition: &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					&specSchemaDefinitionProperty{
						Name: "id",
						Type: typeString,
					},
					&specSchemaDefinitionProperty{
						Name: "message",
						Type: typeString,
					},
				},
			},
		}
		Convey("When createTerraformPropertyBasicSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the tf resource schema returned should not be nil", func() {
				So(tfPropSchema, ShouldNotBeNil)
			})
			Convey("And the tf resource schema returned should contain the sub property - 'message'", func() {
				So(tfPropSchema.Elem.(*schema.Resource).Schema, ShouldContainKey, "message")
			})
			Convey("And the tf resource schema returned should contain the sub property - 'id' and should not be ignored", func() {
				So(tfPropSchema.Elem.(*schema.Resource).Schema, ShouldContainKey, "id")
			})
		})
	})

	Convey("Given a string schemaDefinitionProperty that is required, not computed, forceNew, sensitive, not immutable and has a default value", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", true, false, true, true, false, false, false, "defaultValue")
		Convey("When terraformSchema is called with a schema definition property that is required, force new, sensitive and has a default value", func() {
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
	})

	Convey("Given a schemaDefinitionProperty that is computed", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, false, "")
		Convey("When createTerraformPropertySchema is called with a schema definition property that is readonly", func() {
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
	})

	Convey("Given a schemaDefinitionProperty that is computed and has a default value set", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, false, "defaultValue")
		Convey("When createTerraformPropertySchema is called with a schema definition property that validation fails due to read only field having a default value", func() {
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
	})

	Convey("Given a schemaDefinitionProperty that is forceNew and immutable ", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, false, true, false, true, false, false, "")
		Convey("When createTerraformPropertySchema is called with a schema definition property that validation fails due to immutable and forceNew set", func() {
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
	})

	Convey("Given a schemaDefinitionProperty that is computed and required", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", true, true, false, false, false, false, false, nil)
		Convey("When createTerraformPropertySchema is called with a schema definition property that validation fails due to required and computed set", func() {
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

func TestValidateFunc(t *testing.T) {

	Convey("Given a schemaDefinitionProperty that is computed and has a default value set", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, false, "defaultValue")
		Convey("When validateFunc is called with a schema definition property", func() {
			Convey("And the schema returned should be configured with a validate function", func() {
				So(s.validateFunc(), ShouldNotBeNil)
			})
			Convey("And the schema validate function should return an error due to read only field having a default value", func() {
				_, err := s.validateFunc()(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "'propertyName.' is configured as 'readOnly' and can not have a default expectedValue.")
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is forceNew and immutable ", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, false, true, false, true, false, false, "")
		Convey("When validateFunc is called with a schema definition property", func() {
			Convey("And the schema returned should be configured with a validate function", func() {
				So(s.validateFunc(), ShouldNotBeNil)
			})
			Convey("And the schema validate function should return an error due to immutable and forceNew set", func() {
				_, err := s.validateFunc()(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "property 'propertyName' is configured as immutable and can not be configured with forceNew too")
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is computed and required", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", true, true, false, false, false, false, false, nil)
		Convey("When validateFunc is called with a schema definition property", func() {
			Convey("And the schema returned should be configured with a validate function", func() {
				So(s.validateFunc(), ShouldNotBeNil)
			})
			Convey("And the schema validate function should return an error due to required and computed set", func() {
				_, err := s.validateFunc()(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "property 'propertyName' is configured as required and can not be configured as computed too")
			})
		})
	})
}
