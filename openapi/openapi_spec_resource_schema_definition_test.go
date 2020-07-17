package openapi

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stretchr/testify/assert"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConvertToDataSourceSpecSchemaDefinitionProperty(t *testing.T) {
	s := SpecSchemaDefinition{
		Properties: []*SpecSchemaDefinitionProperty{
			{
				Type:     TypeString,
				Required: true,
				Computed: false,
				Default:  "defaultValue",
			},
		},
	}
	dataSourceSpecSchemaDef := s.ConvertToDataSourceSpecSchemaDefinition()
	prop := dataSourceSpecSchemaDef.Properties[0]
	assert.False(t, prop.Required)
	assert.True(t, prop.Computed)
	assert.Nil(t, prop.Default)
}

func TestConvertToDataSourceSpecSchemaDefinition_WithParentProp(t *testing.T) {
	s := SpecSchemaDefinition{
		Properties: []*SpecSchemaDefinitionProperty{
			{
				Name:             "someParentProp",
				Type:             TypeString,
				Required:         true,
				Computed:         false,
				Default:          "defaultParentPropValue",
				IsParentProperty: true,
			},
			{
				Name:     "someProp",
				Type:     TypeObject,
				Required: true,
				Computed: false,
				Default:  "defaultPropValue",
			},
		},
	}
	dataSourceSpecSchemaDef := s.ConvertToDataSourceSpecSchemaDefinition()
	parentProp := dataSourceSpecSchemaDef.Properties[0]
	assert.True(t, parentProp.Required)
	assert.False(t, parentProp.Computed)
	assert.Equal(t, "defaultParentPropValue", parentProp.Default)
	prop := dataSourceSpecSchemaDef.Properties[1]
	assert.False(t, prop.Required)
	assert.True(t, prop.Computed)
	assert.Nil(t, prop.Default)
}

func TestConvertToDataSourceSpecSchemaDefinitionProperty_ObjectProp(t *testing.T) {
	s := SpecSchemaDefinition{
		Properties: []*SpecSchemaDefinitionProperty{
			{
				Type:     TypeObject,
				Required: true,
				Computed: false,
				Default:  "defaultValue",
				SpecSchemaDefinition: &SpecSchemaDefinition{
					Properties: []*SpecSchemaDefinitionProperty{
						{
							Type:     TypeString,
							Required: true,
							Computed: false,
							Default:  "defaultValue",
						},
					},
				},
			},
		},
	}
	dataSourceSpecSchemaDef := s.ConvertToDataSourceSpecSchemaDefinition()
	prop := dataSourceSpecSchemaDef.Properties[0]
	assert.False(t, prop.Required)
	assert.True(t, prop.Computed)
	assert.Nil(t, prop.Default)
	objProp := prop.SpecSchemaDefinition.Properties[0]
	assert.False(t, objProp.Required)
	assert.True(t, objProp.Computed)
	assert.Nil(t, objProp.Default)
}

