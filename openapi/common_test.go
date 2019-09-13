package openapi

import (
	"errors"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func TestCheckHTTPStatusCode(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		openAPIResource := &specStubResource{name: "resourceName"}
		Convey("When checkHTTPStatusCode is called with a response containing a status codes that matches one of the expected response status codes", func() {
			response := &http.Response{
				StatusCode: http.StatusOK,
			}
			expectedStatusCodes := []int{http.StatusOK}
			err := checkHTTPStatusCode(openAPIResource, response, expectedStatusCodes)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("When checkHTTPStatusCode is called with a response that IS NOT expected", func() {
			response := &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("some backend error")),
				StatusCode: http.StatusInternalServerError,
			}
			expectedStatusCodes := []int{http.StatusOK}
			err := checkHTTPStatusCode(openAPIResource, response, expectedStatusCodes)
			Convey("Then the err returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the err messages should equal", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] (some backend error)")
			})
		})
		Convey("When checkHTTPStatusCode is called with a response known with code 401 Unauthorized", func() {
			response := &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("unauthorized")),
				StatusCode: http.StatusUnauthorized,
			}
			expectedStatusCodes := []int{http.StatusOK}
			err := checkHTTPStatusCode(openAPIResource, response, expectedStatusCodes)
			Convey("Then the err returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the err messages should equal", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] HTTP Response Status Code 401 - Unauthorized: API access is denied due to invalid credentials (unauthorized)")
			})
		})
	})
}

func TestResponseContainsExpectedStatus(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		Convey("When responseContainsExpectedStatus is called with a response code that exists in the given list of expected status codes", func() {
			expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
			responseCode := http.StatusCreated
			exists := responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the expectedValue returned should be true", func() {
				So(exists, ShouldBeTrue)
			})
		})
		Convey("When responseContainsExpectedStatus is called with a response code that DOES NOT exists in 'expectedResponseStatusCodes'", func() {
			expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
			responseCode := http.StatusUnauthorized
			exists := responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the expectedValue returned should be false", func() {
				So(exists, ShouldBeFalse)
			})
		})
	})
}

