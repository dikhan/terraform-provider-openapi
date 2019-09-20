package openapi

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/stretchr/testify/assert"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateDataSourceSchema(t *testing.T) {

	testCases := []struct {
		name           string
		specSchemaDef  specSchemaDefinition
		expectedResult map[string]*schema.Schema
		expectedError  error
	}{
		{
			name: "happy path -- data source schema contains all schema properties configured as computed",
			specSchemaDef: specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					&specSchemaDefinitionProperty{
						Name:     "id",
						Type:     typeString,
						ReadOnly: false,
						Required: true,
					},
					&specSchemaDefinitionProperty{
						Name:     "string_prop",
						Type:     typeString,
						ReadOnly: false,
						Required: true,
					},
					&specSchemaDefinitionProperty{
						Name:     "int_prop",
						Type:     typeInt,
						ReadOnly: false,
						Required: true,
					},
				},
			},
			expectedResult: map[string]*schema.Schema{
				"string_prop": &schema.Schema{
					Required: false,
					Optional: true,
					Computed: true,
					Type:     schema.TypeString,
				},
				"int_prop": &schema.Schema{
					Required: false,
					Optional: true,
					Computed: true,
					Type:     schema.TypeInt,
				},
			},
			expectedError: nil,
		},
		{
			name:           "sad path -- a terraform schema cannot be created from a specSchemaDefinition ",
			specSchemaDef:  specSchemaDefinition{Properties: specSchemaDefinitionProperties{&specSchemaDefinitionProperty{}}},
			expectedResult: nil,
			expectedError:  errors.New("non supported type "),
		},
	}

	for _, tc := range testCases {
		s, err := tc.specSchemaDef.createDataSourceSchema()
		if tc.expectedError == nil {
			assert.Nil(t, err, tc.name)
			for expectedTerraformPropName, tfSchemaProp := range tc.expectedResult {
				assert.Contains(t, s, expectedTerraformPropName, tc.name)
				assert.Nil(t, s["id"])
				assert.Equal(t, tfSchemaProp.Type, s[expectedTerraformPropName].Type, tc.name)
				assert.Equal(t, tfSchemaProp.Optional, s[expectedTerraformPropName].Optional, tc.name)
				assert.Equal(t, tfSchemaProp.Required, s[expectedTerraformPropName].Required, tc.name)
				assert.Equal(t, tfSchemaProp.Computed, s[expectedTerraformPropName].Computed, tc.name)
			}
		} else {
			assert.Equal(t, tc.expectedError.Error(), err.Error(), tc.name)
		}
	}
}

func TestCreateDataSourceSchema_ForNestedObjects(t *testing.T) {
	t.Run("happy path -- a data soruce can be derived from a nested object keeping all the properies attributes as expected", func(t *testing.T) {
		// set up the schema for the nested object
		nestedObjectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		nestedObjectDefault := map[string]interface{}{
			"origin_port": nestedObjectSchemaDefinition.Properties[0].Default,
			"protocol":    nestedObjectSchemaDefinition.Properties[1].Default,
		}
		nestedObject := newObjectSchemaDefinitionPropertyWithDefaults("nested_object", "", true, false, false, nestedObjectDefault, nestedObjectSchemaDefinition)
		propertyWithNestedObjectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				idProperty,
				nestedObject,
			},
		}
		dataValue := map[string]interface{}{
			"id":            propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
			"nested_object": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
		}
		dataSourceSchema := newObjectSchemaDefinitionPropertyWithDefaults("nested-oobj", "", true, false, false, dataValue, propertyWithNestedObjectSchemaDefinition)
		specSchemaNested := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{dataSourceSchema},
		}

		// get the Terraform schema which represents a Data Source
		s, err := specSchemaNested.createDataSourceSchema()

		assert.NotNil(t, s)
		assert.NoError(t, err)

		// assertions on the properties attributes
		assert.Equal(t, false, s["nested_oobj"].Required)
		assert.Equal(t, true, s["nested_oobj"].Optional)
		assert.Equal(t, true, s["nested_oobj"].Computed)

		// 1^ level
		objectResource := s["nested_oobj"].Elem.(*schema.Resource)
		assert.Equal(t, 2, len(objectResource.Schema))

		assert.Equal(t, false, objectResource.Schema["id"].Required)
		assert.Equal(t, true, objectResource.Schema["id"].Optional)
		assert.Equal(t, true, objectResource.Schema["id"].Computed)
		assert.Equal(t, false, objectResource.Schema["nested_object"].Required)
		assert.Equal(t, true, objectResource.Schema["nested_object"].Optional)
		assert.Equal(t, true, objectResource.Schema["nested_object"].Computed)

		// 2^ level
		nestedObjectResource := objectResource.Schema["nested_object"].Elem.(*schema.Resource)
		assert.Equal(t, 2, len(nestedObjectResource.Schema))

		assert.Equal(t, false, nestedObjectResource.Schema["origin_port"].Required)
		assert.Equal(t, true, nestedObjectResource.Schema["origin_port"].Optional)
		assert.Equal(t, true, nestedObjectResource.Schema["origin_port"].Computed)
		assert.Equal(t, false, nestedObjectResource.Schema["protocol"].Required)
		assert.Equal(t, true, nestedObjectResource.Schema["protocol"].Optional)
		assert.Equal(t, true, nestedObjectResource.Schema["protocol"].Computed)
	})
}