func TestCreateDataSourceSchema(t *testing.T) {

	testCases := []struct {
		name           string
		specSchemaDef  SpecSchemaDefinition
		expectedResult map[string]*schema.Schema
		expectedError  error
	}{
		{
			name: "happy path -- data source schema contains all schema properties configured as computed",
			specSchemaDef: SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Name:     "id",
						Type:     TypeString,
						ReadOnly: false,
						Required: true,
					},
					&SpecSchemaDefinitionProperty{
						Name:     "string_prop",
						Type:     TypeString,
						ReadOnly: false,
						Required: true,
					},
					&SpecSchemaDefinitionProperty{
						Name:     "int_prop",
						Type:     TypeInt,
						ReadOnly: false,
						Required: true,
					},
				},
			},
			expectedResult: map[string]*schema.Schema{
				"string_prop": {
					Required: false,
					Optional: true,
					Computed: true,
					Type:     schema.TypeString,
				},
				"int_prop": {
					Required: false,
					Optional: true,
					Computed: true,
					Type:     schema.TypeInt,
				},
			},
			expectedError: nil,
		},
		{
			name:           "sad path -- a terraform schema cannot be created from a SpecSchemaDefinition ",
			specSchemaDef:  SpecSchemaDefinition{Properties: SpecSchemaDefinitionProperties{&SpecSchemaDefinitionProperty{}}},
			expectedResult: nil,
			expectedError:  errors.New("non supported type "),
		},
	}

	for _, tc := range testCases {
		s, err := tc.specSchemaDef.createDataSourceSchema()
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
			for expectedTerraformPropName, expectedTerraformSchemaProp := range tc.expectedResult {
				assert.Nil(t, s["id"])
				assertDataSourceSchemaProperty(t, s[expectedTerraformPropName], expectedTerraformSchemaProp.Type, tc.name)
			}
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func TestCreateDataSourceSchema_Subresources(t *testing.T) {
	t.Run("happy path -- data source schema dor sub-resoruce contains all schema properties configured as computed but parent properties", func(t *testing.T) {

		specSchemaDef := SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:             "parent_id",
					Type:             TypeString,
					ReadOnly:         false,
					Required:         true,
					IsParentProperty: true,
				},
			},
		}

		// get the Terraform schema which represents a Data Source
		s, err := specSchemaDef.createDataSourceSchema()

		assert.NotNil(t, s)
		assert.NoError(t, err)

		assert.Equal(t, schema.TypeString, s["parent_id"].Type)
		assert.True(t, s["parent_id"].Required)
		assert.False(t, s["parent_id"].Optional)
		assert.False(t, s["parent_id"].Computed)
	})
}

func TestCreateDataSourceSchema_ForNestedObjects(t *testing.T) {
	t.Run("happy path -- a data soruce can be derived from a nested object keeping all the properies attributes as expected", func(t *testing.T) {
		// set up the schema for the nested object
		nestedObjectSchemaDefinition := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		nestedObjectDefault := map[string]interface{}{
			"origin_port": nestedObjectSchemaDefinition.Properties[0].Default,
			"protocol":    nestedObjectSchemaDefinition.Properties[1].Default,
		}
		nestedObject := newObjectSchemaDefinitionPropertyWithDefaults("nested_object", "", true, false, false, nestedObjectDefault, nestedObjectSchemaDefinition)
		propertyWithNestedObjectSchemaDefinition := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				idProperty,
				nestedObject,
			},
		}
		dataValue := map[string]interface{}{
			"id":            propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
			"nested_object": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
		}
		dataSourceSchema := newObjectSchemaDefinitionPropertyWithDefaults("nested-oobj", "", true, false, false, dataValue, propertyWithNestedObjectSchemaDefinition)
		specSchemaNested := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{dataSourceSchema},
		}

		// get the Terraform schema which represents a Data Source
		s, err := specSchemaNested.createDataSourceSchema()

		assert.NotNil(t, s)
		assert.NoError(t, err)

		// assertions on the properties attributes
		assertDataSourceSchemaProperty(t, s["nested_oobj"], schema.TypeList)

		// 1^ level
		objectResource := s["nested_oobj"].Elem.(*schema.Resource)
		assert.Equal(t, 2, len(objectResource.Schema))

		assertDataSourceSchemaProperty(t, objectResource.Schema["id"], schema.TypeString)
		assertDataSourceSchemaProperty(t, objectResource.Schema["nested_object"], schema.TypeMap)

		// 2^ level
		nestedObjectResource := objectResource.Schema["nested_object"].Elem.(*schema.Resource)
		assert.Equal(t, 2, len(nestedObjectResource.Schema))
		assertDataSourceSchemaProperty(t, nestedObjectResource.Schema["origin_port"], schema.TypeInt)
		assertDataSourceSchemaProperty(t, nestedObjectResource.Schema["protocol"], schema.TypeString)
	})
}