func TestGetParentIDsAndResourcePath(t *testing.T) {
	Convey("Given an nil openapi resource (internal getParentIDs call fails for some reason)", t, func() {
		Convey("When getParentIDsAndResourcePath is called", func() {
			parentIDs, resourcePath, err := getParentIDsAndResourcePath(nil, nil)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "can't get parent ids from a resourceFactory with no openAPIResource")
			})
			Convey("And the parentIDs should be empty", func() {
				So(parentIDs, ShouldBeEmpty)
			})
			Convey("And the resourcePath should be empty", func() {
				So(resourcePath, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an empty openapi resource (internal getResourcePath() call fails for some reason)", t, func() {
		someFirewallProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")
		testSchema := newTestSchema(someFirewallProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		openAPIResource := &specStubResource{
			funcGetResourcePath: func(parentIDs []string) (s string, e error) {
				return "", errors.New("getResourcePath() failed")
			}}

		Convey("When getParentIDsAndResourcePath is called", func() {
			parentIDs, resourcePath, err := getParentIDsAndResourcePath(openAPIResource, resourceData)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "getResourcePath() failed")
			})
			Convey("And the parentIDs should be empty", func() {
				So(parentIDs, ShouldBeEmpty)
			})
			Convey("And the resourcePath should be empty", func() {
				So(resourcePath, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resource configured with a subreousrce", t, func() {
		someFirewallProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")

		// Pretending the data has already been populated with the parent property
		testSchema := newTestSchema(someFirewallProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		openAPIResource := &SpecV2Resource{
			Path: "/v1/cdns/{id}/firewall",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"some_string_prop"},
					Properties: map[string]spec.Schema{
						"some_string_prop": spec.Schema{
							SchemaProps: spec.SchemaProps{
								Required: []string{},
							},
						},
					},
				},
			},
		}

		Convey("When getParentIDsAndResourcePath is called", func() {
			parentIDs, resourcePath, err := getParentIDsAndResourcePath(openAPIResource, resourceData)
			Convey("Then the error returned should be the expected one", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the parentIDs should match the expected", func() {
				So(len(parentIDs), ShouldEqual, 1)
				So(parentIDs[0], ShouldEqual, "parentPropertyID")
			})
			Convey("And the resourcePath be '/v1/cdns/parentPropertyID/firewall'", func() {
				So(resourcePath, ShouldEqual, "/v1/cdns/parentPropertyID/firewall")
			})
		})
	})
}

func Test_getParentIDs(t *testing.T) {

	Convey("Given a resourceFactory with no openAPIResource", t, func() {
		rf := resourceFactory{}
		Convey("When getParentIDs is called", func() {
			ss, e := rf.getParentIDs(nil)
			Convey("Then an error is raised", func() {
				So(e.Error(), ShouldEqual, "can't get parent ids from a resourceFactory with no openAPIResource")
			})
			Convey("And the slice of string returned is empty", func() {
				So(ss, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resourceFactory with a pointer to a blank SpecV2Resource", t, func() {
		rf := resourceFactory{openAPIResource: &SpecV2Resource{}}
		Convey("When getParentIDs is called with a nil arg", func() {
			ss, err := rf.getParentIDs(nil)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the slice of string returned is empty", func() {
				So(ss, ShouldBeEmpty)
			})
		})
		Convey("When getParentIDs is called with an empty ResourceData", func() {
			ss, err := rf.getParentIDs(&schema.ResourceData{})
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the slice of string returned is empty", func() {
				So(ss, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resourceFactory with a some schema", t, func() {
		someFirewallProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")

		// Pretending the data has already been populated with the parent property
		testSchema := newTestSchema(someFirewallProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		rf := newResourceFactory(&SpecV2Resource{
			Path: "/v1/cdns/{id}/firewall",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"some_string_prop"},
					Properties: map[string]spec.Schema{
						"some_string_prop": spec.Schema{
							SchemaProps: spec.SchemaProps{
								Required: []string{},
							},
						},
					},
				},
			},
		})

		Convey("When getParentIDs is called with non-empty ResourceData", func() {
			parentIDs, err := rf.getParentIDs(resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the parent IDs returned should be populated as expected", func() {
				So(parentIDs[0], ShouldEqual, "parentPropertyID")
			})
		})
	})
}

func TestUpdateStateWithPayloadData(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		objectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		objectStateValue := map[string]interface{}{
			"origin_port": objectSchemaDefinition.Properties[0].Default,
			"protocol":    objectSchemaDefinition.Properties[1].Default,
		}
		objectProperty := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, false, objectStateValue, objectSchemaDefinition)
		arrayObjectStateValue := []map[string]interface{}{
			{
				"origin_port": 80,
				"protocol":    "http",
			},
		}
		listOfObjectsProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, arrayObjectStateValue, typeObject, objectSchemaDefinition)

		propertyWithNestedObjectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				idProperty,
				objectProperty,
			},
		}
		objectWithNestedObjectStateValue := map[string]interface{}{
			"id":            propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
			"nested_object": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
		}

		propertyWithNestedObject := newObjectSchemaDefinitionPropertyWithDefaults("property_with_nested_object", "", true, false, false, objectWithNestedObjectStateValue, propertyWithNestedObjectSchemaDefinition)
		r, resourceData := testCreateResourceFactory(t, stringWithPreferredNameProperty, intProperty, numberProperty, boolProperty, slicePrimitiveProperty, objectProperty, listOfObjectsProperty, propertyWithNestedObject)
		Convey("When  is called with a map containing all property types supported (string, int, number, bool, slice of primitives, objects, list of objects and property with nested objects)", func() {
			remoteData := map[string]interface{}{
				stringWithPreferredNameProperty.Name: "someUpdatedStringValue",
				intProperty.Name:                     15,
				numberProperty.Name:                  26.45,
				boolProperty.Name:                    true,
				slicePrimitiveProperty.Name:          []interface{}{"value1"},
				objectProperty.Name: map[string]interface{}{
					"origin_port": 80,
					"protocol":    "http",
				},
				listOfObjectsProperty.Name: []interface{}{
					map[string]interface{}{
						"origin_port": 80,
						"protocol":    "http",
					},
				},
				propertyWithNestedObject.Name: map[string]interface{}{
					idProperty.Name: propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
					objectProperty.Name: map[string]interface{}{
						"origin_port": 80,
						"protocol":    "http",
					},
				},
			}
			err := updateStateWithPayloadData(r.openAPIResource, remoteData, resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the expectedValue should equal to the expectedValue coming from remote, and also the key expectedValue should be the preferred as defined in the property", func() {
				// keys stores in the resource data struct are always snake case
				So(resourceData.Get(stringWithPreferredNameProperty.getTerraformCompliantPropertyName()), ShouldEqual, remoteData[stringWithPreferredNameProperty.Name])
				So(resourceData.Get(intProperty.getTerraformCompliantPropertyName()), ShouldEqual, remoteData[intProperty.Name])
				So(resourceData.Get(numberProperty.getTerraformCompliantPropertyName()), ShouldEqual, remoteData[numberProperty.Name])
				So(resourceData.Get(boolProperty.getTerraformCompliantPropertyName()), ShouldEqual, remoteData[boolProperty.Name])
				So(len(resourceData.Get(slicePrimitiveProperty.getTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 1)
				So(resourceData.Get(slicePrimitiveProperty.getTerraformCompliantPropertyName()).([]interface{})[0], ShouldEqual, remoteData[slicePrimitiveProperty.Name].([]interface{})[0])
				So(resourceData.Get(objectProperty.getTerraformCompliantPropertyName()).(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(objectProperty.getTerraformCompliantPropertyName()).(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(objectProperty.getTerraformCompliantPropertyName()).(map[string]interface{})["origin_port"], ShouldEqual, strconv.Itoa(remoteData[objectProperty.Name].(map[string]interface{})["origin_port"].(int)))
				So(resourceData.Get(objectProperty.getTerraformCompliantPropertyName()).(map[string]interface{})["protocol"], ShouldEqual, remoteData[objectProperty.Name].(map[string]interface{})["protocol"])

				So(len(resourceData.Get(listOfObjectsProperty.getTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 1)
				So(resourceData.Get(listOfObjectsProperty.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(listOfObjectsProperty.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(listOfObjectsProperty.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["origin_port"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[0].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(listOfObjectsProperty.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["protocol"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[0].(map[string]interface{})["protocol"])

				So(len(resourceData.Get(propertyWithNestedObject.getTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 1)
				So(resourceData.Get(propertyWithNestedObject.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, idProperty.Name)
				So(resourceData.Get(propertyWithNestedObject.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, objectProperty.Name)
				So(resourceData.Get(propertyWithNestedObject.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(propertyWithNestedObject.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(propertyWithNestedObject.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].(map[string]interface{})["origin_port"], ShouldEqual, strconv.Itoa(remoteData[propertyWithNestedObject.Name].(map[string]interface{})[objectProperty.Name].(map[string]interface{})["origin_port"].(int)))
				So(resourceData.Get(propertyWithNestedObject.getTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].(map[string]interface{})["protocol"], ShouldEqual, remoteData[propertyWithNestedObject.Name].(map[string]interface{})[objectProperty.Name].(map[string]interface{})["protocol"])
			})
		})
	})

	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactory(t, stringWithPreferredNameProperty)
		Convey("When is called with a map remoteData containing more properties than then ones specified in the schema (this means the API is returning more info than the one specified in the swagger file)", func() {
			remoteData := map[string]interface{}{
				stringWithPreferredNameProperty.Name:                "someUpdatedStringValue",
				"some_other_property_not_documented_in_openapi_doc": 15,
			}
			err := updateStateWithPayloadData(r.openAPIResource, remoteData, resourceData)
			Convey("Then the err returned should matched the expected one", func() {
				So(err.Error(), ShouldEqual, "failed to update state with remote data. This usually happens when the API returns properties that are not specified in the resource's schema definition in the OpenAPI document - error = property with name 'some_other_property_not_documented_in_openapi_doc' not existing in resource schema definition")
			})
		})
	})
}

func TestConvertPayloadToLocalStateDataValue(t *testing.T) {

	Convey("Given a resource factory", t, func() {

		Convey("When convertPayloadToLocalStateDataValue is called with ", func() {
			property := newStringSchemaDefinitionPropertyWithDefaults("string_property", "", false, false, nil)
			dataValue := "someValue"
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value with the right type string", func() {
				So(resultValue, ShouldEqual, dataValue)
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with a bool property and a bool value", func() {
			property := newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", false, false, nil)
			dataValue := true
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value with the right type boolean", func() {
				So(resultValue, ShouldEqual, dataValue)
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with a bool property, a bool value true and the desired output is string", func() {
			property := newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", false, false, nil)
			dataValue := true
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, true)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value with the right type boolean", func() {
				So(resultValue, ShouldEqual, "true")
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with a int property, a bool value false and the desired output is string", func() {
			property := newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", false, false, nil)
			dataValue := false
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, true)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value formatted string with the right type boolean", func() {
				So(resultValue, ShouldEqual, "false")
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with an int property and a int value", func() {
			property := newIntSchemaDefinitionPropertyWithDefaults("int_property", "", false, false, nil)
			dataValue := 10
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value with the right type int", func() {
				So(resultValue, ShouldEqual, dataValue)
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an int property and a int value and the desired output is string", func() {
			property := newIntSchemaDefinitionPropertyWithDefaults("int_property", "", false, false, nil)
			dataValue := 10
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, true)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value formatted string with the right type int", func() {
				So(resultValue, ShouldEqual, "10")
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with an rune property and a rune value and the desired output is nil", func() {
			property := newIntSchemaDefinitionPropertyWithDefaults("int_property", "", false, false, nil)
			dataValue := 'f'
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, true)
			Convey("Then the error should not be nil", func() {
				So(err.Error(), ShouldEqual, "'int32' type not supported")
			})
			Convey("Then the result value should be the expected value formatted string with the right type int", func() {
				So(resultValue, ShouldEqual, nil)
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with an float property and a float value", func() {
			property := newNumberSchemaDefinitionPropertyWithDefaults("float_property", "", false, false, nil)
			dataValue := 45.23
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value formatted string with the right type float", func() {
				So(resultValue, ShouldEqual, dataValue)
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an float property and a float value Zero and the desired output is string", func() {
			property := newNumberSchemaDefinitionPropertyWithDefaults("float_property", "", false, false, nil)
			dataValue := 0
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, true)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value 0 formatted string with the right type float", func() {
				So(resultValue, ShouldEqual, "0")
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an float property and a float value and the desired output is string", func() {
			property := newNumberSchemaDefinitionPropertyWithDefaults("float_property", "", false, false, nil)
			dataValue := 10.12
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, true)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value formatted string with the right type float", func() {
				So(resultValue, ShouldEqual, "10.12")
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an float property and a float value but the swagger property is an integer", func() {
			property := newIntSchemaDefinitionPropertyWithDefaults("int_property", "", false, false, nil)
			dataValue := 45
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value formatted string with the right type integer", func() {
				So(resultValue, ShouldEqual, dataValue)
				So(resultValue, ShouldHaveSameTypeAs, int(dataValue))
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an float property and a float value but the swagger property is an integer and the expected output format is string", func() {
			property := newIntSchemaDefinitionPropertyWithDefaults("int_property", "", false, false, nil)
			dataValue := 45
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, true)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value formatted string with the right type integer", func() {
				So(resultValue, ShouldEqual, "45")
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with an list property and a with items object", func() {
			objectSchemaDefinition := &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newIntSchemaDefinitionPropertyWithDefaults("example_int", "", true, false, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_string", "", true, false, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_bool", "", true, false, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_float", "", true, false, nil),
				},
			}
			objectDefault := map[string]interface{}{
				"example_int":    80,
				"example_string": "http",
				"example_bool":   true,
				"example_float":  10.45,
			}
			property := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, nil, typeObject, objectSchemaDefinition)
			dataValue := []interface{}{objectDefault}
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the list containing the object items with the expected types (int, string, bool and float)", func() {
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_int")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_int"].(int), ShouldEqual, objectDefault["example_int"])
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_string")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_string"].(string), ShouldEqual, objectDefault["example_string"])
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_bool")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_bool"].(bool), ShouldEqual, objectDefault["example_bool"])
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_float")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_float"].(float64), ShouldEqual, objectDefault["example_float"])
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with a list property and an array with items string value", func() {
			property := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, nil, typeString, nil)
			dataValue := []interface{}{"value1"}
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value with the right type array", func() {
				So(resultValue.([]interface{}), ShouldContain, dataValue[0])
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with simple object property and an empty map as value", func() {
			property := &specSchemaDefinitionProperty{
				Name:     "some_object",
				Type:     typeObject,
				Required: true,
			}
			resultValue, err := convertPayloadToLocalStateDataValue(property, map[string]interface{}{}, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the expected value with the right type array", func() {
				So(resultValue.(map[string]interface{}), ShouldBeEmpty)
			})
		})

		// Edge case
		Convey("When convertPayloadToLocalStateDataValue is called with a slice of map interfaces", func() {
			property := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, nil, typeString, nil)
			_, err := convertPayloadToLocalStateDataValue(property, []map[string]interface{}{}, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with a property list that the array items are of unknown type", func() {
			property := &specSchemaDefinitionProperty{
				Name:           "not_well_configured_property",
				Type:           typeList,
				ArrayItemsType: schemaDefinitionPropertyType("unknown"),
			}
			_, err := convertPayloadToLocalStateDataValue(property, []interface{}{}, false)
			Convey("Then the error should match the expected one", func() {
				So(err.Error(), ShouldEqual, "property 'not_well_configured_property' is supposed to be an array objects")
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with a simple object", func() {
			// Simple objects are considered objects that all the properties are of the same type and are not computed
			objectSchemaDefinition := &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("example_string", "", true, false, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_string_2", "", true, false, nil),
				},
			}
			dataValue := map[string]interface{}{
				"example_string":   "http",
				"example_string_2": "something",
			}
			property := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, false, nil, objectSchemaDefinition)
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the list containing the object items all being string type (as terraform only supports maps of strings, hence values need to be stored as strings)", func() {
				So(resultValue.(map[string]interface{})["example_string"].(string), ShouldEqual, "http")
				So(resultValue.(map[string]interface{})["example_string_2"].(string), ShouldEqual, "something")
			})
		})

		// Simple objects are considered objects that contain properties that are of different types and configuration (e,g: mix of required/optional/computed properties)
		Convey("When convertPayloadToLocalStateDataValue is called with a complex object", func() {
			objectSchemaDefinition := &specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					newIntSchemaDefinitionPropertyWithDefaults("example_int", "", true, false, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_string", "", true, false, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_bool", "", true, true, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_float", "", true, false, nil),
				},
			}
			dataValue := map[string]interface{}{
				"example_int":    80,
				"example_string": "http",
				"example_bool":   true,
				"example_float":  10.45,
			}
			property := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, false, nil, objectSchemaDefinition)
			property.EnableLegacyComplexObjectBlockConfiguration = true
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue, false)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result value should be the list containing the object items all being string type (as terraform only supports maps of strings, hence values need to be stored as strings)", func() {
				So(resultValue.([]interface{})[0], ShouldContainKey, "example_int")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_int"].(string), ShouldEqual, "80")
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_string")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_string"].(string), ShouldEqual, "http")
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_bool")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_bool"].(string), ShouldEqual, "true")
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_float")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_float"].(string), ShouldEqual, "10.45")
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with an object containing objects", func() {
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
			// The below represents the JSON representation of the response payload received by the API
			dataValue := map[string]interface{}{
				"id":            propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
				"nested_object": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
			}

			expectedPropertyWithNestedObjectName := "property_with_nested_object"
			propertyWithNestedObject := newObjectSchemaDefinitionPropertyWithDefaults(expectedPropertyWithNestedObjectName, "", true, false, false, dataValue, propertyWithNestedObjectSchemaDefinition)
			resultValue, err := convertPayloadToLocalStateDataValue(propertyWithNestedObject, dataValue, false)

			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the result value should be the list containing just one element (as per the nested struct workaround)", func() {
				// Tag(NestedStructsWorkaround)
				// Note: This is the workaround needed to support properties with nested structs. The current Terraform sdk version
				// does not support this now, hence the suggestion from the Terraform maintainer was to use a list of map[string]interface{}
				// with the list containing just one element. The below represents the internal representation of the terraform state
				// for an object property that contains other objects
				So(resultValue.([]interface{}), ShouldNotBeEmpty)
				So(len(resultValue.([]interface{})), ShouldEqual, 1)
			})
			Convey("AND the object should have the expected properties including the nested object", func() {
				So(resultValue.([]interface{})[0], ShouldContainKey, propertyWithNestedObjectSchemaDefinition.Properties[0].Name)
				So(resultValue.([]interface{})[0], ShouldContainKey, propertyWithNestedObjectSchemaDefinition.Properties[1].Name)
			})
			Convey("AND the object property with nested object should have the expected configuration", func() {
				nestedObject := propertyWithNestedObjectSchemaDefinition.Properties[1]
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name], ShouldContainKey, nestedObjectSchemaDefinition.Properties[0].Name)
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name].(map[string]interface{})[nestedObjectSchemaDefinition.Properties[0].Name], ShouldEqual, strconv.Itoa(nestedObjectSchemaDefinition.Properties[0].Default.(int)))
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name], ShouldContainKey, nestedObjectSchemaDefinition.Properties[1].Name)
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name].(map[string]interface{})[nestedObjectSchemaDefinition.Properties[1].Name], ShouldEqual, nestedObjectSchemaDefinition.Properties[1].Default)
			})
		})
	})
}

