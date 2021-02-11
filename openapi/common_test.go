package openapi

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestCRUDWithContext(t *testing.T) {
	Convey("Given a create function (which returns successfully), a create timeout and a resource name", t, func() {
		stubCreateFunction := func(data *schema.ResourceData, i interface{}) error {
			return nil // this means the function returned successfully
		}
		createTimeout := 1 * time.Second
		resourceName := "cdn_v1"
		Convey("When crudWithContext is called", func() {
			contextAwareFunc := crudWithContext(stubCreateFunction, schema.TimeoutCreate, resourceName)
			Convey("Then the returned function which is context aware should not timeout and return an empty diagnosis", func() {
				ctx := context.Background()
				ctx, cancel := context.WithTimeout(ctx, createTimeout)
				defer cancel()
				diagnosis := contextAwareFunc(ctx, &schema.ResourceData{}, nil)
				So(diagnosis, ShouldBeEmpty)
			})
		})
	})
	Convey("Given a create function (which returns an error), a create timeout and a resource name", t, func() {
		expectedError := "some error"
		stubCreateFunction := func(data *schema.ResourceData, i interface{}) error {
			return errors.New(expectedError)
		}
		createTimeout := 1 * time.Second
		resourceName := "cdn_v1"
		Convey("When crudWithContext is called", func() {
			contextAwareFunc := crudWithContext(stubCreateFunction, schema.TimeoutCreate, resourceName)
			Convey("Then the returned function which is context aware should not timeout and return the error from the create function", func() {
				ctx := context.Background()
				ctx, cancel := context.WithTimeout(ctx, createTimeout)
				defer cancel()
				diagnosis := contextAwareFunc(ctx, &schema.ResourceData{}, nil)
				So(diagnosis, ShouldNotBeEmpty)
				So(diagnosis[0].Summary, ShouldEqual, expectedError)
			})
		})
	})
	Convey("Given a create function (configured to timeout on purpose), a create timeout and a resource name", t, func() {
		stubCreateFunction := func(data *schema.ResourceData, i interface{}) error {
			time.Sleep(2 * time.Second)
			return nil
		}
		createTimeout := 1 * time.Second
		resourceName := "cdn_v1"
		Convey("When crudWithContext is called", func() {
			contextAwareFunc := crudWithContext(stubCreateFunction, schema.TimeoutCreate, resourceName)
			Convey("Then the returned function which is context aware should time out since the create operation takes longer than the context timeout", func() {
				ctx := context.Background()
				ctx, cancel := context.WithTimeout(ctx, createTimeout)
				defer cancel()
				diagnosis := contextAwareFunc(ctx, &schema.ResourceData{}, nil)
				So(diagnosis, ShouldNotBeEmpty)
				So(diagnosis[0].Summary, ShouldEqual, "context deadline exceeded: 'cdn_v1' create timeout is 20m0s") // the 20m0s is the default timeout if the openAPIResource is not configured with specific timeouts
			})
		})
	})
}

func TestCheckHTTPStatusCode(t *testing.T) {
	testCases := []struct {
		name             string
		inputResponse    *http.Response
		inputStatusCodes []int
		expectedError    error
	}{
		{
			name: "response containing a status codes that matches one of the expected response status codes",
			inputResponse: &http.Response{
				StatusCode: http.StatusOK,
			},
			inputStatusCodes: []int{http.StatusOK},
			expectedError:    nil,
		},
		{
			name: "response that IS NOT expected",
			inputResponse: &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("some backend error")),
				StatusCode: http.StatusInternalServerError,
			},
			inputStatusCodes: []int{http.StatusOK},
			expectedError:    errors.New("[resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] (some backend error)"),
		},
		{
			name: "response known with code 401 Unauthorized",
			inputResponse: &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("unauthorized")),
				StatusCode: http.StatusUnauthorized,
			},
			inputStatusCodes: []int{http.StatusOK},
			expectedError:    errors.New("[resource='resourceName'] HTTP Response Status Code 401 - Unauthorized: API access is denied due to invalid credentials (unauthorized)"),
		},
	}
	Convey("Given a specStubResource", t, func() {
		openAPIResource := &specStubResource{name: "resourceName"}
		for _, tc := range testCases {
			Convey(fmt.Sprintf("When checkHTTPStatusCode is called: %s", tc.name), func() {
				err := checkHTTPStatusCode(openAPIResource, tc.inputResponse, tc.inputStatusCodes)
				Convey("Then the error returned should be the expected one", func() {
					So(err, ShouldResemble, tc.expectedError)
				})
			})
		}
	})
}