func TestCreateResourceSchema(t *testing.T) {
	Convey("Given a swagger schema definition that has few properties including the id all with terraform compliant names", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:     "string_prop",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(tfResourceSchema, ShouldNotContainKey, "id")
				So(tfResourceSchema, ShouldContainKey, "string_prop")
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including the id all with NON terraform compliant names", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "ID",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:     "stringProp",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(tfResourceSchema, ShouldNotContainKey, "id")
				So(tfResourceSchema, ShouldContainKey, "string_prop")
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including an object property with internal ID field", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:     "string_prop",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     TypeObject,
					ReadOnly: true,
					SpecSchemaDefinition: &SpecSchemaDefinition{
						Properties: SpecSchemaDefinitionProperties{
							&SpecSchemaDefinitionProperty{
								Name:               "id",
								Type:               TypeString,
								ReadOnly:           true,
								IsStatusIdentifier: true,
							},
							&SpecSchemaDefinitionProperty{
								Name:               "name",
								Type:               TypeString,
								ReadOnly:           true,
								IsStatusIdentifier: true,
							},
						},
					},
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(tfResourceSchema, ShouldNotContainKey, "id")
				So(tfResourceSchema, ShouldContainKey, "string_prop")
				So(tfResourceSchema["string_prop"].Type, ShouldEqual, schema.TypeString)
				So(tfResourceSchema, ShouldContainKey, statusDefaultPropertyName)
				So(tfResourceSchema[statusDefaultPropertyName].Type, ShouldEqual, schema.TypeMap)
				So(tfResourceSchema[statusDefaultPropertyName].Elem.(*schema.Resource).Schema, ShouldContainKey, "id")
				So(tfResourceSchema[statusDefaultPropertyName].Elem.(*schema.Resource).Schema, ShouldContainKey, "name")
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including an array of string primitives", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:           "listeners",
					Type:           TypeList,
					ArrayItemsType: TypeString,
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(tfResourceSchema, ShouldNotContainKey, "id")
				So(tfResourceSchema, ShouldContainKey, "listeners")
				So(tfResourceSchema["listeners"].Type, ShouldEqual, schema.TypeList)
				So(tfResourceSchema["listeners"].Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including an array of objects", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: false,
					Required: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:           "listeners",
					Type:           TypeList,
					ArrayItemsType: TypeObject,
					SpecSchemaDefinition: &SpecSchemaDefinition{
						Properties: SpecSchemaDefinitionProperties{
							&SpecSchemaDefinitionProperty{
								Name: "protocol",
								Type: TypeString,
							},
						},
					},
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(tfResourceSchema, ShouldNotContainKey, "id")
				So(tfResourceSchema, ShouldContainKey, "listeners")
				So(tfResourceSchema["listeners"].Type, ShouldEqual, schema.TypeList)
				So(tfResourceSchema["listeners"].Elem.(*schema.Resource).Schema, ShouldContainKey, "protocol")
				So(tfResourceSchema["listeners"].Elem.(*schema.Resource).Schema["protocol"].Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition that has a combination of properties supported", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				idProperty,
				stringProperty,
				intProperty,
				numberProperty,
				boolProperty,
				slicePrimitiveProperty,
				computedProperty,
				optionalProperty,
				sensitiveProperty,
				forceNewProperty,
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the schema for the resource should contain the expected attributes
				So(tfResourceSchema, ShouldContainKey, stringProperty.Name)
				So(tfResourceSchema, ShouldContainKey, computedProperty.Name)
				So(tfResourceSchema, ShouldContainKey, intProperty.Name)
				So(tfResourceSchema, ShouldContainKey, numberProperty.Name)
				So(tfResourceSchema, ShouldContainKey, boolProperty.Name)
				So(tfResourceSchema, ShouldContainKey, slicePrimitiveProperty.Name)
				So(tfResourceSchema, ShouldContainKey, optionalProperty.Name)
				So(tfResourceSchema, ShouldContainKey, sensitiveProperty.Name)
				So(tfResourceSchema, ShouldContainKey, forceNewProperty.Name)
				// the schema property types should match the expected configuration
				So(tfResourceSchema[stringProperty.Name].Type, ShouldEqual, schema.TypeString)
				So(tfResourceSchema[intProperty.Name].Type, ShouldEqual, schema.TypeInt)
				So(tfResourceSchema[numberProperty.Name].Type, ShouldEqual, schema.TypeFloat)
				So(tfResourceSchema[boolProperty.Name].Type, ShouldEqual, schema.TypeBool)
				So(tfResourceSchema[slicePrimitiveProperty.Name].Type, ShouldEqual, schema.TypeList)
				// the schema property options should match the expected configuration
				So(tfResourceSchema[computedProperty.Name].Computed, ShouldBeTrue)
				So(tfResourceSchema[optionalProperty.Name].Optional, ShouldBeTrue)
				So(tfResourceSchema[sensitiveProperty.Name].Sensitive, ShouldBeTrue)
				So(tfResourceSchema[forceNewProperty.Name].ForceNew, ShouldBeTrue)
			})
		})
	})
}

