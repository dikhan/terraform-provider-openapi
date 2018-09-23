package openapi

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateTerraformResource(t *testing.T) {
	Convey("Given a resource factory initialised with a spec resource that has an id and string property and supports all CRUD operations", t, func() {
		r, resourceData := testCreateResourceFactory(t, idProperty, stringProperty)
		Convey("When createTerraformResource is called", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     idProperty.Default,
					stringProperty.Name: stringProperty.Default,
				},
			}
			schemaResource, err := r.createTerraformResource()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema resource should not be empty", func() {
				So(schemaResource.Schema, ShouldNotBeEmpty)
			})
			Convey("And the create function is invokable and returns nil error", func() {
				err := schemaResource.Create(resourceData, client)
				So(err, ShouldBeNil)
			})
			Convey("And the read function is invokable and returns nil error", func() {
				err := schemaResource.Read(resourceData, client)
				So(err, ShouldBeNil)
			})
			Convey("And the update function is invokable and returns nil error", func() {
				err := schemaResource.Update(resourceData, client)
				So(err, ShouldBeNil)
			})
			Convey("And the delete function is invokable and returns nil error", func() {
				err := schemaResource.Delete(resourceData, client)
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestCreateResourceSchema(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, _ := testCreateResourceFactory(t, idProperty, stringProperty)
		Convey("When createResourceSchema is called", func() {
			schema, err := r.createResourceSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema returned should not contain the ID property as schema alreayd has a reserved ID field to store the unique identifier", func() {
				So(schema, ShouldNotContainKey, idProperty.Name)
			})
			Convey("And the schema returned should contain the resource properties", func() {
				So(schema, ShouldContainKey, stringProperty.Name)
			})
		})
	})
}

func TestCreateTerraformPropertySchema(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r := resourceFactory{}
		Convey("When createTerraformPropertySchema is called with a schema definition property that is required, force new, sensitive and has a default value", func() {
			s := newStringSchemaDefinitionProperty("propertyName", "", true, false, true, true, false, false, "defaultValue")
			terraformPropertySchema, err := r.createTerraformPropertySchema(s)
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
			s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, "")
			terraformPropertySchema, err := r.createTerraformPropertySchema(s)
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
			s := newStringSchemaDefinitionProperty("propertyName", "", false, true, false, false, false, false, "defaultValue")
			terraformPropertySchema, err := r.createTerraformPropertySchema(s)
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
			s := newStringSchemaDefinitionProperty("propertyName", "", false, false, true, false, true, false, "")
			terraformPropertySchema, err := r.createTerraformPropertySchema(s)
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
			s := newStringSchemaDefinitionProperty("propertyName", "", true, true, false, false, false, false, nil)
			terraformPropertySchema, err := r.createTerraformPropertySchema(s)
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

func TestCreate(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactory(t, idProperty, stringProperty)
		Convey("When create is called with resource data and a client", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     "someID",
					stringProperty.Name: "someExtraValueThatProvesResponseDataIsPersisted",
				},
			}
			err := r.create(resourceData, client)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(resourceData.Id(), ShouldEqual, client.responsePayload[idProperty.Name])
				So(resourceData.Get(stringProperty.Name), ShouldEqual, client.responsePayload[stringProperty.Name])
			})
		})
		Convey("When create is called with resource data and a client configured to return an error when POST is called", func() {
			createError := fmt.Errorf("some error when deleting")
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name: idProperty.Default,
				},
				error: createError,
			}
			err := r.create(resourceData, client)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error returned should be the error returned by the client delete operation", func() {
				So(err, ShouldEqual, createError)
			})
		})

		Convey("When update is called with resource data and a client returns a non expected http code", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
				returnHTTPCode:  http.StatusInternalServerError,
			}
			err := r.create(resourceData, client)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should be", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] POST /v1/resource failed: [resource='resourceName'] HTTP Response Status Code 500 not matching expected one [201 202] ()")
			})
		})

		Convey("When update is called with resource data and a client returns a response that does not have an id property", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
			}
			err := r.create(resourceData, client)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should be", func() {
				So(err.Error(), ShouldEqual, "response object returned from the API is missing mandatory identifier property 'id'")
			})
		})
	})
}

func TestRead(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactory(t, idProperty, stringProperty)
		Convey("When readRemote is called with resource data and a client that returns ", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someOtherStringValue",
				},
			}
			err := r.read(resourceData, client)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And resourceData values should be the values got from the response payload (original values)", func() {
				So(resourceData.Get(stringProperty.Name), ShouldEqual, client.responsePayload[stringProperty.Name])
			})
		})
	})
}