func TestResponseContainsExpectedStatus(t *testing.T) {
	testCases := []struct {
		name                     string
		inputResponseStatusCodes []int
		inputResponseCode        int
		expectedResult           bool
	}{
		{
			name:                     "response code that exists in the given list of input status codes",
			inputResponseStatusCodes: []int{http.StatusCreated, http.StatusAccepted},
			inputResponseCode:        http.StatusCreated,
			expectedResult:           true,
		},
		{
			name:                     "response code that DOES NOT exists in the given list of input status codes",
			inputResponseStatusCodes: []int{http.StatusCreated, http.StatusAccepted},
			inputResponseCode:        http.StatusUnauthorized,
			expectedResult:           false,
		},
	}
	for _, tc := range testCases {
		Convey(fmt.Sprintf("When responseContainsExpectedStatus is called: %s", tc.name), t, func() {
			exists := responseContainsExpectedStatus(tc.inputResponseStatusCodes, tc.inputResponseCode)
			Convey("Then the result returned should be the expected one", func() {
				So(exists, ShouldEqual, tc.expectedResult)
			})
		})
	}
}

func TestGetParentIDsAndResourcePath(t *testing.T) {
	Convey("Given an nil openapi resource (internal getParentIDs call fails for some reason)", t, func() {
		Convey("When getParentIDsAndResourcePath is called", func() {
			parentIDs, resourcePath, err := getParentIDsAndResourcePath(nil, nil)
			Convey("Then the result returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "can't get parent ids from an empty SpecResource")
				So(parentIDs, ShouldBeEmpty)
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
			Convey("Then the result returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "getResourcePath() failed")
				So(parentIDs, ShouldBeEmpty)
				So(resourcePath, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resource configured with a subresource", t, func() {
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
						"some_string_prop": {
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
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(len(parentIDs), ShouldEqual, 1)
				So(parentIDs[0], ShouldEqual, "parentPropertyID")
				So(resourcePath, ShouldEqual, "/v1/cdns/parentPropertyID/firewall")
			})
		})
	})
}

func Test_getParentIDs(t *testing.T) {
	Convey("Given a nil openAPIResource", t, func() {
		Convey("When getParentIDs is called with a nil SpecResource", func() {
			ss, e := getParentIDs(nil, nil)
			Convey("Then the result returned should be the expected one", func() {
				So(e.Error(), ShouldEqual, "can't get parent ids from an empty SpecResource")
				So(ss, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a SpecResource (with no parent info)", t, func() {
		s := &SpecV2Resource{}
		Convey("When getParentIDs is called with an empty ResourceData", func() {
			ss, err := getParentIDs(s, &schema.ResourceData{})
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(ss, ShouldBeEmpty)
			})
		})
		Convey("When getParentIDs is called with a nil ResourceData", func() {
			ss, err := getParentIDs(s, nil)
			Convey("Then the result returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "can't get parent ids from a nil ResourceData")
				So(ss, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a spec resource with parent info", t, func() {
		s := &specStubResource{
			name:                   "firewall",
			path:                   "/v1/cdns/{id}/firewall",
			schemaDefinition:       &SpecSchemaDefinition{},
			parentResourceNames:    []string{"cdns_v1"},
			fullParentResourceName: "cdns_v1",
		}
		Convey("When getParentIDs is called with an empty ResourceData", func() {
			ss, err := getParentIDs(s, &schema.ResourceData{})
			Convey("Then the result returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "could not find ID value in the state file for subresource parent property 'cdns_v1_id'")
				So(ss, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a spec resource with a some schema including the parent properties", t, func() {
		someFirewallProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")

		// Pretending the data has already been populated with the parent property
		testSchema := newTestSchema(someFirewallProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		s := &SpecV2Resource{
			Path: "/v1/cdns/{id}/firewall",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"some_string_prop"},
					Properties: map[string]spec.Schema{
						"some_string_prop": {
							SchemaProps: spec.SchemaProps{
								Required: []string{},
							},
						},
					},
				},
			},
		}
		Convey("When getParentIDs is called with non-empty ResourceData", func() {
			parentIDs, err := getParentIDs(s, resourceData)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(parentIDs[0], ShouldEqual, "parentPropertyID")
			})
		})
	})
}

func TestUpdateStateWithPayloadData(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		objectSchemaDefinition := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		objectStateValue := map[string]interface{}{
			"origin_port": objectSchemaDefinition.Properties[0].Default,
			"protocol":    objectSchemaDefinition.Properties[1].Default,
		}
		objectProperty := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, false, objectStateValue, objectSchemaDefinition)
		arrayObjectStateValue := []interface{}{
			map[string]interface{}{
				"origin_port": 80,
				"protocol":    "http",
			},
		}
		listOfObjectsProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, arrayObjectStateValue, TypeObject, objectSchemaDefinition)

		propertyWithNestedObjectSchemaDefinition := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
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
		Convey("When updateStateWithPayloadData is called with a map containing all property types supported (string, int, number, bool, slice of primitives, objects, list of objects and property with nested objects)", func() {
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
			Convey("Then the expectedValue should equal to the expectedValue coming from remote, the key expectedValue should be the preferred as defined in the property, the error should be nil", func() {
				So(err, ShouldBeNil)
				// keys stores in the resource data struct are always snake case
				So(resourceData.Get(stringWithPreferredNameProperty.GetTerraformCompliantPropertyName()), ShouldEqual, remoteData[stringWithPreferredNameProperty.Name])
				So(resourceData.Get(intProperty.GetTerraformCompliantPropertyName()), ShouldEqual, remoteData[intProperty.Name])
				So(resourceData.Get(numberProperty.GetTerraformCompliantPropertyName()), ShouldEqual, remoteData[numberProperty.Name])
				So(resourceData.Get(boolProperty.GetTerraformCompliantPropertyName()), ShouldEqual, remoteData[boolProperty.Name])
				So(len(resourceData.Get(slicePrimitiveProperty.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 1)
				So(resourceData.Get(slicePrimitiveProperty.GetTerraformCompliantPropertyName()).([]interface{})[0], ShouldEqual, remoteData[slicePrimitiveProperty.Name].([]interface{})[0])
				So(len(resourceData.Get(objectProperty.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 1)
				So(resourceData.Get(objectProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(objectProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(objectProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["origin_port"], ShouldEqual, remoteData[objectProperty.Name].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(objectProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["protocol"], ShouldEqual, remoteData[objectProperty.Name].(map[string]interface{})["protocol"])

				So(len(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 1)
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["origin_port"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[0].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["protocol"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[0].(map[string]interface{})["protocol"])

				So(len(resourceData.Get(propertyWithNestedObject.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 1)
				So(resourceData.Get(propertyWithNestedObject.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, idProperty.Name)
				So(resourceData.Get(propertyWithNestedObject.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, objectProperty.Name)
				So(resourceData.Get(propertyWithNestedObject.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].([]interface{})[0].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(propertyWithNestedObject.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].([]interface{})[0].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(propertyWithNestedObject.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].([]interface{})[0].(map[string]interface{})["origin_port"], ShouldEqual, remoteData[propertyWithNestedObject.Name].(map[string]interface{})[objectProperty.Name].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(propertyWithNestedObject.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})[objectProperty.Name].([]interface{})[0].(map[string]interface{})["protocol"], ShouldEqual, remoteData[propertyWithNestedObject.Name].(map[string]interface{})[objectProperty.Name].(map[string]interface{})["protocol"])
			})
		})
	})

	Convey("Given a resource factory containing a schema with property lists that have the IgnoreItemsOrder set to true", t, func() {
		objectSchemaDefinition := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		arrayObjectStateValue := []interface{}{
			map[string]interface{}{
				"origin_port": 80,
				"protocol":    "http",
			},
			map[string]interface{}{
				"origin_port": 443,
				"protocol":    "https",
			},
		}
		listOfObjectsProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, arrayObjectStateValue, TypeObject, objectSchemaDefinition)
		listOfObjectsProperty.IgnoreItemsOrder = true

		listOfStrings := newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, false, []interface{}{"value1", "value2"}, TypeString, nil)
		listOfStrings.IgnoreItemsOrder = true

		r, resourceData := testCreateResourceFactory(t, listOfStrings, listOfObjectsProperty)
		Convey("When updateStateWithPayloadData is called", func() {
			remoteData := map[string]interface{}{
				listOfStrings.Name: []interface{}{"value2", "value1"},
				listOfObjectsProperty.Name: []interface{}{
					map[string]interface{}{
						"origin_port": 443,
						"protocol":    "https",
					},
					map[string]interface{}{
						"origin_port": 80,
						"protocol":    "http",
					},
				},
			}
			err := updateStateWithPayloadData(r.openAPIResource, remoteData, resourceData)
			Convey("Then the expectedValue should maintain the order of the local input (not the order of the remote lists) and error should be nil", func() {
				So(err, ShouldBeNil)
				// keys stores in the resource data struct are always snake case
				So(len(resourceData.Get(listOfStrings.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 2)
				So(resourceData.Get(listOfStrings.GetTerraformCompliantPropertyName()).([]interface{})[0], ShouldEqual, listOfStrings.Default.([]interface{})[0])
				So(resourceData.Get(listOfStrings.GetTerraformCompliantPropertyName()).([]interface{})[1], ShouldEqual, listOfStrings.Default.([]interface{})[1])

				So(len(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 2)
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["origin_port"], ShouldEqual, arrayObjectStateValue[0].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["protocol"], ShouldEqual, arrayObjectStateValue[0].(map[string]interface{})["protocol"])

				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{})["origin_port"], ShouldEqual, arrayObjectStateValue[1].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{})["protocol"], ShouldEqual, arrayObjectStateValue[1].(map[string]interface{})["protocol"])
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
			Convey("Then the resource state data only contains the properties and values for the documented properties and error should be nil", func() {
				So(err, ShouldBeNil)
				So(resourceData.Get(stringWithPreferredNameProperty.GetTerraformCompliantPropertyName()), ShouldEqual, remoteData[stringWithPreferredNameProperty.Name])
				So(resourceData.Get("some_other_property_not_documented_in_openapi_doc"), ShouldBeNil)
			})
		})
	})
}

func TestDataSourceUpdateStateWithPayloadData(t *testing.T) {
	Convey("Given a resource factory containing a schema with property lists that have the IgnoreItemsOrder set to true", t, func() {
		objectSchemaDefinition := &SpecSchemaDefinition{
			Properties: SpecSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		arrayObjectStateValue := []interface{}{}
		listOfObjectsProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, arrayObjectStateValue, TypeObject, objectSchemaDefinition)
		listOfObjectsProperty.IgnoreItemsOrder = true

		listOfStrings := newListSchemaDefinitionPropertyWithDefaults("slice_property", "", true, false, false, []interface{}{"value1", "value2"}, TypeString, nil)
		listOfStrings.IgnoreItemsOrder = true

		r, resourceData := testCreateResourceFactory(t, listOfStrings, listOfObjectsProperty)
		Convey("When dataSourceUpdateStateWithPayloadData is called", func() {
			remoteData := map[string]interface{}{
				listOfStrings.Name: []interface{}{"value2", "value1"},
				listOfObjectsProperty.Name: []interface{}{
					map[string]interface{}{
						"origin_port": 443,
						"protocol":    "https",
					},
					map[string]interface{}{
						"origin_port": 80,
						"protocol":    "http",
					},
				},
			}
			err := dataSourceUpdateStateWithPayloadData(r.openAPIResource, remoteData, resourceData)
			Convey("Then the error should be nil and the expectedValue should equal to the expectedValue coming from remote", func() {
				So(err, ShouldBeNil)
				// keys stores in the resource data struct are always snake case
				So(len(resourceData.Get(listOfStrings.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 2)
				So(resourceData.Get(listOfStrings.GetTerraformCompliantPropertyName()).([]interface{})[0], ShouldEqual, remoteData[listOfStrings.Name].([]interface{})[0])
				So(resourceData.Get(listOfStrings.GetTerraformCompliantPropertyName()).([]interface{})[1], ShouldEqual, remoteData[listOfStrings.Name].([]interface{})[1])

				So(len(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})), ShouldEqual, 2)
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["origin_port"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[0].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[0].(map[string]interface{})["protocol"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[0].(map[string]interface{})["protocol"])

				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{}), ShouldContainKey, "origin_port")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{}), ShouldContainKey, "protocol")
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{})["origin_port"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[1].(map[string]interface{})["origin_port"].(int))
				So(resourceData.Get(listOfObjectsProperty.GetTerraformCompliantPropertyName()).([]interface{})[1].(map[string]interface{})["protocol"], ShouldEqual, remoteData[listOfObjectsProperty.Name].([]interface{})[1].(map[string]interface{})["protocol"])
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
			Convey("Then the error should be nil and the resource state data only contains the properties and values for the documented properties", func() {
				So(err, ShouldBeNil)
				So(resourceData.Get(stringWithPreferredNameProperty.GetTerraformCompliantPropertyName()), ShouldEqual, remoteData[stringWithPreferredNameProperty.Name])
				So(resourceData.Get("some_other_property_not_documented_in_openapi_doc"), ShouldBeNil)
			})
		})
	})
}

func TestUpdateStateWithPayloadDataAndOptions(t *testing.T) {
	Convey("Given a resource factory containing a schema with property lists that have the IgnoreItemsOrder set to true", t, func() {
		specResource := &specStubResource{
			error: fmt.Errorf("some error"),
		}
		Convey("When updateStateWithPayloadDataAndOptions is called", func() {
			err := updateStateWithPayloadDataAndOptions(specResource, nil, nil, true)
			Convey("Then the err returned should match the expected one", func() {
				So(err, ShouldEqual, specResource.error)
			})
		})
	})
	Convey("Given a resource factory containing just a property ID", t, func() {
		specResource := &specStubResource{
			schemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{idProperty},
			},
		}
		Convey("When updateStateWithPayloadDataAndOptions is called", func() {
			remoteData := map[string]interface{}{
				idProperty.Name: "someID",
			}
			var resourceLocalData *schema.ResourceData
			err := updateStateWithPayloadDataAndOptions(specResource, remoteData, resourceLocalData, true)
			Convey("Then the error returned should be nil and the resource local data should be intact since the id property is ignored when updating the resource data file behind the scenes", func() {
				So(err, ShouldBeNil)
				So(resourceLocalData, ShouldEqual, nil)
			})
		})
	})
	Convey("Given a resource factory containing a property with certain type", t, func() {
		r, resourceData := testCreateResourceFactory(t, &SpecSchemaDefinitionProperty{
			Name:                 "wrong_property",
			Type:                 TypeObject,
			SpecSchemaDefinition: &SpecSchemaDefinition{},
		})
		Convey("When updateStateWithPayloadDataAndOptions is called with a remote data containing the property but the value does not match the property type", func() {
			remoteData := map[string]interface{}{
				"wrong_property": "someValueNotMatchingTheType",
			}
			err := updateStateWithPayloadDataAndOptions(r.openAPIResource, remoteData, resourceData, true)
			Convey("Then the err returned should match the expected one", func() {
				So(err.Error(), ShouldEqual, "wrong_property: '': source data must be an array or slice, got string")
			})
		})
	})
	Convey("Given a resource factory containing a property with a type that the remote value does not match", t, func() {
		r := &specStubResource{
			schemaDefinition: &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					&SpecSchemaDefinitionProperty{
						Name:           "not_well_configured_property",
						Type:           TypeList,
						ArrayItemsType: schemaDefinitionPropertyType("unknown"),
					},
				},
			},
		}
		Convey("When updateStateWithPayloadDataAndOptions", func() {
			remoteData := map[string]interface{}{
				"not_well_configured_property": []interface{}{"something"},
			}
			err := updateStateWithPayloadDataAndOptions(r, remoteData, nil, true)
			Convey("Then the err returned should match the expected one", func() {
				So(err.Error(), ShouldEqual, "property 'not_well_configured_property' is supposed to be an array objects")
			})
		})
	})
}

func TestConvertPayloadToLocalStateDataValue(t *testing.T) {

	Convey("Given a resource factory", t, func() {

		Convey("When convertPayloadToLocalStateDataValue is called with ", func() {
			property := newStringSchemaDefinitionPropertyWithDefaults("string_property", "", false, false, nil)
			dataValue := "someValue"
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the expected value with the right type string", func() {
				So(err, ShouldBeNil)
				So(resultValue, ShouldEqual, dataValue)
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with a bool property and a bool value", func() {
			property := newBoolSchemaDefinitionPropertyWithDefaults("bool_property", "", false, false, nil)
			dataValue := true
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the expected value with the right type boolean", func() {
				So(err, ShouldBeNil)
				So(resultValue, ShouldEqual, dataValue)
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with an int property and a int value", func() {
			property := newIntSchemaDefinitionPropertyWithDefaults("int_property", "", false, false, nil)
			dataValue := 10
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the expected value with the right type int", func() {
				So(err, ShouldBeNil)
				So(resultValue, ShouldEqual, dataValue)
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an float property and a float value", func() {
			property := newNumberSchemaDefinitionPropertyWithDefaults("float_property", "", false, false, nil)
			dataValue := 45.23
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then error should be nil and the result value should be the expected value formatted string with the right type float", func() {
				So(err, ShouldBeNil)
				So(resultValue, ShouldEqual, dataValue)
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an float property and a float value but the swagger property is an integer", func() {
			property := newIntSchemaDefinitionPropertyWithDefaults("int_property", "", false, false, nil)
			dataValue := 45
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the expected value formatted string with the right type integer", func() {
				So(err, ShouldBeNil)
				So(resultValue, ShouldEqual, dataValue)
				So(resultValue, ShouldHaveSameTypeAs, int(dataValue))
			})
		})
		Convey("When convertPayloadToLocalStateDataValue is called with an list property and a with items object", func() {
			objectSchemaDefinition := &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
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
			property := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, nil, TypeObject, objectSchemaDefinition)
			dataValue := []interface{}{objectDefault}
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the list containing the object items with the expected types (int, string, bool and float)", func() {
				So(err, ShouldBeNil)
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
			property := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, nil, TypeString, nil)
			dataValue := []interface{}{"value1"}
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the expected value with the right type array", func() {
				So(err, ShouldBeNil)
				So(resultValue.([]interface{}), ShouldContain, dataValue[0])
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with simple object property and an empty map as value", func() {
			property := &SpecSchemaDefinitionProperty{
				Name:     "some_object",
				Type:     TypeObject,
				Required: true,
			}
			resultValue, err := convertPayloadToLocalStateDataValue(property, map[string]interface{}{})
			Convey("Then the error should be nil and the result value should be the expected value with the right type array", func() {
				So(err, ShouldBeNil)
				So(resultValue.([]interface{}), ShouldNotBeEmpty) // By default objects' internal terraform schema is Type List with Max 1 elem *Resource
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldBeEmpty)
			})
		})

		// Edge case
		Convey("When convertPayloadToLocalStateDataValue is called with a slice of map interfaces", func() {
			property := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, nil, TypeString, nil)
			_, err := convertPayloadToLocalStateDataValue(property, []map[string]interface{}{})
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with a property list that the array items are of unknown type", func() {
			property := &SpecSchemaDefinitionProperty{
				Name:           "not_well_configured_property",
				Type:           TypeList,
				ArrayItemsType: schemaDefinitionPropertyType("unknown"),
			}
			_, err := convertPayloadToLocalStateDataValue(property, []interface{}{})
			Convey("Then the error should match the expected one", func() {
				So(err.Error(), ShouldEqual, "property 'not_well_configured_property' is supposed to be an array objects")
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with a simple object", func() {
			// Simple objects are considered objects that all the properties are of the same type and are not computed
			objectSchemaDefinition := &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
					newStringSchemaDefinitionPropertyWithDefaults("example_string", "", true, false, nil),
					newStringSchemaDefinitionPropertyWithDefaults("example_string_2", "", true, false, nil),
				},
			}
			dataValue := map[string]interface{}{
				"example_string":   "http",
				"example_string_2": "something",
			}
			property := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, false, nil, objectSchemaDefinition)
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the list containing the object items all being string type (as terraform only supports maps of strings, hence values need to be stored as strings)", func() {
				So(err, ShouldBeNil)
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_string"].(string), ShouldEqual, "http")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_string_2"].(string), ShouldEqual, "something")
			})
		})

		// Simple objects are considered objects that contain properties that are of different types and configuration (e,g: mix of required/optional/computed properties)
		Convey("When convertPayloadToLocalStateDataValue is called with a complex object", func() {
			objectSchemaDefinition := &SpecSchemaDefinition{
				Properties: SpecSchemaDefinitionProperties{
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
			resultValue, err := convertPayloadToLocalStateDataValue(property, dataValue)
			Convey("Then the error should be nil and the result value should be the list containing the object items all being string type (as terraform only supports maps of strings, hence values need to be stored as strings)", func() {
				So(err, ShouldBeNil)
				So(resultValue.([]interface{})[0], ShouldContainKey, "example_int")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_int"], ShouldEqual, 80)
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_string")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_string"], ShouldEqual, "http")
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_bool")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_bool"], ShouldEqual, true)
				So(resultValue.([]interface{})[0].(map[string]interface{}), ShouldContainKey, "example_float")
				So(resultValue.([]interface{})[0].(map[string]interface{})["example_float"], ShouldEqual, 10.45)
			})
		})

		Convey("When convertPayloadToLocalStateDataValue is called with an object containing objects", func() {
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
			// The below represents the JSON representation of the response payload received by the API
			dataValue := map[string]interface{}{
				"id":            propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
				"nested_object": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
			}

			expectedPropertyWithNestedObjectName := "property_with_nested_object"
			propertyWithNestedObject := newObjectSchemaDefinitionPropertyWithDefaults(expectedPropertyWithNestedObjectName, "", true, false, false, dataValue, propertyWithNestedObjectSchemaDefinition)
			resultValue, err := convertPayloadToLocalStateDataValue(propertyWithNestedObject, dataValue)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)

				// The result value should be the list containing just one element (as per the nested struct workaround)
				// Tag(NestedStructsWorkaround)
				// Note: This is the workaround needed to support properties with nested structs. The current Terraform sdk version
				// does not support this now, hence the suggestion from the Terraform maintainer was to use a list of map[string]interface{}
				// with the list containing just one element. The below represents the internal representation of the terraform state
				// for an object property that contains other objects
				So(resultValue.([]interface{}), ShouldNotBeEmpty)
				So(len(resultValue.([]interface{})), ShouldEqual, 1)

				// AND the object should have the expected properties including the nested object
				So(resultValue.([]interface{})[0], ShouldContainKey, propertyWithNestedObjectSchemaDefinition.Properties[0].Name)
				So(resultValue.([]interface{})[0], ShouldContainKey, propertyWithNestedObjectSchemaDefinition.Properties[1].Name)

				// AND the object property with nested object should have the expected configuration
				nestedObject := propertyWithNestedObjectSchemaDefinition.Properties[1]
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name].([]interface{})[0].(map[string]interface{}), ShouldContainKey, nestedObjectSchemaDefinition.Properties[0].Name)
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name].([]interface{})[0].(map[string]interface{})[nestedObjectSchemaDefinition.Properties[0].Name], ShouldEqual, nestedObjectSchemaDefinition.Properties[0].Default.(int))
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name].([]interface{})[0].(map[string]interface{}), ShouldContainKey, nestedObjectSchemaDefinition.Properties[1].Name)
				So(resultValue.([]interface{})[0].(map[string]interface{})[nestedObject.Name].([]interface{})[0].(map[string]interface{})[nestedObjectSchemaDefinition.Properties[1].Name], ShouldEqual, nestedObjectSchemaDefinition.Properties[1].Default)
			})
		})
	})
}