func TestCreateResourceSchema(t *testing.T) {
	Convey("Given a swagger schema definition that has few properties including the id all with terraform compliant names", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
				&specSchemaDefinitionProperty{
					Name:     "string_prop",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the  err returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the resulted tfResourceSchema should not contain ID property", func() {
				So(tfResourceSchema, ShouldNotContainKey, "id")
			})
			Convey("Then the resulted tfResourceSchema should contain string_prop property", func() {
				So(tfResourceSchema, ShouldContainKey, "string_prop")
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including the id all with NON terraform compliant names", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "ID",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
				&specSchemaDefinitionProperty{
					Name:     "stringProp",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the  err returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the resulted tfResourceSchema should not contain ID property", func() {
				So(tfResourceSchema, ShouldNotContainKey, "id")
			})
			Convey("Then the resulted tfResourceSchema should contain a key with the terraform compliant name string_prop ", func() {
				So(tfResourceSchema, ShouldContainKey, "string_prop")
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including an object property with internal ID field", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
				&specSchemaDefinitionProperty{
					Name:     "string_prop",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeObject,
					ReadOnly: true,
					SpecSchemaDefinition: &specSchemaDefinition{
						Properties: specSchemaDefinitionProperties{
							&specSchemaDefinitionProperty{
								Name:               "id",
								Type:               typeString,
								ReadOnly:           true,
								IsStatusIdentifier: true,
							},
							&specSchemaDefinitionProperty{
								Name:               "name",
								Type:               typeString,
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
			Convey("Then the  err returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resulted tfResourceSchema should not contain ID property", func() {
				So(tfResourceSchema, ShouldNotContainKey, "id")
			})
			Convey("And the resulted tfResourceSchema should contain string_prop property of type string", func() {
				So(tfResourceSchema, ShouldContainKey, "string_prop")
				So(tfResourceSchema["string_prop"].Type, ShouldEqual, schema.TypeString)
			})
			Convey("And the resulted tfResourceSchema should contain status property", func() {
				So(tfResourceSchema, ShouldContainKey, statusDefaultPropertyName)
			})
			Convey("And the resulted tfResourceSchema status field be of type map", func() {
				So(tfResourceSchema[statusDefaultPropertyName].Type, ShouldEqual, schema.TypeMap)
			})
			Convey("And the resulted tfResourceSchema status field should contain all the sub-properties including the id property", func() {
				So(tfResourceSchema[statusDefaultPropertyName].Elem.(*schema.Resource).Schema, ShouldContainKey, "id")
				So(tfResourceSchema[statusDefaultPropertyName].Elem.(*schema.Resource).Schema, ShouldContainKey, "name")
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including an array of string primitives", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
				&specSchemaDefinitionProperty{
					Name:           "listeners",
					Type:           typeList,
					ArrayItemsType: typeString,
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the  err returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resulted tfResourceSchema should not contain ID property", func() {
				So(tfResourceSchema, ShouldNotContainKey, "id")
			})
			Convey("And the resulted tfResourceSchema should contain the array property", func() {
				So(tfResourceSchema, ShouldContainKey, "listeners")
			})
			Convey("And the resulted tfResourceSchema listeners field should be of type list and contain the right item schema string type", func() {
				So(tfResourceSchema["listeners"].Type, ShouldEqual, schema.TypeList)
				So(tfResourceSchema["listeners"].Elem.(*schema.Schema).Type, ShouldEqual, schema.TypeString)
			})
		})
	})

	Convey("Given a swagger schema definition that has few properties including an array of objects", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: false,
					Required: true,
				},
				&specSchemaDefinitionProperty{
					Name:           "listeners",
					Type:           typeList,
					ArrayItemsType: typeObject,
					SpecSchemaDefinition: &specSchemaDefinition{
						Properties: specSchemaDefinitionProperties{
							&specSchemaDefinitionProperty{
								Name: "protocol",
								Type: typeString,
							},
						},
					},
				},
			},
		}
		Convey("When createResourceSchema method is called", func() {
			tfResourceSchema, err := s.createResourceSchema()
			Convey("Then the  err returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resulted tfResourceSchema should not contain ID property", func() {
				So(tfResourceSchema, ShouldNotContainKey, "id")
			})
			Convey("And the resulted tfResourceSchema should contain the array property", func() {
				So(tfResourceSchema, ShouldContainKey, "listeners")
			})
			Convey("And the resulted tfResourceSchema should contain a field of type list and the list should have the right object elem schema of type Resource", func() {
				So(tfResourceSchema["listeners"].Type, ShouldEqual, schema.TypeList)
				So(tfResourceSchema["listeners"].Elem.(*schema.Resource).Schema, ShouldContainKey, "protocol")
				So(tfResourceSchema["listeners"].Elem.(*schema.Resource).Schema["protocol"].Type, ShouldEqual, schema.TypeString)
			})
		})
	})
}