func TestSetResourceDataProperty(t *testing.T) {
	Convey("Given a resource factory initialized with a spec resource with some schema definition", t, func() {
		r, resourceData := testCreateResourceFactory(t, stringProperty, stringWithPreferredNameProperty)
		Convey("When setResourceDataProperty is called with a schema definition property name that exists in terraform resource data object and with a new expectedValue", func() {
			expectedValue := "newValue"
			err := setResourceDataProperty(r.openAPIResource, stringProperty.Name, expectedValue, resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then expectedValue should equal", func() {
				// keys stores in the resource data struct are always snake case
				So(resourceData.Get(stringProperty.Name), ShouldEqual, expectedValue)
			})
		})
		Convey("When setResourceDataProperty is called with a schema definition property preferred name that exists in terraform resource data object and with a new expectedValue", func() {
			expectedValue := "theNewValue"
			err := setResourceDataProperty(r.openAPIResource, stringWithPreferredNameProperty.Name, expectedValue, resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then expectedValue should equal the expected value (note the state is queried using the preferred name)", func() {
				// keys stores in the resource data struct are always snake case
				So(resourceData.Get(stringWithPreferredNameProperty.PreferredName), ShouldEqual, expectedValue)
			})
		})
		Convey("When setResourceDataProperty is called with a schema definition property name does NOT exist", func() {
			err := setResourceDataProperty(r.openAPIResource, "nonExistingKey", "", resourceData)
			Convey("Then the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And then expectedValue should equal", func() {
				// keys stores in the resource data struct are always snake case
				So(err.Error(), ShouldEqual, "could not find schema definition property name nonExistingKey in the resource data: property with name 'nonExistingKey' not existing in resource schema definition")
			})
		})
	})
}