func TestReadRemote(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, _ := testCreateResourceFactoryWithID(t, idProperty, stringProperty)
		Convey("When readRemote is called with resource data and a client that returns ", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     "someID",
					stringProperty.Name: "someOtherStringValue",
				},
			}
			response, err := r.readRemote("", client)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the map returned should be contain the properties in the response payload", func() {
				So(response, ShouldContainKey, idProperty.Name)
				So(response, ShouldContainKey, stringProperty.Name)
			})
			Convey("And the values of the keys should match the values that came in the response", func() {
				So(response[idProperty.Name], ShouldEqual, client.responsePayload[idProperty.Name])
				So(response[stringProperty.Name], ShouldEqual, client.responsePayload[stringProperty.Name])
			})
		})

		Convey("When readRemote is called with resource data and a client returns a non expected http code", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
				returnHTTPCode:  http.StatusInternalServerError,
			}
			_, err := r.readRemote("", client)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should be", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] GET /v1/resource/ failed: [resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] ()")
			})
		})
	})
}

func TestUpdate(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty, stringProperty, immutableProperty)
		Convey("When update is called with resource data and a client", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name:    "someExtraValueThatProvesResponseDataIsPersisted",
					immutableProperty.Name: immutableProperty.Default,
				},
			}
			err := r.update(resourceData, client)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And resourceData should contain the ID property", func() {
				So(resourceData.Id(), ShouldEqual, idProperty.Default)
			})
			Convey("And resourceData should be populated with the values returned by the API", func() {
				So(resourceData.Get(stringProperty.Name), ShouldEqual, client.responsePayload[stringProperty.Name])
				So(resourceData.Get(immutableProperty.Name), ShouldEqual, client.responsePayload[immutableProperty.Name])
			})
		})
		Convey("When update is called with a resource data containing updated values and the immutable check fails due to an immutable property being updated", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name:    "stringOriginalValue",
					immutableProperty.Name: "immutableOriginalValue",
				},
			}
			err := r.update(resourceData, client)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should equal ", func() {
				So(err.Error(), ShouldEqual, "property string_immutable_property is immutable and therefore can not be updated. Update operation was aborted; no updates were performed")
			})
			Convey("And resourceData values should be the values got from the response payload (original values)", func() {
				So(resourceData.Id(), ShouldEqual, idProperty.Default)
				So(resourceData.Get(stringProperty.Name), ShouldEqual, client.responsePayload[stringProperty.Name])
				So(resourceData.Get(immutableProperty.Name), ShouldEqual, client.responsePayload[immutableProperty.Name])
			})
		})
		Convey("When update is called with resource data and a client configured to return an error when delete is called", func() {
			deleteError := fmt.Errorf("some error when deleting")
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name: idProperty.Default,
				},
				error: deleteError,
			}
			err := r.update(resourceData, client)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error returned should be the error returned by the client delete operation", func() {
				So(err, ShouldEqual, deleteError)
			})
		})

		Convey("When update is called with resource data and a client returns a non expected http code", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
				returnHTTPCode:  http.StatusInternalServerError,
			}
			err := r.update(resourceData, client)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should be", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] GET /v1/resource/id failed: [resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] ()")
			})
		})
	})

	Convey("Given a resource factory with no delete operation configured", t, func() {
		specResource := newSpecStubResource("resourceName", "/v1/resource", false, nil)
		r := resourceFactory{specResource}
		Convey("When update is called with resource data and a client", func() {
			client := &clientOpenAPIStub{}
			err := r.delete(nil, client)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] resource does not support DELETE opperation, check the swagger file exposed on '/v1/resource'")
			})
		})
	})
}