func TestGetImmutableProperties(t *testing.T) {
	Convey("Given resource info is configured with schemaDefinition that contains a property 'immutable_property' that is immutable", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:      "id",
					Type:      typeString,
					ReadOnly:  false,
					Immutable: true,
				},
				&specSchemaDefinitionProperty{
					Name:      "immutable_property",
					Type:      typeString,
					ReadOnly:  false,
					Immutable: true,
				},
			},
		}
		Convey("When getImmutableProperties method is called", func() {
			immutableProperties := s.getImmutableProperties()
			Convey("Then the array returned should contain 'immutable_property'", func() {
				So(immutableProperties, ShouldContain, "immutable_property")
			})
			Convey("And the 'id' property should be ignored even if it's marked as immutable (should never happen though, edge case)", func() {
				So(immutableProperties, ShouldNotContain, "id")
			})
		})
	})

	Convey("Given resource info is configured with schemaDefinition that DOES NOT contain immutable properties", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: false,
				},
				&specSchemaDefinitionProperty{
					Name:      "mutable_property",
					Type:      typeString,
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
	Convey("Given a specSchemaDefinition containing a field named id", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			id, err := s.getResourceIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the id returned should be the same as the one in the specSchemaDefinition", func() {
				So(id, ShouldEqual, s.Properties[0].Name)
			})
		})
	})

	Convey("Given a specSchemaDefinition that does not contain a field named id but has a property tagged as IsIdentifier", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:         "someOtherID",
					Type:         typeString,
					ReadOnly:     true,
					IsIdentifier: true,
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			id, err := s.getResourceIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the id returned should be the same as the one in the specSchemaDefinition", func() {
				So(id, ShouldEqual, s.Properties[0].Name)
			})
		})
	})

	Convey("Given a specSchemaDefinition not containing a field named id nor tagged as identifier", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "someOtherField",
					Type:     typeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getResourceIdentifier method is called", func() {
			_, err := s.getResourceIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then error message should equal", func() {
				So(err.Error(), ShouldEqual, "could not find any identifier property in the resource schema definition")
			})
		})
	})
}