func TestSetResourceDataProperty(t *testing.T) {
	Convey("Given a resource data (state) loaded with couple propeprties", t, func() {
		_, resourceData := testCreateResourceFactory(t, stringProperty, stringWithPreferredNameProperty)
		Convey("When setResourceDataProperty is called with a schema definition property that exists in terraform resource data object and with a new expectedValue", func() {
			expectedValue := "newValue"
			err := setResourceDataProperty(*stringProperty, expectedValue, resourceData)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// keys stores in the resource data struct are always snake case
				So(resourceData.Get(stringProperty.Name), ShouldEqual, expectedValue)
			})
		})
		Convey("When setResourceDataProperty is called with a schema definition property preferred name that exists in terraform resource data object and with a new expectedValue", func() {
			expectedValue := "theNewValue"
			err := setResourceDataProperty(*stringWithPreferredNameProperty, expectedValue, resourceData)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// keys stores in the resource data struct are always snake case
				// note the state is queried using the preferred name
				So(resourceData.Get(stringWithPreferredNameProperty.PreferredName), ShouldEqual, expectedValue)
			})
		})
		Convey("When setResourceDataProperty is called with a schema definition property name does NOT exist", func() {
			err := setResourceDataProperty(SpecSchemaDefinitionProperty{Name: "nonExistingKey"}, "", resourceData)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				// keys stores in the resource data struct are always snake case
				So(err.Error(), ShouldEqual, `Invalid address to set: []string{"non_existing_key"}`)
			})
		})
	})
}