func TestDelete(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty)
		Convey("When delete is called with resource data and a client", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name: idProperty.Default,
				},
			}
			err := r.delete(resourceData, client)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the expectedValue returned should be true", func() {
				So(client.responsePayload, ShouldNotContainKey, idProperty.Name)
			})
		})
		Convey("When delete is called with resource data and a client configured to return an error when delete is called", func() {
			deleteError := fmt.Errorf("some error when deleting")
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name: idProperty.Default,
				},
				error: deleteError,
			}
			err := r.delete(resourceData, client)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error returned should be the error returned by the client delete operation", func() {
				So(err, ShouldEqual, deleteError)
			})
		})

		Convey("When update is called with resource data and a client returns a non expected http code", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
				returnHTTPCode:  http.StatusInternalServerError,
			}
			err := r.delete(resourceData, client)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should be", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] DELETE /v1/resource/id failed: [resource='resourceName'] HTTP Response Status Code 500 not matching expected one [204 200] ()")
			})
		})
	})

	Convey("Given a resource factory with no delete operation configured", t, func() {
		specResource := newSpecStubResource("resourceName", "/v1/resource", false, nil)
		r := resourceFactory{specResource}
		Convey("When delete is called with resource data and a client", func() {
			client := &clientOpenAPIStub{}
			err := r.delete(nil, client)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] resource does not support DELETE opperation, check the swagger file exposed on '/v1/resource'")
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
			err := r.setStateID(resourceData, responsePayload)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(resourceData.Id(), ShouldEqual, responsePayload[idProperty.Name])
			})
		})

		Convey("When setStateID is called with a resourceData that contains an id property but the responsePayload does not have it", func() {
			responsePayload := map[string]interface{}{
				"someOtherProperty": "idValue",
			}
			err := r.setStateID(resourceData, responsePayload)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
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
			err := r.setStateID(resourceData, responsePayload)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
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
			err := r.setStateID(resourceData, responsePayload)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(err.Error(), ShouldEqual, "could not find any identifier property in the resource schema definition")
			})
		})
	})
}

func TestUpdateLocalState(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactory(t, idProperty, computedProperty, stringProperty, intProperty, numberProperty, boolProperty, sliceProperty)
		Convey("When updateLocalState is called ", func() {
			responsePayload := map[string]interface{}{
				idProperty.Name:       "idValue",
				computedProperty.Name: "someComputedValue",
				intProperty.Name:      intProperty.Default,
				numberProperty.Name:   numberProperty.Default,
				boolProperty.Name:     boolProperty.Default,
				sliceProperty.Name:    sliceProperty.Default,
			}
			err := r.updateLocalState(resourceData, responsePayload)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(resourceData.Id(), ShouldEqual, responsePayload[idProperty.Name])
				So(resourceData.Get(computedProperty.Name), ShouldEqual, responsePayload[computedProperty.Name])
				So(resourceData.Get(intProperty.Name), ShouldEqual, responsePayload[intProperty.Name])
				So(resourceData.Get(numberProperty.Name), ShouldEqual, responsePayload[numberProperty.Name])
				So(resourceData.Get(boolProperty.Name), ShouldEqual, responsePayload[boolProperty.Name])
				So(resourceData.Get(sliceProperty.Name), ShouldContain, sliceProperty.Default.([]string)[0])
			})
		})
	})
}

func TestCheckHTTPStatusCode(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, _ := testCreateResourceFactory(t)
		Convey("When checkHTTPStatusCode is called with a response containing a status codes that matches one of the expected response status codes", func() {
			response := &http.Response{
				StatusCode: http.StatusOK,
			}
			expectedStatusCodes := []int{http.StatusOK}
			err := r.checkHTTPStatusCode(response, expectedStatusCodes)
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
			err := r.checkHTTPStatusCode(response, expectedStatusCodes)
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
			err := r.checkHTTPStatusCode(response, expectedStatusCodes)
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
		r, _ := testCreateResourceFactory(t)
		Convey("When responseContainsExpectedStatus is called with a response code that exists in the given list of expected status codes", func() {
			expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
			responseCode := http.StatusCreated
			exists := r.responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the expectedValue returned should be true", func() {
				So(exists, ShouldBeTrue)
			})
		})
		Convey("When responseContainsExpectedStatus is called with a response code that DOES NOT exists in 'expectedResponseStatusCodes'", func() {
			expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
			responseCode := http.StatusUnauthorized
			exists := r.responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the expectedValue returned should be false", func() {
				So(exists, ShouldBeFalse)
			})
		})
	})
}