func TestGetStatusId(t *testing.T) {
	Convey("Given a SpecSchemaDefinition that has an status property that is not an object", t, func() {
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeString,
					ReadOnly: true,
				},
			},
		}

		Convey("When getStatusIdentifier method is called", func() {
			statuses, err := s.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the statuses returned should not be empty'", func() {
				So(statuses, ShouldNotBeEmpty)
			})
			Convey("Then the value returned should contain the name of the property 'statuses'", func() {
				So(statuses[0], ShouldEqual, statusDefaultPropertyName)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with IsStatusIdentifier set to TRUE", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the status returned should not be empty'", func() {
				So(status, ShouldNotBeEmpty)
			})
			Convey("Then the value returned should contain the name of the property 'some-other-property-holding-status'", func() {
				So(status[0], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that HAS BOTH an 'status' property AND ALSO a property configured with 'x-terraform-field-status' set to true", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeString,
					ReadOnly: true,
				},
				&specSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifier method is called", func() {
			status, err := s.getStatusIdentifier()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the status returned should not be empty'", func() {
				So(status, ShouldNotBeEmpty)
			})
			Convey("Then the value returned should be 'some-other-property-holding-status' as it takes preference over the default 'status' property", func() {
				So(status[0], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that HAS an status field which is an object containing a status field", t, func() {
		expectedStatusProperty := "actualStatus"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     "id",
					Type:     typeString,
					ReadOnly: true,
				},
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeObject,
					ReadOnly: true,
					SpecSchemaDefinition: &specSchemaDefinition{
						Properties: specSchemaDefinitionProperties{
							&specSchemaDefinitionProperty{
								Name:               expectedStatusProperty,
								Type:               typeString,
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the status returned should not be empty'", func() {
				So(status, ShouldNotBeEmpty)
			})
			Convey("Then the status array returned should contain the different the trace of property names (hierarchy) until the last one which is the one used as status, in this case 'actualStatus' on the last index", func() {
				So(status[0], ShouldEqual, "status")
				So(status[1], ShouldEqual, expectedStatusProperty)
			})
		})
	})

	Convey("Given a swagger schema definition that DOES NOT have an 'status' property but has a property configured with 'x-terraform-field-status' set to FALSE", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               expectedStatusProperty,
					Type:               typeString,
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
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               "prop-that-is-not-status",
					Type:               typeString,
					ReadOnly:           true,
					IsStatusIdentifier: false,
				},
				&specSchemaDefinitionProperty{
					Name:               "prop-that-is-not-status-and-does-not-have-status-metadata-either",
					Type:               typeString,
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
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     statusDefaultPropertyName,
					Type:     typeString,
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
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:               statusDefaultPropertyName,
					Type:               typeString,
					ReadOnly:           false,
					IsStatusIdentifier: true,
				},
			},
		}
		Convey("When getStatusIdentifierFor method is called with a schema definition and forceReadOnly check is disabled (this happens when the method is called recursively)", func() {
			status, err := s.getStatusIdentifierFor(s, true, false)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the status array returned should contain the status property even though it's not read only...readonly checks are only perform on root level properties", func() {
				So(status[0], ShouldEqual, "status")
			})
		})
	})
}

func TestGetProperty(t *testing.T) {
	Convey("Given a specSchemaDefinition", t, func() {
		existingPropertyName := "existingPropertyName"
		s := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				&specSchemaDefinitionProperty{
					Name:     existingPropertyName,
					Type:     typeString,
					ReadOnly: false,
				},
			},
		}
		Convey("When getProperty method is called with an existing property name", func() {
			property, err := s.getProperty(existingPropertyName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the property returned should be the same as the one in the specSchemaDefinition", func() {
				So(property, ShouldEqual, s.Properties[0])
			})
		})
		Convey("When getProperty method is called with a NON existing property name", func() {
			_, err := s.getProperty("nonExistingPropertyName")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the property returned should be the same as the one in the specSchemaDefinition", func() {
				So(err.Error(), ShouldEqual, "property with name 'nonExistingPropertyName' not existing in resource schema definition")
			})
		})
	})
}

func TestGetPropertyBasedOnTerraformName(t *testing.T) {
	existingPropertyName := "existingPropertyName"
	s := &specSchemaDefinition{
		Properties: specSchemaDefinitionProperties{
			&specSchemaDefinitionProperty{
				Name:     existingPropertyName,
				Type:     typeString,
				ReadOnly: false,
			},
		},
	}
	_, err := s.getPropertyBasedOnTerraformName("badTerraformPropertyName")
	assert.EqualError(t, err, "property with terraform name 'badTerraformPropertyName' not existing in resource schema definition")

}