func TestSetStateID(t *testing.T) {
	Convey("Given a resource factory configured with a schema definition that as an id property", t, func() {
		r, resourceData := testCreateResourceFactory(t, idProperty)
		Convey("When setStateID is called with the resourceData and responsePayload", func() {
			responsePayload := map[string]interface{}{
				idProperty.Name: "idValue",
			}
			err := setStateID(r.openAPIResource, resourceData, responsePayload)
			Convey("Then resourceData should be populated with the values returned by the API including the ID", func() {
				So(err, ShouldBeNil)
				So(resourceData.Id(), ShouldEqual, responsePayload[idProperty.Name])
			})
		})

		Convey("When setStateID is called with a resourceData that contains an id property but the responsePayload does not have it", func() {
			responsePayload := map[string]interface{}{
				"someOtherProperty": "idValue",
			}
			err := setStateID(r.openAPIResource, resourceData, responsePayload)
			Convey("Then resourceData should be populated with the values returned by the API including the ID", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "response object returned from the API is missing mandatory identifier property 'id'")
			})
		})
	})

	Convey("Given a resource factory configured with a schema definition that DOES not have an id property but one of the properties is tagged as id", t, func() {
		r, resourceData := testCreateResourceFactory(t, someIdentifierProperty)
		Convey("When setStateID is called with the resourceData and responsePayload", func() {
			responsePayload := map[string]interface{}{
				someIdentifierProperty.Name: "idValue",
			}
			err := setStateID(r.openAPIResource, resourceData, responsePayload)
			Convey("Then resourceData should be populated with the values returned by the API including the ID", func() {
				So(err, ShouldBeNil)
				So(resourceData.Id(), ShouldEqual, responsePayload[someIdentifierProperty.Name])
			})
		})
	})

	Convey("Given a resource factory configured with a schema definition that DOES not have an id property nor a property that should be used as the identifier", t, func() {
		r, resourceData := testCreateResourceFactory(t)
		Convey("When setStateID is called with the resourceData and responsePayload", func() {
			responsePayload := map[string]interface{}{
				"someOtherProperty": "idValue",
			}
			err := setStateID(r.openAPIResource, resourceData, responsePayload)
			Convey("Then resourceData should be populated with the values returned by the API including the ID", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "could not find any identifier property in the resource schema definition")
			})
		})
	})
}