func TestCheckImmutableFields(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactory(t, immutableProperty, nonImmutableProperty)
		Convey("When checkImmutableFields is called with an update resource data and an open api client that returns the old expectedValue of the property being changed", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					immutableProperty.Name:    "originalImmutablePropertyValue",
					nonImmutableProperty.Name: "originalNonImmutablePropertyValue",
				},
			}
			err := r.checkImmutableFields(resourceData, client)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the err message returned should be", func() {
				So(err.Error(), ShouldEqual, fmt.Sprintf("property %s is immutable and therefore can not be updated. Update operation was aborted; no updates were performed", immutableProperty.Name))
			})
			Convey("And the resource data should contain the original values coming from the responsePayload (so it's assured that local state was not updated)", func() {
				So(resourceData.Get(immutableProperty.Name), ShouldEqual, client.responsePayload[immutableProperty.Name])
				So(resourceData.Get(nonImmutableProperty.Name), ShouldEqual, client.responsePayload[nonImmutableProperty.Name])
			})
		})
	})
}

func TestUpdateStateWithPayloadData(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactory(t, stringWithPreferredNameProperty)
		Convey("When  is called with a map containing some properties", func() {
			remoteData := map[string]interface{}{
				stringWithPreferredNameProperty.Name: "someUpdatedStringValue",
			}
			err := r.updateStateWithPayloadData(remoteData, resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the expectedValue should equal to the expectedValue coming from remote, and also the key expectedValue should be the preferred as defined in the property", func() {
				// keys stores in the resource data struct are always snake case
				So(resourceData.Get(stringWithPreferredNameProperty.PreferredName), ShouldEqual, "someUpdatedStringValue")
			})
		})
	})
}

func TestCreatePayloadFromLocalStateData(t *testing.T) {
	Convey("Given a resource factory initialized with a spec resource with some schema definition", t, func() {
		r, resourceData := testCreateResourceFactory(t, idProperty, computedProperty, stringProperty, intProperty, numberProperty, boolProperty, sliceProperty)
		Convey("When createPayloadFromLocalStateData is called with a terraform resource data", func() {
			payload := r.createPayloadFromLocalStateData(resourceData)
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should not include the following keys as they are either an identifier or read only (computed) properties", func() {
				So(payload, ShouldNotContainKey, idProperty.Name)
				So(payload, ShouldNotContainKey, computedProperty.Name)
			})
			Convey("And then payload returned should include the following keys ", func() {
				So(payload, ShouldContainKey, stringProperty.Name)
				So(payload, ShouldContainKey, intProperty.Name)
				So(payload, ShouldContainKey, numberProperty.Name)
				So(payload, ShouldContainKey, boolProperty.Name)
				So(payload, ShouldContainKey, sliceProperty.Name)
			})
			Convey("And then payload key values should match the values stored in the terraform resource data", func() {
				So(payload[stringProperty.Name], ShouldEqual, stringProperty.Default)
				So(payload[intProperty.Name], ShouldEqual, intProperty.Default)
				So(payload[numberProperty.Name], ShouldEqual, numberProperty.Default)
				So(payload[boolProperty.Name], ShouldEqual, boolProperty.Default)
				So(payload[sliceProperty.Name], ShouldContain, sliceProperty.Default.([]string)[0])
			})
		})
	})

	Convey("Given a resource factory initialized with a spec resource with some schema definition and zero values", t, func() {
		r, resourceData := testCreateResourceFactory(t, intZeroValueProperty, numberZeroValueProperty, boolZeroValueProperty, sliceZeroValueProperty)
		Convey("When createPayloadFromLocalStateData is called with a terraform resource data", func() {
			payload := r.createPayloadFromLocalStateData(resourceData)
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should include the following keys ", func() {
				So(payload, ShouldContainKey, intZeroValueProperty.Name)
				So(payload, ShouldContainKey, numberZeroValueProperty.Name)
				So(payload, ShouldContainKey, boolZeroValueProperty.Name)
				So(payload, ShouldContainKey, sliceZeroValueProperty.Name)
			})
			Convey("And then payload key values should match the values stored in the terraform resource data", func() {
				So(payload[intZeroValueProperty.Name], ShouldEqual, intZeroValueProperty.Default)
				So(payload[numberZeroValueProperty.Name], ShouldEqual, numberZeroValueProperty.Default)
				So(payload[boolZeroValueProperty.Name], ShouldEqual, boolZeroValueProperty.Default)
			})
		})
	})
}