func TestGetImmutableProperties(t *testing.T) {
	Convey("Given resource info is configured with schemaDefinition that contains a property 'immutable_property' that is immutable", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:      "id",
					Type:      TypeString,
					ReadOnly:  false,
					Immutable: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:      "immutable_property",
					Type:      TypeString,
					ReadOnly:  false,
					Immutable: true,
				},
			},
		}
		Convey("When getImmutableProperties method is called", func() {
			immutableProperties := s.getImmutableProperties()
			Convey("Then the result returned should be the expected one", func() {
				So(immutableProperties, ShouldContain, "immutable_property")
				So(immutableProperties, ShouldNotContain, "id")
			})
		})
	})

	Convey("Given resource info is configured with schemaDefinition that DOES NOT contain immutable properties", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: false,
				},
				&SpecSchemaDefinitionProperty{
					Name:      "mutable_property",
					Type:      TypeString,
					ReadOnly:  false,
					Immutable: false,
				},
			},
		}
		Convey("When getImmutableProperties method is called", func() {
			immutableProperties := s.getImmutableProperties()
			Convey("Then the array returned should be empty", func() {
				So(immutableProperties, ShouldBeEmpty)
			})
		})
	})

}

func TestGetResourceIdentifier(t *testing.T) {
	Convey("Given a SpecSchemaDefinition containing a field named id", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			id, err := s.getResourceIdentifier()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(id, ShouldEqual, s.Properties[0].Name)
			})
		})
	})

	Convey("Given a SpecSchemaDefinition that does not contain a field named id but has a property tagged as IsIdentifier", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:         "someOtherID",
					Type:         TypeString,
					ReadOnly:     true,
					IsIdentifier: true,
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			id, err := s.getResourceIdentifier()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(id, ShouldEqual, s.Properties[0].Name)
			})
		})
	})

	Convey("Given a SpecSchemaDefinition not containing a field named id nor tagged as identifier", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "someOtherField",
					Type:     TypeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			_, err := s.getResourceIdentifier()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "could not find any identifier property in the resource schema definition")
			})
		})
	})
}

