package openapi

import (
	"github.com/hashicorp/go-cty/cty"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTerraformCompliantPropertyName(t *testing.T) {
	Convey("Given a SpecSchemaDefinitionProperty that has a name and not preferred name and name is compliant already", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "compliant_prop_name",
			Type: TypeString,
		}
		Convey("When GetTerraformCompliantPropertyName method is called", func() {
			compliantName := s.GetTerraformCompliantPropertyName()
			Convey("Then the resulting name should be terraform compliant", func() {
				So(compliantName, ShouldEqual, "compliant_prop_name")
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that has a name and not preferred name and name is NOT compliant", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "nonCompliantName",
			Type: TypeString,
		}
		Convey("When GetTerraformCompliantPropertyName method is called", func() {
			compliantName := s.GetTerraformCompliantPropertyName()
			Convey("Then the resulting name should be terraform compliant", func() {
				So(compliantName, ShouldEqual, "non_compliant_name")
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that has a name AND a preferred name and name is compliant", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:          "compliant_prop_name",
			PreferredName: "preferred_compliant_name",
			Type:          TypeString,
		}
		Convey("When GetTerraformCompliantPropertyName method is called", func() {
			compliantName := s.GetTerraformCompliantPropertyName()
			Convey("Then the resulting name should be the preferred name", func() {
				So(compliantName, ShouldEqual, "preferred_compliant_name")
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that has a name AND a preferred name and preferred name is NOT compliant", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:          "compliant_prop_name",
			PreferredName: "preferredNonCompliantName",
			Type:          TypeString,
		}
		Convey("When GetTerraformCompliantPropertyName method is called", func() {
			compliantName := s.GetTerraformCompliantPropertyName()
			Convey("Then the resulting name should be preferred name verbatim", func() {
				// If preferred name is set, whether the value is compliant or not that will be the value configured for
				// the terraform schema property. If the name is not terraform name compliant, Terraform will complain about
				// it at runtime
				So(compliantName, ShouldEqual, "preferredNonCompliantName")
			})
		})
	})
}

func TestIsPrimitiveProperty(t *testing.T) {
	testCases := []struct {
		name                 string
		specSchemaDefinition SpecSchemaDefinitionProperty
		expectedResult       bool
	}{
		{
			name: "pecSchemaDefinitionProperty that is a primitive string",
			specSchemaDefinition: SpecSchemaDefinitionProperty{
				Name: "primitive_property",
				Type: TypeString,
			},
			expectedResult: true,
		},
		{
			name: "pecSchemaDefinitionProperty that is a primitive int",
			specSchemaDefinition: SpecSchemaDefinitionProperty{
				Name: "primitive_property",
				Type: TypeInt,
			},
			expectedResult: true,
		},
		{
			name: "pecSchemaDefinitionProperty that is a primitive float",
			specSchemaDefinition: SpecSchemaDefinitionProperty{
				Name: "primitive_property",
				Type: TypeFloat,
			},
			expectedResult: true,
		},
		{
			name: "pecSchemaDefinitionProperty that is a primitive bool",
			specSchemaDefinition: SpecSchemaDefinitionProperty{
				Name: "primitive_property",
				Type: TypeBool,
			},
			expectedResult: true,
		},
		{
			name: "pecSchemaDefinitionProperty that is not a primitive",
			specSchemaDefinition: SpecSchemaDefinitionProperty{
				Name: "primitive_property",
				Type: TypeObject,
			},
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		isPrimitiveProperty := tc.specSchemaDefinition.isPrimitiveProperty()
		assert.Equal(t, tc.expectedResult, isPrimitiveProperty, tc.name)
	}
}

func TestIsPropertyNamedID(t *testing.T) {
	Convey("Given a SpecSchemaDefinitionProperty that is PropertyNamedID", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: idDefaultPropertyName,
			Type: TypeString,
		}
		Convey("When isPropertyNamedID method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedID()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that is PropertyNamedID with no compliant name", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "ID",
			Type: TypeString,
		}
		Convey("When isPropertyNamedID method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedID()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that is NOT PropertyNamedID", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "some_other_property_name",
			Type:     TypeString,
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
	Convey("Given a SpecSchemaDefinitionProperty that is PropertyNamedStatus", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: statusDefaultPropertyName,
			Type: TypeString,
		}
		Convey("When isPropertyNamedStatus method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedStatus()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that is PropertyNamedStatus with no compliant name", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "Status",
			Type: TypeString,
		}
		Convey("When isPropertyNamedStatus method is called", func() {
			isPropertyNamedStatus := s.isPropertyNamedStatus()
			Convey("Then the resulted bool should be true", func() {
				So(isPropertyNamedStatus, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that is NOT PropertyNamedStatus", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "some_other_property_name",
			Type:     TypeString,
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
	Convey("Given a SpecSchemaDefinitionProperty that is ObjectProperty", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "object_prop",
			Type: TypeObject,
		}
		Convey("When isObjectProperty method is called", func() {
			isArrayProperty := s.isObjectProperty()
			Convey("Then the resulted bool should be true", func() {
				So(isArrayProperty, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that is NOT ObjectProperty", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
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
	Convey("Given a SpecSchemaDefinitionProperty that is ArrayProperty", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeList,
			Required: true,
		}
		Convey("When isArrayTypeProperty method is called", func() {
			isArrayProperty := s.isArrayProperty()
			Convey("Then the resulted bool should be true", func() {
				So(isArrayProperty, ShouldBeTrue)
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that is NOT ArrayProperty", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: false,
		}
		Convey("When isArrayTypeProperty method is called", func() {
			isArrayProperty := s.isArrayProperty()
			Convey("Then the resulted bool should be false", func() {
				So(isArrayProperty, ShouldBeFalse)
			})
		})
	})
}

func TestIsReadOnly(t *testing.T) {
	Convey("Given a SpecSchemaDefinitionProperty that is readOnly", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			ReadOnly: true,
		}
		Convey("When isReadOnly method is called", func() {
			isOptional := s.isReadOnly()
			Convey("Then the resulted bool should be true", func() {
				So(isOptional, ShouldBeTrue)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is NOT readOnly", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			ReadOnly: false,
		}
		Convey("When isReadOnly method is called", func() {
			isOptional := s.isReadOnly()
			Convey("Then the resulted bool should be false", func() {
				So(isOptional, ShouldBeFalse)
			})
		})
	})
}

func TestIsOptional(t *testing.T) {
	Convey("Given a SpecSchemaDefinitionProperty that is NOT Required", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: false,
		}
		Convey("When isOptional method is called", func() {
			isOptional := s.isOptional()
			Convey("Then the resulted bool should be true", func() {
				So(isOptional, ShouldBeTrue)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is NOT Required", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: true,
		}
		Convey("When isOptional method is called", func() {
			isOptional := s.isOptional()
			Convey("Then the resulted bool should be false", func() {
				So(isOptional, ShouldBeFalse)
			})
		})
	})
}

func TestIsRequired(t *testing.T) {
	Convey("Given a SpecSchemaDefinitionProperty that is Required", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: true,
		}
		Convey("When IsRequired method is called", func() {
			isRequired := s.IsRequired()
			Convey("Then the resulted bool should be true", func() {
				So(isRequired, ShouldBeTrue)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is NOT Required", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: false,
		}
		Convey("When IsRequired method is called", func() {
			isRequired := s.IsRequired()
			Convey("Then the resulted bool should be false", func() {
				So(isRequired, ShouldBeFalse)
			})
		})
	})
}

func TestShouldIgnoreArrayItemsOrder(t *testing.T) {
	Convey("Given a SpecSchemaDefinitionProperty that is a TypeList and where the 'x-terraform-ignore-order' ext is set to true", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:             "array_prop",
			Type:             TypeList,
			IgnoreItemsOrder: true,
		}
		Convey("When shouldIgnoreArrayItemsOrder method is called", func() {
			isRequired := s.shouldIgnoreArrayItemsOrder()
			Convey("Then the resulted bool should be true", func() {
				So(isRequired, ShouldBeTrue)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is a TypeList and where the 'x-terraform-ignore-order' ext is set to false", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:             "array_prop",
			Type:             TypeList,
			IgnoreItemsOrder: false,
		}
		Convey("When shouldIgnoreArrayItemsOrder method is called", func() {
			isRequired := s.shouldIgnoreArrayItemsOrder()
			Convey("Then the resulted bool should be false", func() {
				So(isRequired, ShouldBeFalse)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is a TypeList and where the 'x-terraform-ignore-order' ext is NOT set", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "array_prop",
			Type: TypeList,
		}
		Convey("When shouldIgnoreArrayItemsOrder method is called", func() {
			isRequired := s.shouldIgnoreArrayItemsOrder()
			Convey("Then the resulted bool should be false", func() {
				So(isRequired, ShouldBeFalse)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is NOT a TypeList where the 'x-terraform-ignore-order' ext is set", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:             "string_prop",
			Type:             TypeString,
			IgnoreItemsOrder: true,
		}
		Convey("When shouldIgnoreArrayItemsOrder method is called", func() {
			isRequired := s.shouldIgnoreArrayItemsOrder()
			Convey("Then the resulted bool should be false", func() {
				So(isRequired, ShouldBeFalse)
			})
		})
	})
}

func TestSchemaDefinitionPropertyIsComputed(t *testing.T) {
	Convey("Given a SpecSchemaDefinitionProperty that is optional and readonly", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: false,
			ReadOnly: true,
		}
		Convey("When isComputed method is called", func() {
			isReadOnly := s.isComputed()
			Convey("Then the resulted bool should be true", func() {
				So(isReadOnly, ShouldBeTrue)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is optional, NOT readonly BUT is optional-computed", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: false,
			ReadOnly: false,
			Computed: true,
			Default:  nil,
		}
		Convey("When isComputed method is called", func() {
			isReadOnly := s.isComputed()
			Convey("Then the resulted bool should be true", func() {
				So(isReadOnly, ShouldBeTrue)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that NOT optional", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			Required: true,
		}
		Convey("When isComputed method is called", func() {
			isReadOnly := s.isComputed()
			Convey("Then the resulted bool should be false", func() {
				So(isReadOnly, ShouldBeFalse)
			})
		})
	})
	Convey("Given a SpecSchemaDefinitionProperty that is NOT readonly", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			ReadOnly: false,
		}
		Convey("When isComputed method is called", func() {
			isReadOnly := s.isComputed()
			Convey("Then the resulted bool should be false", func() {
				So(isReadOnly, ShouldBeFalse)
			})
		})
	})
}

func TestSchemaDefinitionPropertyIsOptionalComputed(t *testing.T) {
	Convey("Given a property that is optional, not readOnly, is computed and does not have a default value (optional-computed of property where value is not known at plan time)", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type:     TypeString,
			Required: false,
			ReadOnly: false,
			Computed: true,
			Default:  nil,
		}
		Convey("When IsOptionalComputed method is called", func() {
			isOptionalComputed := s.IsOptionalComputed()
			Convey("Then value returned should be true", func() {
				So(isOptionalComputed, ShouldBeTrue)
			})
		})
	})
	Convey("Given a property that is not optional", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type:     TypeString,
			Required: true,
		}
		Convey("When IsOptionalComputed method is called", func() {
			isOptionalComputed := s.IsOptionalComputed()
			Convey("Then value returned should be false", func() {
				So(isOptionalComputed, ShouldBeFalse)
			})
		})
	})
	Convey("Given a property that is optional but readOnly", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type:     TypeString,
			Required: false,
			ReadOnly: true,
		}
		Convey("When IsOptionalComputed method is called", func() {
			isOptionalComputed := s.IsOptionalComputed()
			Convey("Then value returned should be false", func() {
				So(isOptionalComputed, ShouldBeFalse)
			})
		})
	})
	Convey("Given a property that is optional, not readOnly and it's not computed (purely optional use case)", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type:     TypeString,
			Required: false,
			ReadOnly: false,
			Computed: false,
			Default:  nil,
		}
		Convey("When IsOptionalComputed method is called", func() {
			isOptionalComputed := s.IsOptionalComputed()
			Convey("Then value returned should be false", func() {
				So(isOptionalComputed, ShouldBeFalse)
			})
		})
	})
	Convey("Given a property that is optional, not readOnly, computed but has a default value (optional-computed use case, but as far as terraform is concerned the default will be set om the terraform schema, making it available at plan time - this is by design in terraform)", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type:     TypeString,
			Required: false,
			ReadOnly: false,
			Computed: true,
			Default:  "something",
		}
		Convey("When IsOptionalComputed method is called", func() {
			isOptionalComputed := s.IsOptionalComputed()
			Convey("Then value returned should be false", func() {
				So(isOptionalComputed, ShouldBeFalse)
			})
		})
	})
}

func TestTerraformType(t *testing.T) {
	Convey("Given a swagger schema definition that has a property of type string", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type: TypeString,
		}
		Convey("When terraformType method is called", func() {
			valueType, err := s.terraformType()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(valueType, ShouldEqual, schema.TypeString)
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type int", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type: TypeInt,
		}
		Convey("When terraformType method is called", func() {
			valueType, err := s.terraformType()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(valueType, ShouldEqual, schema.TypeInt)
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type float", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type: TypeFloat,
		}
		Convey("When terraformType method is called", func() {
			valueType, err := s.terraformType()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(valueType, ShouldEqual, schema.TypeFloat)
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type bool", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type: TypeBool,
		}
		Convey("When terraformType method is called", func() {
			valueType, err := s.terraformType()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(valueType, ShouldEqual, schema.TypeBool)
			})
		})
	})
	Convey("Given a swagger schema definition that has an unsupported property type", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type: "unsupported",
		}
		Convey("When terraformType method is called", func() {
			valueType, err := s.terraformType()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				So(valueType, ShouldEqual, schema.TypeInvalid)
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type object", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type: TypeObject,
		}
		Convey("When terraformType method is called", func() {
			valueType, err := s.terraformType()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// Refer to: https://github.com/hashicorp/terraform-plugin-sdk/issues/616
				So(valueType, ShouldEqual, schema.TypeList)
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type list", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Type: TypeList,
		}
		Convey("When terraformType method is called", func() {
			valueType, err := s.terraformType()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(valueType, ShouldEqual, schema.TypeList)
			})
		})
	})
}

func TestIsTerraformListOfSimpleValues(t *testing.T) {
	Convey("Given a swagger schema definition that has a property of type 'list' with elements of type string", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "list_prop",
			Type:           TypeList,
			ArrayItemsType: TypeString,
		}
		Convey("When isTerraformListOfSimpleValues method is called", func() {
			isTerraformListOfSimpleValues, listSchema := s.isTerraformListOfSimpleValues()
			Convey("Then the result returned should be the expected one", func() {
				So(isTerraformListOfSimpleValues, ShouldBeTrue)
				So(reflect.TypeOf(*listSchema), ShouldEqual, reflect.TypeOf(schema.Schema{}))
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type 'list' with elements of type int", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "list_prop",
			Type:           TypeList,
			ArrayItemsType: TypeInt,
		}
		Convey("When isTerraformListOfSimpleValues method is called", func() {
			isTerraformListOfSimpleValues, listSchema := s.isTerraformListOfSimpleValues()
			Convey("Then the result returned should be the expected one", func() {
				So(isTerraformListOfSimpleValues, ShouldBeTrue)
				So(reflect.TypeOf(*listSchema), ShouldEqual, reflect.TypeOf(schema.Schema{}))
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type 'list' with elements of type float", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "list_prop",
			Type:           TypeList,
			ArrayItemsType: TypeFloat,
		}
		Convey("When isTerraformListOfSimpleValues method is called", func() {
			isTerraformListOfSimpleValues, listSchema := s.isTerraformListOfSimpleValues()
			Convey("Then the result returned should be the expected one", func() {
				So(isTerraformListOfSimpleValues, ShouldBeTrue)
				So(reflect.TypeOf(*listSchema), ShouldEqual, reflect.TypeOf(schema.Schema{}))
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type 'list' with elements of type bool", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "list_prop",
			Type:           TypeList,
			ArrayItemsType: TypeBool,
		}
		Convey("When isTerraformListOfSimpleValues method is called", func() {
			isTerraformListOfSimpleValues, listSchema := s.isTerraformListOfSimpleValues()
			Convey("Then the result returned should be the expected one", func() {
				So(isTerraformListOfSimpleValues, ShouldBeTrue)
				So(reflect.TypeOf(*listSchema), ShouldEqual, reflect.TypeOf(schema.Schema{}))
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type 'list' with non primitive element", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "list_prop",
			Type:           TypeList,
			ArrayItemsType: TypeObject,
		}
		Convey("When isTerraformListOfSimpleValues method is called", func() {
			isTerraformListOfSimpleValues, listSchema := s.isTerraformListOfSimpleValues()
			Convey("Then the result returned should be the expected one", func() {
				So(isTerraformListOfSimpleValues, ShouldBeFalse)
				So(listSchema, ShouldBeNil)
			})
		})
	})
}

func TestTerraformObjectSchema(t *testing.T) {
	Convey("Given a swagger schema definition that has a property of type 'object'", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "object_prop",
			Type: TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Name: "protocol",
						Type: TypeString,
					},
				},
			},
		}
		Convey("When terraformObjectSchema method is called", func() {
			tfPropSchema, err := s.terraformObjectSchema()
			Convey("Then the resulted tfPropSchema should be of type string too", func() {
				So(err, ShouldBeNil)
				So(reflect.TypeOf(*tfPropSchema), ShouldEqual, reflect.TypeOf(schema.Resource{}))
				So(tfPropSchema.Schema, ShouldContainKey, "protocol")
			})
		})
	})
	Convey("Given a swagger schema definition that has a property of type 'list' and arrays items type object", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "array_prop",
			Type:           TypeList,
			ArrayItemsType: TypeObject,
			ReadOnly:       false,
			Required:       true,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Name: "protocol",
						Type: TypeString,
					},
				},
			},
		}
		Convey("When terraformObjectSchema method is called", func() {
			tfPropSchema, err := s.terraformObjectSchema()
			Convey("Then the resulted tfPropSchema should be of type string too", func() {
				So(err, ShouldBeNil)
				So(reflect.TypeOf(*tfPropSchema), ShouldEqual, reflect.TypeOf(schema.Resource{}))
				So(tfPropSchema.Schema, ShouldContainKey, "protocol")
			})
		})
	})

	Convey("Given a swagger schema definition that has a non supported property type for building object schmea", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "prop",
			Type: TypeString,
		}
		Convey("When terraformObjectSchema method is called", func() {
			_, err := s.terraformObjectSchema()
			Convey("Then the error message returned should match the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "object schema can only be formed for types object or types list with elems of type object: found type='string' elemType='' instead")
			})
		})
	})

	Convey("Given a swagger schema definition that has a object property type for building object schema", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "prop",
			Type: TypeObject,
		}
		Convey("When terraformObjectSchema method is called", func() {
			_, err := s.terraformObjectSchema()
			Convey("Then the error message returned should match the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "missing spec schema definition for property 'prop' of type 'object'")
			})
		})
	})
}

func TestSpecSchemaDefinitionIsPropertyWithNestedObjects(t *testing.T) {
	testcases := []struct {
		name                         string
		schemaDefinitionPropertyType schemaDefinitionPropertyType
		specSchemaDefinition         *SpecSchemaDefinition
		expected                     bool
	}{
		{name: "swagger schema definition property that is not of type 'object'",
			schemaDefinitionPropertyType: TypeBool,
			specSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type: TypeString,
					},
				},
			},
			expected: false},
		{name: "swagger schema definition property that has nested objects",
			schemaDefinitionPropertyType: TypeObject,
			specSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type: TypeObject,
						SpecSchemaDefinition: &SpecSchemaDefinition{
							Properties: SpecSchemaDefinitionProperties{
								&SpecSchemaDefinitionProperty{
									Type: TypeString,
								},
							},
						},
					},
				},
			},
			expected: true},
		{name: "swagger schema definition property that DOES NOT have nested objects",
			specSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type: TypeString,
					},
				},
			},
			expected: false},
		{name: "spec definition property of type object that does not have a corresponding spec schema definition",
			schemaDefinitionPropertyType: TypeObject,
			specSchemaDefinition:         nil,
			expected:                     false},
	}
	for _, tc := range testcases {
		s := &SpecSchemaDefinitionProperty{
			Type:                 tc.schemaDefinitionPropertyType,
			SpecSchemaDefinition: tc.specSchemaDefinition,
		}
		isPropertyWithNestedObjects := s.isPropertyWithNestedObjects()
		assert.Equal(t, tc.expected, isPropertyWithNestedObjects, tc.name)

	}

}

func TestTerraformSchema(t *testing.T) {
	Convey("Given a swagger schema definition that has two nested properties - one being a simple object and the other one a primitive", t, func() {
		expectedNestedObjectPropertyName := "nested_object1"
		s := &SpecSchemaDefinitionProperty{
			Name: "top_level_object",
			Type: TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type: TypeObject,
						Name: expectedNestedObjectPropertyName,
						SpecSchemaDefinition: &SpecSchemaDefinition{
							Properties: SpecSchemaDefinitionProperties{
								&SpecSchemaDefinitionProperty{
									Type: TypeString,
									Name: "string_property_1",
								},
							},
						},
					},
					&SpecSchemaDefinitionProperty{
						Type:          TypeFloat,
						Name:          "nested_float2",
						PreferredName: "nested_float_2",
					},
				},
			}}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulting tfPropSchema should have a top level that is a 1 element list (workaround for object property with nested object)
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				So(tfPropSchema.MaxItems, ShouldEqual, 1)
				// the returned terraform schema contains the 'nested_object1' with the right configuration
				nestedObject1 := tfPropSchema.Elem.(*schema.Resource).Schema["nested_object1"]
				So(nestedObject1, ShouldNotBeNil)
				So(nestedObject1.Type, ShouldEqual, schema.TypeList)
				So(nestedObject1.Elem.(*schema.Resource).Schema["string_property_1"].Type, ShouldEqual, schema.TypeString)
				// the returned terraform schema contains the 'nested_float_2' with the right configuration
				nestedObject2 := tfPropSchema.Elem.(*schema.Resource).Schema["nested_float_2"]
				So(nestedObject2.Type, ShouldEqual, schema.TypeFloat)
			})
		})
	})

	Convey("Given a swagger schema definition that has two nested simple object properties", t, func() {
		expectedNestedObjectPropertyName1 := "nested_object1"
		expectedNestedObjectPropertyName2 := "nested_object2"
		s := &SpecSchemaDefinitionProperty{
			Name: "top_level_object",
			Type: TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type: TypeObject,
						Name: expectedNestedObjectPropertyName1,
						SpecSchemaDefinition: &SpecSchemaDefinition{
							Properties: SpecSchemaDefinitionProperties{
								&SpecSchemaDefinitionProperty{
									Type: TypeString,
									Name: "string_property_1",
								},
							},
						},
					},
					&SpecSchemaDefinitionProperty{
						Type: TypeObject,
						Name: expectedNestedObjectPropertyName2,
						SpecSchemaDefinition: &SpecSchemaDefinition{
							Properties: SpecSchemaDefinitionProperties{
								&SpecSchemaDefinitionProperty{
									Type: TypeString,
									Name: "string_property_2",
								},
							},
						},
					},
				},
			}}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulting tfPropSchema should have a top level that is a 1 element list (workaround for object property with nested object)
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				So(tfPropSchema.MaxItems, ShouldEqual, 1)
				// the returned terraform schema contains the schema for the first nested object property with the right configuration
				nestedObject1 := tfPropSchema.Elem.(*schema.Resource).Schema[expectedNestedObjectPropertyName1]
				So(nestedObject1, ShouldNotBeNil)
				So(nestedObject1.Type, ShouldEqual, schema.TypeList)
				So(nestedObject1.Elem.(*schema.Resource).Schema["string_property_1"].Type, ShouldEqual, schema.TypeString)
				// the returned terraform schema contains the schema for the Second nested object property with the right configuration
				nestedObject2 := tfPropSchema.Elem.(*schema.Resource).Schema[expectedNestedObjectPropertyName2]
				So(nestedObject2, ShouldNotBeNil)
				So(nestedObject2.Type, ShouldEqual, schema.TypeList)
				So(nestedObject2.Elem.(*schema.Resource).Schema["string_property_2"].Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition of type object and a complex object nested into it", t, func() {
		complexObjectName := "complex_object_which_is_nested"
		s := &SpecSchemaDefinitionProperty{
			Name: "top_level_object",
			Type: TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type: TypeObject,
						Name: complexObjectName,
						SpecSchemaDefinition: &SpecSchemaDefinition{
							Properties: SpecSchemaDefinitionProperties{
								&SpecSchemaDefinitionProperty{
									Type: TypeString,
									Name: "string_property",
								},
								&SpecSchemaDefinitionProperty{
									Type:     TypeInt,
									Name:     "int_property",
									ReadOnly: true,
								},
							},
						},
					},
				},
			}}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulting tfPropSchema should have a top level that is a 1 element list (workaround for object property with nested object)
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				So(tfPropSchema.MaxItems, ShouldEqual, 1)
				// the returned terraform schema contains the schema for the first nested AND complex object property with the right configuration --> TypeList
				nestedAndComplexObj := tfPropSchema.Elem.(*schema.Resource).Schema[complexObjectName]
				So(nestedAndComplexObj, ShouldNotBeNil)
				So(nestedAndComplexObj.Type, ShouldEqual, schema.TypeList)
				So(nestedAndComplexObj.MaxItems, ShouldEqual, 1)
				So(nestedAndComplexObj.Elem.(*schema.Resource).Schema["string_property"].Type, ShouldEqual, schema.TypeString)
				So(nestedAndComplexObj.Elem.(*schema.Resource).Schema["int_property"].Type, ShouldEqual, schema.TypeInt)
				So(nestedAndComplexObj.Elem.(*schema.Resource).Schema["int_property"].Computed, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger schema definition that contains a nested object", t, func() {
		complexObjectName := "complex_object_which_is_nested"
		s := &SpecSchemaDefinitionProperty{
			Name: "complex object",
			Type: TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type:                 TypeObject,
						Name:                 complexObjectName,
						ReadOnly:             true,
						SpecSchemaDefinition: &SpecSchemaDefinition{},
					},
				},
			}}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulting tfPropSchema should have mapped the complex object as a 1 element list BECAUSE even if the complex object doesn't have EnableLegacyComplexObjectBlockConfiguration=true being nested == being complex object", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				So(tfPropSchema.MaxItems, ShouldEqual, 1)
				So(tfPropSchema.Elem.(*schema.Resource).Schema[complexObjectName].Computed, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger schema definition that contains a object with no SpecSchemaDefinition", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "object",
			Type: TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type: TypeObject,
						Name: "the object",
					},
				},
			}}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then return an error because of the lack of SpecSchemaDefinition", func() {
				So(err, ShouldNotBeNil)
				So(tfPropSchema, ShouldBeNil)
				So(err.Error(), ShouldEqual, `missing spec schema definition for property 'the object' of type 'object'`)

			})
		})
	})

	Convey("Given a swagger schema definition tha is a complex object", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "top level complex object",
			Type: TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type:                 TypeInt,
						Name:                 "my_int_prop",
						ReadOnly:             true,
						SpecSchemaDefinition: &SpecSchemaDefinition{},
					},
				},
			}}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulting tfPropSchema should have mapped the object as a 1 elem list BECAUSE it has EnableLegacyComplexObjectBlockConfiguration = true", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				So(tfPropSchema.MaxItems, ShouldEqual, 1)
				So(tfPropSchema.Elem.(*schema.Resource).Schema["my_int_prop"].Computed, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger schema definition tha is a simple object (EnableLegacyComplexObjectBlockConfiguration not present or set to false)", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "top level complex object",
			Type: TypeObject,
			//EnableLegacyComplexObjectBlockConfiguration: true, ==> This field is not present or set to false
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Type:                 TypeInt,
						Name:                 "my_int_prop",
						ReadOnly:             true,
						SpecSchemaDefinition: &SpecSchemaDefinition{},
					},
				},
			}}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulting tfPropSchema should match the following using TypeList as type", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				So(tfPropSchema.Elem.(*schema.Resource).Schema["my_int_prop"].Computed, ShouldBeTrue)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'string' which is required", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "string_prop",
			Type:     TypeString,
			ReadOnly: false,
			Required: true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted tfPropSchema should be of type string too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeString)
				So(tfPropSchema.Required, ShouldBeTrue)
			})
		})
	})
	Convey("Given a swagger schema definition that has an unsupported property type", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name: "rune_prop",
			Type: "unsupported",
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be nil", func() {
				So(err.Error(), ShouldEqual, "non supported type unsupported")
				So(tfPropSchema, ShouldBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'integer'", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "int_prop",
			Type:     TypeInt,
			ReadOnly: false,
			Required: true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be of type int too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeInt)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'number'", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "number_prop",
			Type:     TypeFloat,
			ReadOnly: false,
			Required: true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be of type float too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeFloat)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'boolean'", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "boolean_prop",
			Type:     TypeBool,
			ReadOnly: false,
			Required: true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should be of type int too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeBool)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property with a description", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:        "string_prop",
			Type:        TypeString,
			Description: "string_prop description",
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the resulted terraform property schema should contain the description too", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema.Description, ShouldEqual, s.Description)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array' and the elems are type string", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "array_prop",
			Type:           TypeList,
			ArrayItemsType: TypeString,
			ReadOnly:       false,
			Required:       true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulted terraform property schema should be of type array too
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				// the array elements are of type string
				So(reflect.TypeOf(tfPropSchema.Elem).Elem(), ShouldEqual, reflect.TypeOf(schema.Schema{}))
				So(tfPropSchema.Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array' and the elems are type integer", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "array_prop",
			Type:           TypeList,
			ArrayItemsType: TypeInt,
			ReadOnly:       false,
			Required:       true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulted terraform property schema should be of type array too
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				// the array elements are of type int
				So(reflect.TypeOf(tfPropSchema.Elem).Elem(), ShouldEqual, reflect.TypeOf(schema.Schema{}))
				So(tfPropSchema.Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeInt)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array' and the elems are type number", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "array_prop",
			Type:           TypeList,
			ArrayItemsType: TypeFloat,
			ReadOnly:       false,
			Required:       true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulted terraform property schema should be of type array too
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				// the array elements are of type float
				So(reflect.TypeOf(tfPropSchema.Elem).Elem(), ShouldEqual, reflect.TypeOf(schema.Schema{}))
				So(tfPropSchema.Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeFloat)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array' and the elems are type bool", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "array_prop",
			Type:           TypeList,
			ArrayItemsType: TypeBool,
			ReadOnly:       false,
			Required:       true,
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulted terraform property schema should be of type array too
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				// the array elements are of type bool
				So(reflect.TypeOf(tfPropSchema.Elem).Elem(), ShouldEqual, reflect.TypeOf(schema.Schema{}))
				So(tfPropSchema.Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeBool)
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array' and the elems are type object", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:           "array_prop",
			Type:           TypeList,
			ArrayItemsType: TypeObject,
			ReadOnly:       false,
			Required:       true,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Name: "protocol",
						Type: TypeString,
					},
				},
			},
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				// the resulted terraform property schema should be of type array too
				So(err, ShouldBeNil)
				So(tfPropSchema.Type, ShouldEqual, schema.TypeList)
				// the array elements are of type object (resource object) containing the object schema properties
				So(reflect.TypeOf(tfPropSchema.Elem).Elem(), ShouldEqual, reflect.TypeOf(schema.Resource{}))
				So(tfPropSchema.Elem.(*schema.Resource).Schema, ShouldContainKey, "protocol")
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type 'array' and the elems are not set", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "array_prop",
			Type:     TypeList,
			ReadOnly: false,
			Required: true,
		}
		Convey("When terraformSchema method is called", func() {
			_, err := s.terraformSchema()
			Convey("Then the error message returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "object schema can only be formed for types object or types list with elems of type object: found type='list' elemType='' instead")
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type object and a ref pointing to the schema", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "object_prop",
			Type:     TypeObject,
			ReadOnly: false,
			Required: true,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Name: "message",
						Type: TypeString,
					},
				},
			},
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(tfPropSchema, ShouldNotBeNil)
				// the tf resource schema returned should contained the sub property - 'message'
				So(tfPropSchema.Elem.(*schema.Resource).Schema, ShouldContainKey, "message")
			})
		})
	})

	Convey("Given a swagger schema definition that has a property of type object that has nested schema and property named id", t, func() {
		s := &SpecSchemaDefinitionProperty{
			Name:     "object_prop",
			Type:     TypeObject,
			ReadOnly: false,
			Required: true,
			SpecSchemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Name: "id",
						Type: TypeString,
					},
					&SpecSchemaDefinitionProperty{
						Name: "message",
						Type: TypeString,
					},
				},
			},
		}
		Convey("When terraformSchema method is called", func() {
			tfPropSchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the tf resource schema returned should not be nil
				So(tfPropSchema, ShouldNotBeNil)
				// the tf resource schema returned should contain the sub property - 'message'
				So(tfPropSchema.Elem.(*schema.Resource).Schema, ShouldContainKey, "message")
				// the tf resource schema returned should contain the sub property - 'id' and should not be ignored
				So(tfPropSchema.Elem.(*schema.Resource).Schema, ShouldContainKey, "id")
			})
		})
	})

	Convey("Given a string schemaDefinitionProperty that is required, not readOnly, forceNew, sensitive, not immutable and has a default value", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", true, false, false, true, true, false, false, false, "default value")
		Convey("When terraformSchema is called with a schema definition property that is required, force new, sensitive and has a default value", func() {
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(terraformPropertySchema.Optional, ShouldBeFalse)
				So(terraformPropertySchema.Required, ShouldBeTrue)
				So(terraformPropertySchema.Computed, ShouldBeFalse)
				So(terraformPropertySchema.ForceNew, ShouldBeTrue)
				So(terraformPropertySchema.Sensitive, ShouldBeTrue)
				// the schema returned should have a default value (note: terraform will complain in this case at runtime since required properties cannot have default values)
				So(terraformPropertySchema.Default, ShouldEqual, s.Default)
				So(terraformPropertySchema.ValidateDiagFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is readOnly and does not have a default value (meaning the value is not known at plan time)", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, true, true, false, false, false, false, false, "")
		Convey("When terraformSchema is called with a schema definition property that is readonly", func() {
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(terraformPropertySchema.Required, ShouldBeFalse)
				So(terraformPropertySchema.Optional, ShouldBeTrue)
				So(terraformPropertySchema.Computed, ShouldBeTrue)
				So(terraformPropertySchema.ValidateDiagFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is readOnly and does have a default value (meaning the default value is known at plan time)", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, true, true, false, false, false, false, false, "some value")
		Convey("When terraformSchema is called with a schema definition property that is readonly", func() {
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(terraformPropertySchema.Required, ShouldBeFalse)
				So(terraformPropertySchema.Optional, ShouldBeTrue)
				So(terraformPropertySchema.Computed, ShouldBeTrue)
				So(terraformPropertySchema.Default, ShouldBeNil)
				So(terraformPropertySchema.ValidateDiagFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is optional computed and does not have a default value (meaning the value is not known at plan time)", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, false, true, false, false, false, false, false, nil)
		Convey("When terraformSchema is called with a schema definition property that is optional computed", func() {
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(terraformPropertySchema.Required, ShouldBeFalse)
				So(terraformPropertySchema.Optional, ShouldBeTrue)
				So(terraformPropertySchema.Computed, ShouldBeTrue)
				So(terraformPropertySchema.Default, ShouldBeNil)
				So(terraformPropertySchema.ValidateDiagFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is optional computed and does have a default value (meaning the value is known at plan time)", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, false, true, false, false, false, false, false, "some known value")
		Convey("When terraformSchema is called with a schema definition property that is optional computed", func() {
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(terraformPropertySchema.Optional, ShouldBeTrue)
				So(terraformPropertySchema.Required, ShouldBeFalse)
				// the schema returned should not be configured as computed since in this case terraform will make use of the default value as input for the user. This makes the default value visible at plan time
				So(terraformPropertySchema.Computed, ShouldBeFalse)
				So(terraformPropertySchema.Default, ShouldNotBeNil)
				So(terraformPropertySchema.ValidateDiagFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is forceNew and immutable ", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, false, false, true, false, true, false, false, "")
		Convey("When terraformSchema is called with a schema definition property that validation fails due to immutable and forceNew set", func() {
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(terraformPropertySchema.ValidateDiagFunc, ShouldNotBeNil)
				// the schema validate function should return an error
				diagnostics := terraformPropertySchema.ValidateDiagFunc(nil, cty.Path{})
				So(diagnostics, ShouldNotBeNil)
				So(diagnostics, ShouldNotBeEmpty)
				So(diagnostics[0].Summary, ShouldContainSubstring, "property 'propertyName' is configured as immutable and can not be configured with forceNew too")
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is readOnly and required", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", true, true, false, false, false, false, false, false, nil)
		Convey("When terraformSchema is called with a schema definition property that validation fails due to required and computed set", func() {
			terraformPropertySchema, err := s.terraformSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(terraformPropertySchema.Required, ShouldBeTrue)
				So(terraformPropertySchema.Computed, ShouldBeFalse)
				So(terraformPropertySchema.ValidateDiagFunc, ShouldNotBeNil)
				diagnostics := terraformPropertySchema.ValidateDiagFunc(nil, cty.Path{})
				So(diagnostics, ShouldNotBeNil)
				So(diagnostics, ShouldNotBeEmpty)
				So(diagnostics[0].Summary, ShouldContainSubstring, "property 'propertyName' is configured as required and can not be configured as computed too")
			})
		})
	})
}

func TestValidateDiagFunc(t *testing.T) {

	Convey("Given a schemaDefinitionProperty that is computed and has a default value set", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, false, false, "defaultValue")
		Convey("When validateDiagFunc is called with a schema definition property", func() {
			Convey("Then the result returned should be the expected one", func() {
				So(s.validateDiagFunc(), ShouldNotBeNil)
				diagnostics := s.validateDiagFunc()(nil, cty.Path{})
				So(diagnostics, ShouldBeNil)
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is forceNew and immutable ", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, false, false, true, false, true, false, false, "")
		Convey("When validateDiagFunc is called with a schema definition property", func() {
			Convey("Then the result returned should be the expected one", func() {
				So(s.validateDiagFunc(), ShouldNotBeNil)
				diagnostics := s.validateDiagFunc()(nil, cty.Path{})
				So(diagnostics, ShouldNotBeNil)
				So(diagnostics, ShouldNotBeEmpty)
				So(diagnostics[0].Summary, ShouldContainSubstring, "property 'propertyName' is configured as immutable and can not be configured with forceNew too")
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is computed and required", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", true, true, false, false, false, false, false, false, nil)
		Convey("When validateDiagFunc is called with a schema definition property", func() {
			Convey("Then the result returned should be the expected one", func() {
				So(s.validateDiagFunc(), ShouldNotBeNil)
				diagnostics := s.validateDiagFunc()(nil, cty.Path{})
				So(diagnostics, ShouldNotBeNil)
				So(diagnostics, ShouldNotBeEmpty)
				So(diagnostics[0].Summary, ShouldContainSubstring, "property 'propertyName' is configured as required and can not be configured as computed too")
			})
		})
	})
}

func TestValidateFunc(t *testing.T) {

	Convey("Given a schemaDefinitionProperty that is computed and has a default value set", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, false, false, "defaultValue")
		Convey("When validateFunc is called with a schema definition property", func() {
			Convey("Then the result returned should be the expected one", func() {
				So(s.validateFunc(), ShouldNotBeNil)
				_, err := s.validateFunc()(nil, "")
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is forceNew and immutable ", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", false, false, false, true, false, true, false, false, "")
		Convey("When validateFunc is called with a schema definition property", func() {
			Convey("Then the result returned should be the expected one", func() {
				So(s.validateFunc(), ShouldNotBeNil)
				_, err := s.validateFunc()(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "property 'propertyName' is configured as immutable and can not be configured with forceNew too")
			})
		})
	})

	Convey("Given a schemaDefinitionProperty that is computed and required", t, func() {
		s := newStringSchemaDefinitionProperty("propertyName", "", true, true, false, false, false, false, false, false, nil)
		Convey("When validateFunc is called with a schema definition property", func() {
			Convey("Then the result returned should be the expected one", func() {
				So(s.validateFunc(), ShouldNotBeNil)
				_, err := s.validateFunc()(nil, "")
				So(err, ShouldNotBeNil)
				So(err, ShouldNotBeEmpty)
				So(err[0].Error(), ShouldContainSubstring, "property 'propertyName' is configured as required and can not be configured as computed too")
			})
		})
	})
}

func TestEqualItems(t *testing.T) {
	testCases := []struct {
		name               string
		schemaDefProp      SpecSchemaDefinitionProperty
		propertyType       schemaDefinitionPropertyType
		arrayItemsPropType schemaDefinitionPropertyType
		inputItem          interface{}
		remoteItem         interface{}
		expectedOutput     bool
	}{
		// String use cases
		{
			name:           "string input value matches string remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeString},
			inputItem:      "inputVal1",
			remoteItem:     "inputVal1",
			expectedOutput: true,
		},
		{
			name:           "string input value doesn't match string remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeString},
			inputItem:      "inputVal1",
			remoteItem:     "inputVal2",
			expectedOutput: false,
		},
		// Integer use cases
		{
			name:           "int input value matches int remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeInt},
			inputItem:      1,
			remoteItem:     1,
			expectedOutput: true,
		},
		{
			name:           "int input value doesn't match int remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeInt},
			inputItem:      1,
			remoteItem:     2,
			expectedOutput: false,
		},
		// Float use cases
		{
			name:           "float input value matches float remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeFloat},
			inputItem:      1.0,
			remoteItem:     1.0,
			expectedOutput: true,
		},
		{
			name:           "float input value doesn't match float remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeFloat},
			inputItem:      1.0,
			remoteItem:     2.0,
			expectedOutput: false,
		},
		// Bool use cases
		{
			name:           "bool input value matches bool remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeBool},
			inputItem:      true,
			remoteItem:     true,
			expectedOutput: true,
		},
		{
			name:           "bool input value doesn't match bool remote value",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeBool},
			inputItem:      true,
			remoteItem:     false,
			expectedOutput: false,
		},
		// List use cases
		{
			name: "list input value matches list remote value",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type:           TypeList,
				ArrayItemsType: TypeString,
			},
			inputItem:      []interface{}{"role1", "role2"},
			remoteItem:     []interface{}{"role1", "role2"},
			expectedOutput: true,
		},
		{
			name: "list input value doesn't match list remote value (same list length)",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type:           TypeList,
				ArrayItemsType: TypeString,
			},
			inputItem:      []interface{}{"role1", "role2"},
			remoteItem:     []interface{}{"role3", "role4"},
			expectedOutput: false,
		},
		{
			name: "list input value doesn't match list remote value (same list length band same items but unordered) but property is marked as ignore order",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputItem:      []interface{}{"role1", "role2"},
			remoteItem:     []interface{}{"role2", "role1"},
			expectedOutput: true,
		},
		{
			name: "list input value doesn't match list remote value (same list length band) but property is marked as ignore order",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputItem:      []interface{}{"role1", "role2"},
			remoteItem:     []interface{}{"role3", "role1"},
			expectedOutput: false,
		},
		{
			name: "list input value doesn't match list remote value (different list length)",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type:           TypeList,
				ArrayItemsType: TypeString,
			},
			inputItem:      []interface{}{"role1", "role2"},
			remoteItem:     []interface{}{"role1"},
			expectedOutput: false,
		},
		// Object use cases
		{
			name: "object input value matches object remote value",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type: TypeObject,
				SpecSchemaDefinition: &SpecSchemaDefinition{
					Properties: SpecSchemaDefinitionProperties{
						&SpecSchemaDefinitionProperty{
							Name: "group",
							Type: TypeString,
						},
					},
				},
			},
			inputItem:      map[string]interface{}{"group": "someGroup"},
			remoteItem:     map[string]interface{}{"group": "someGroup"},
			expectedOutput: true,
		},
		{
			name: "object input value doesn't match object remote value",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type: TypeObject,
				SpecSchemaDefinition: &SpecSchemaDefinition{
					Properties: SpecSchemaDefinitionProperties{
						&SpecSchemaDefinitionProperty{
							Name: "group",
							Type: TypeString,
						},
					},
				},
			},
			inputItem:      map[string]interface{}{"group": "someGroup"},
			remoteItem:     map[string]interface{}{"group": "someOtherGroup"},
			expectedOutput: false,
		},
		// Negative cases
		{
			name:           "string input value is not a string",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeString},
			inputItem:      1,
			remoteItem:     "inputVal1",
			expectedOutput: false,
		},
		{
			name:           "int input value is not an int",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeInt},
			inputItem:      "not_an_int",
			remoteItem:     1,
			expectedOutput: false,
		},
		{
			name:           "float input value is not a float",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeFloat},
			inputItem:      1.0,
			remoteItem:     "not_an_float",
			expectedOutput: false,
		},
		{
			name:           "bool input value is nto a bool",
			schemaDefProp:  SpecSchemaDefinitionProperty{Type: TypeBool},
			inputItem:      true,
			remoteItem:     "not_a_bool",
			expectedOutput: false,
		},
		{
			name: "list input value is not a list",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type:           TypeList,
				ArrayItemsType: TypeString,
			},
			inputItem:      []interface{}{"role1", "role2"},
			remoteItem:     "not a list",
			expectedOutput: false,
		},
		{
			name: "object input value is not an object",
			schemaDefProp: SpecSchemaDefinitionProperty{
				Type: TypeObject,
				SpecSchemaDefinition: &SpecSchemaDefinition{
					Properties: SpecSchemaDefinitionProperties{
						&SpecSchemaDefinitionProperty{
							Name: "group",
							Type: TypeString,
						},
					},
				},
			},
			inputItem:      "not_an_object",
			remoteItem:     map[string]interface{}{"group": "someGroup"},
			expectedOutput: false,
		},
	}

	for _, tc := range testCases {
		output := tc.schemaDefProp.equal(tc.inputItem, tc.remoteItem)
		assert.Equal(t, tc.expectedOutput, output, tc.name)
	}
}

func TestValidateValueType(t *testing.T) {
	testCases := []struct {
		name           string
		item           interface{}
		itemKind       reflect.Kind
		expectedOutput bool
	}{
		// String use cases
		{
			name:           "expect string kind and item is a string",
			item:           "inputVal1",
			itemKind:       reflect.String,
			expectedOutput: true,
		},
		{
			name:           "expect string kind and item is NOT a string",
			item:           1,
			itemKind:       reflect.String,
			expectedOutput: false,
		},
		// Int use cases
		{
			name:           "expect int kind and item is a int",
			item:           1,
			itemKind:       reflect.Int,
			expectedOutput: true,
		},
		{
			name:           "expect int kind and item is NOT a int",
			item:           "not an int",
			itemKind:       reflect.Int,
			expectedOutput: false,
		},
		// Float use cases
		{
			name:           "expect float kind and item is a float",
			item:           1.0,
			itemKind:       reflect.Float64,
			expectedOutput: true,
		},
		{
			name:           "expect float kind and item is NOT a float",
			item:           "not a float",
			itemKind:       reflect.Float64,
			expectedOutput: false,
		},
		// Bool use cases
		{
			name:           "expect bool kind and item is a bool",
			item:           true,
			itemKind:       reflect.Bool,
			expectedOutput: true,
		},
		{
			name:           "expect bool kind and item is NOT a bool",
			item:           "not a bool",
			itemKind:       reflect.Bool,
			expectedOutput: false,
		},
		//  List use cases
		{
			name:           "expect slice kind and item is a slice",
			item:           []interface{}{"item1", "item2"},
			itemKind:       reflect.Slice,
			expectedOutput: true,
		},
		{
			name:           "expect slice kind and item is NOT a slice",
			item:           "not a slice",
			itemKind:       reflect.Slice,
			expectedOutput: false,
		},
		//  Object use cases
		{
			name:           "expect map kind and item is a map",
			item:           map[string]interface{}{"group": "someGroup"},
			itemKind:       reflect.Map,
			expectedOutput: true,
		},
		{
			name:           "expect map kind and item is NOT a map",
			item:           "not a map",
			itemKind:       reflect.Map,
			expectedOutput: false,
		},
	}
	for _, tc := range testCases {
		s := SpecSchemaDefinitionProperty{}
		output := s.validateValueType(tc.item, tc.itemKind)
		assert.Equal(t, tc.expectedOutput, output, tc.name)
	}
}

func Test_shouldIgnoreOrder(t *testing.T) {
	Convey("Given a scjema definition property that is a list and configured with ignore order", t, func() {
		p := &SpecSchemaDefinitionProperty{
			Type:             TypeList,
			IgnoreItemsOrder: true,
		}
		Convey("When shouldIgnoreOrder is called", func() {
			shouldIgnoreOrder := p.shouldIgnoreOrder()
			Convey("Then the err returned should be true", func() {
				So(shouldIgnoreOrder, ShouldBeTrue)
			})
		})
	})
	Convey("Given a scjema definition property that is NOT a list", t, func() {
		p := &SpecSchemaDefinitionProperty{
			Type: TypeString,
		}
		Convey("When shouldIgnoreOrder is called", func() {
			shouldIgnoreOrder := p.shouldIgnoreOrder()
			Convey("Then the err returned should be false", func() {
				So(shouldIgnoreOrder, ShouldBeFalse)
			})
		})
	})
	Convey("Given a scjema definition property that is a list but the ignore order is set to false", t, func() {
		p := &SpecSchemaDefinitionProperty{
			Type:             TypeList,
			IgnoreItemsOrder: false,
		}
		Convey("When shouldIgnoreOrder is called", func() {
			shouldIgnoreOrder := p.shouldIgnoreOrder()
			Convey("Then the err returned should be false", func() {
				So(shouldIgnoreOrder, ShouldBeFalse)
			})
		})
	})
}

func Test_shouldUseLegacyTerraformSDKBlockApproachForComplexObjects(t *testing.T) {

	Convey("Given a SpecSchemaDefinitionProperty that is not of TypeObject", t, func() {
		p := &SpecSchemaDefinitionProperty{
			Type: TypeString,
		}
		Convey("When shouldUseLegacyTerraformSDKBlockApproachForComplexObjects is called", func() {
			shouldUseLegacyTerraformSDKBlockApproachForComplexObjects := p.shouldUseLegacyTerraformSDKBlockApproachForComplexObjects()
			Convey("Then the value returned should be false", func() {
				So(shouldUseLegacyTerraformSDKBlockApproachForComplexObjects, ShouldBeFalse)
			})
		})
	})

	Convey("Given a SpecSchemaDefinitionProperty that is of TypeObject", t, func() {
		p := &SpecSchemaDefinitionProperty{
			Type: TypeObject,
		}
		Convey("When shouldUseLegacyTerraformSDKBlockApproachForComplexObjects is called", func() {
			shouldUseLegacyTerraformSDKBlockApproachForComplexObjects := p.shouldUseLegacyTerraformSDKBlockApproachForComplexObjects()
			Convey("Then the value returned should be true", func() {
				So(shouldUseLegacyTerraformSDKBlockApproachForComplexObjects, ShouldBeTrue)
			})
		})
	})
}