func TestSetResourceDataProperty(t *testing.T) {
	Convey("Given a resource factory initialized with a spec resource with some schema definition", t, func() {
		r, resourceData := testCreateResourceFactory(t, stringProperty, stringWithPreferredNameProperty)
		Convey("When setResourceDataProperty is called with a schema definition property name that exists in terraform resource data object and with a new expectedValue", func() {
			expectedValue := "newValue"
			err := r.setResourceDataProperty(stringProperty.Name, expectedValue, resourceData)
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
			err := r.setResourceDataProperty(stringWithPreferredNameProperty.Name, expectedValue, resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And then expectedValue should equal the expected value (note the state is queried using the preferred name)", func() {
				// keys stores in the resource data struct are always snake case
				So(resourceData.Get(stringWithPreferredNameProperty.PreferredName), ShouldEqual, expectedValue)
			})
		})
		Convey("When setResourceDataProperty is called with a schema definition property name does NOT exist", func() {
			err := r.setResourceDataProperty("nonExistingKey", "", resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And then expectedValue should equal", func() {
				// keys stores in the resource data struct are always snake case
				So(err.Error(), ShouldEqual, "could not find schema definition property name nonExistingKey in the resource data")
			})
		})
	})
}

func TestGetResourceDataOKExists(t *testing.T) {
	Convey("Given a resource factory initialized with a spec resource with some schema definition and resource data", t, func() {
		r, resourceData := testCreateResourceFactory(t, stringProperty, stringWithPreferredNameProperty)
		Convey("When getResourceDataOKExists is called with a schema definition property name that exists in terraform resource data object", func() {
			value, exists := r.getResourceDataOKExists(stringProperty.Name, resourceData)
			Convey("Then the bool returned should be true", func() {
				So(exists, ShouldBeTrue)
			})
			Convey("And then expectedValue should equal", func() {
				So(value, ShouldEqual, stringProperty.Default)
			})
		})

		Convey("When getResourceDataOKExists is called with a schema definition property name that has a preferred name and that exists in terraform resource data object", func() {
			value, exists := r.getResourceDataOKExists(stringWithPreferredNameProperty.Name, resourceData)
			Convey("Then the bool returned should be true", func() {
				So(exists, ShouldBeTrue)
			})
			Convey("And then expectedValue should equal", func() {
				So(value, ShouldEqual, stringWithPreferredNameProperty.Default)
			})
		})

		Convey("When getResourceDataOKExists is called with a schema definition property name that DOES NOT exists in terraform resource data object", func() {
			_, exists := r.getResourceDataOKExists("nonExistingProperty", resourceData)
			Convey("Then the bool returned should be true", func() {
				So(exists, ShouldBeFalse)
			})
		})
	})

	Convey("Given a resource factory initialized with a spec resource with some schema definition and resource data", t, func() {
		var stringPropertyWithNonCompliantName = newStringSchemaDefinitionPropertyWithDefaults("stringProperty", "", true, false, "updatedValue")
		r, resourceData := testCreateResourceFactory(t, stringPropertyWithNonCompliantName)
		Convey("When getResourceDataOKExists is called with a schema definition property name that exists in terraform resource data object", func() {
			value, exists := r.getResourceDataOKExists(stringPropertyWithNonCompliantName.Name, resourceData)
			Convey("Then the bool returned should be true", func() {
				So(exists, ShouldBeTrue)
			})
			Convey("And then expectedValue should equal", func() {
				So(value, ShouldEqual, stringPropertyWithNonCompliantName.Default)
			})
		})
	})
}

// testCreateResourceFactoryWithID configures the resourceData with the Id field. This is used for tests that rely on the
// resource state to be fully created. For instance, update or delete operations.
func testCreateResourceFactoryWithID(t *testing.T, idSchemaDefinitionProperty *SchemaDefinitionProperty, schemaDefinitionProperties ...*SchemaDefinitionProperty) (resourceFactory, *schema.ResourceData) {
	schemaDefinitionProperties = append(schemaDefinitionProperties, idSchemaDefinitionProperty)
	resourceFactory, resourceData := testCreateResourceFactory(t, schemaDefinitionProperties...)
	resourceData.SetId(idSchemaDefinitionProperty.Default.(string))
	return resourceFactory, resourceData
}

// testCreateResourceFactory configures the resourceData with some properties.
func testCreateResourceFactory(t *testing.T, schemaDefinitionProperties ...*SchemaDefinitionProperty) (resourceFactory, *schema.ResourceData) {
	testSchema := newTestSchema(schemaDefinitionProperties...)
	resourceData := testSchema.getResourceData(t)
	specResource := newSpecStubResourceWithOperations("resourceName", "/v1/resource", false, testSchema.getSchemaDefinition(), &ResourceOperation{}, &ResourceOperation{}, &ResourceOperation{}, &ResourceOperation{})
	return resourceFactory{specResource}, resourceData
}