func TestGetStatusId(t *testing.T) {
	Convey("Given a SpecSchemaDefinition that has an status property that is not an object", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     TypeString,
					ReadOnly: true,
				},
			},
		}

		Convey("When getStatusIdentifier method is called", func() {
			statuses, err := s.getStatusIdentifier()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(statuses, ShouldNotBeEmpty)
				So(statuses[0], ShouldEqual, statusDefaultPropertyName)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with IsStatusIdentifier set to TRUE", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               TypeString,
					ReadOnly:           true,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(status, ShouldNotBeEmpty)
				So(status[0], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that HAS BOTH an 'status' property AND ALSO a property configured with 'x-terraform-field-status' set to true", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     TypeString,
					ReadOnly: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               TypeString,
					ReadOnly:           true,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(status, ShouldNotBeEmpty)
				So(status[0], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that HAS an status field which is an object containing a status field", t, func() {
		expectedStatusProperty := "actualStatus"
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     "id",
					Type:     TypeString,
					ReadOnly: true,
				},
				&SpecSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     TypeObject,
					ReadOnly: true,
					SpecSchemaDefinition: &SpecSchemaDefinition{
						Properties: SpecSchemaDefinitionProperties{
							&SpecSchemaDefinitionProperty{
								Name:               expectedStatusProperty,
								Type:               TypeString,
								ReadOnly:           true,
								IsStatusIdentifier: true,
							},
						},
					},
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(status, ShouldNotBeEmpty)
				// the status array returned should contain the different the trace of property names (hierarchy) until the last one which is the one used as status, in this case 'actualStatus' on the last index
				So(status[0], ShouldEqual, "status")
				So(status[1], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with 'x-terraform-field-status' set to FALSE", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               TypeString,
					ReadOnly:           true,
					IsStatusIdentifier: false,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := s.getStatusIdentifier()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition that NEITHER HAS an 'status' property NOR a property configured with 'x-terraform-field-status' set to true", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:               "prop-that-is-not-status",
					Type:               TypeString,
					ReadOnly:           true,
					IsStatusIdentifier: false,
				},
				&SpecSchemaDefinitionProperty{
					Name:               "prop-that-is-not-status-and-does-not-have-status-metadata-either",
					Type:               TypeString,
					ReadOnly:           true,
					IsStatusIdentifier: false,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := s.getStatusIdentifier()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a swagger schema definition with a status property that is not readonly", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     TypeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			_, err := s.getStatusIdentifier()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

}

func TestGetStatusIdentifierFor(t *testing.T) {
	Convey("Given a swagger schema definition with a property configured with 'x-terraform-field-status' set to true and it is not readonly", t, func() {
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:               statusDefaultPropertyName,
					Type:               TypeString,
					ReadOnly:           false,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifierFor method is called with a schema definition and forceReadOnly check is disabled (this happens when the method is called recursively)", func() {
			status, err := s.getStatusIdentifierFor(s, true, false)
			Convey("Then the status array returned should contain the status property even though it's not read only...readonly checks are only perform on root level properties", func() {
				So(err, ShouldBeNil)
				So(status[0], ShouldEqual, "status")
			})
		})
	})
}

func TestGetProperty(t *testing.T) {
	Convey("Given a SpecSchemaDefinition", t, func() {
		existingPropertyName := "existingPropertyName"
		s := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				&SpecSchemaDefinitionProperty{
					Name:     existingPropertyName,
					Type:     TypeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getProperty method is called with an existing property name", func() {
			property, err := s.getProperty(existingPropertyName)
			Convey("Then the property returned should be the same as the one in the SpecSchemaDefinition", func() {
				So(err, ShouldBeNil)
				So(property, ShouldEqual, s.Properties[0])
			})
		})
		Convey("When getProperty method is called with a NON existing property name", func() {
			_, err := s.getProperty("nonExistingPropertyName")
			Convey("Then the property returned should be the same as the one in the SpecSchemaDefinition", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "property with name 'nonExistingPropertyName' not existing in resource schema definition")
			})
		})
	})
}

func TestGetPropertyBasedOnTerraformName(t *testing.T) {
	existingPropertyName := "existingPropertyName"
	s := &SpecSchemaDefinition{
		Properties: SpecSchemaDefinitionProperties{
			&SpecSchemaDefinitionProperty{
				Name:     existingPropertyName,
				Type:     TypeString,
				ReadOnly: false,
			},
		},
	}
	_, err := s.getPropertyBasedOnTerraformName("badTerraformPropertyName")
	assert.EqualError(t, err, "property with terraform name 'badTerraformPropertyName' not existing in resource schema definition")

}