func TestProcessIgnoreOrderIfEnabled(t *testing.T) {
	testCases := []struct {
		name               string
		property           SpecSchemaDefinitionProperty
		inputPropertyValue interface{}
		remoteValue        interface{}
		expectedOutput     interface{}
	}{
		// String use cases
		{
			name: "required input list (of strings) matches the value returned by the API where order of input values match",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{"inputVal1", "inputVal2", "inputVal3"},
			remoteValue:        []interface{}{"inputVal1", "inputVal2", "inputVal3"},
			expectedOutput:     []interface{}{"inputVal1", "inputVal2", "inputVal3"},
		},
		{
			name: "required input list (of strings) matches the value returned by the API where order of input values doesn't match",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{"inputVal3", "inputVal1", "inputVal2"},
			remoteValue:        []interface{}{"inputVal2", "inputVal3", "inputVal1"},
			expectedOutput:     []interface{}{"inputVal3", "inputVal1", "inputVal2"},
		},
		{
			name: "required input list (of strings) has a value that isn't returned by the API (input order maintained)",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{"inputVal1", "inputVal2", "inputVal3"},
			remoteValue:        []interface{}{"inputVal2", "inputVal1"},
			expectedOutput:     []interface{}{"inputVal1", "inputVal2"},
		},
		{
			name: "required input list (of strings) is missing a value returned by the API (input order maintained)",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{"inputVal1", "inputVal2"},
			remoteValue:        []interface{}{"inputVal3", "inputVal2", "inputVal1"},
			expectedOutput:     []interface{}{"inputVal1", "inputVal2", "inputVal3"},
		},

		// Integer use cases
		{
			name: "required input list (of ints) matches the value returned by the API where order of input values doesn't match",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeInt,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{3, 1, 2},
			remoteValue:        []interface{}{2, 3, 1},
			expectedOutput:     []interface{}{3, 1, 2},
		},
		{
			name: "required input list (of ints) has a value that isn't returned by the API (input order maintained)",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeInt,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{1, 2, 3},
			remoteValue:        []interface{}{2, 1},
			expectedOutput:     []interface{}{1, 2},
		},
		{
			name: "required input list (of ints) is missing a value returned by the API (input order maintained)",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeInt,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{1, 2},
			remoteValue:        []interface{}{3, 2, 1},
			expectedOutput:     []interface{}{1, 2, 3},
		},

		// Float use cases
		{
			name: "required input list (of floats) matches the value returned by the API where order of input values doesn't match",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeFloat,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{3.0, 1.0, 2.0},
			remoteValue:        []interface{}{2.0, 3.0, 1.0},
			expectedOutput:     []interface{}{3.0, 1.0, 2.0},
		},
		{
			name: "required input list (of floats) has a value that isn't returned by the API (input order maintained)",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeFloat,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{1.0, 2.0, 3.0},
			remoteValue:        []interface{}{2.0, 1.0},
			expectedOutput:     []interface{}{1.0, 2.0},
		},
		{
			name: "required input list (of floats) is missing a value returned by the API (input order maintained)",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeFloat,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{1.0, 2.0},
			remoteValue:        []interface{}{3.0, 2.0, 1.0},
			expectedOutput:     []interface{}{1.0, 2.0, 3.0},
		},

		// List of objects use cases
		{
			name: "required input list (objects) matches the value returned by the API where order of input values doesn't match",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				IgnoreItemsOrder: true,
				Type:             TypeList,
				ArrayItemsType:   TypeObject,
				SpecSchemaDefinition: &SpecSchemaDefinition{
					Properties: SpecSchemaDefinitionProperties{
						&SpecSchemaDefinitionProperty{
							Name: "group",
							Type: TypeString,
						},
						&SpecSchemaDefinitionProperty{
							Name:           "roles",
							Type:           TypeList,
							ArrayItemsType: TypeString,
						},
					},
				},
			},
			inputPropertyValue: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
				map[string]interface{}{
					"group": "someOtherGroup",
					"roles": []interface{}{"role3", "role4"},
				},
			},
			remoteValue: []interface{}{
				map[string]interface{}{
					"group": "someOtherGroup",
					"roles": []interface{}{"role3", "role4"},
				},
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
				map[string]interface{}{
					"group": "someOtherGroup",
					"roles": []interface{}{"role3", "role4"},
				},
			},
		},
		{
			name: "required input list (objects) has a value that isn't returned by the API (input order maintained)",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				IgnoreItemsOrder: true,
				Type:             TypeList,
				ArrayItemsType:   TypeObject,
				SpecSchemaDefinition: &SpecSchemaDefinition{
					Properties: SpecSchemaDefinitionProperties{
						&SpecSchemaDefinitionProperty{
							Name: "group",
							Type: TypeString,
						},
						&SpecSchemaDefinitionProperty{
							Name:           "roles",
							Type:           TypeList,
							ArrayItemsType: TypeString,
						},
					},
				},
			},
			inputPropertyValue: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
				map[string]interface{}{
					"group": "someOtherGroup",
					"roles": []interface{}{"role3", "role4"},
				},
			},
			remoteValue: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
			},
		},
		{
			name: "required input list (objects) doesn't have a value returned by the API",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				IgnoreItemsOrder: true,
				Type:             TypeList,
				ArrayItemsType:   TypeObject,
				SpecSchemaDefinition: &SpecSchemaDefinition{
					Properties: SpecSchemaDefinitionProperties{
						&SpecSchemaDefinitionProperty{
							Name: "group",
							Type: TypeString,
						},
						&SpecSchemaDefinitionProperty{
							Name:           "roles",
							Type:           TypeList,
							ArrayItemsType: TypeString,
						},
					},
				},
			},
			inputPropertyValue: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
			},
			remoteValue: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
				map[string]interface{}{
					"group": "unexpectedGroup",
					"roles": []interface{}{"role3", "role4"},
				},
			},
			expectedOutput: []interface{}{
				map[string]interface{}{
					"group": "someGroup",
					"roles": []interface{}{"role1", "role2"},
				},
				map[string]interface{}{
					"group": "unexpectedGroup",
					"roles": []interface{}{"role3", "role4"},
				},
			},
		},
		{
			name: "inputPropertyValue is nil",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: nil,
			remoteValue:        []interface{}{},
			expectedOutput:     []interface{}{},
		},
		{
			name: "remoteValue is nil",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: true,
			},
			inputPropertyValue: []interface{}{},
			remoteValue:        nil,
			expectedOutput:     nil,
		},
		{
			name: "IgnoreItemsOrder is set to false",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeString,
				IgnoreItemsOrder: false,
			},
			inputPropertyValue: []interface{}{"inputVal1"},
			remoteValue:        []interface{}{"inputVal1", "inputVal2"},
			expectedOutput:     []interface{}{"inputVal1", "inputVal2"},
		},
		{
			name: "list of bools property definition and the corresponding input/remote lists",
			property: SpecSchemaDefinitionProperty{
				Name:             "list_prop",
				Type:             TypeList,
				ArrayItemsType:   TypeBool,
				IgnoreItemsOrder: true,
				Required:         true,
			},
			inputPropertyValue: []interface{}{true},
			remoteValue:        []interface{}{false},
			expectedOutput:     []interface{}{false},
		},
	}

	for _, tc := range testCases {
		output := processIgnoreOrderIfEnabled(tc.property, tc.inputPropertyValue, tc.remoteValue)
		assert.Equal(t, tc.expectedOutput, output, tc.name)
	}
}
