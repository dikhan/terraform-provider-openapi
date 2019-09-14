package openapi

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/go-openapi/spec"

	"github.com/hashicorp/terraform/helper/schema"

	. "github.com/smartystreets/goconvey/convey"
)

func TestCreateSchemaResourceTimeout(t *testing.T) {
	Convey("Given a resource factory initialised with a spec resource that has some timeouts", t, func() {
		duration, _ := time.ParseDuration("30m")
		expectedTimeouts := &specTimeouts{
			Get:    &duration,
			Post:   &duration,
			Put:    &duration,
			Delete: &duration,
		}
		r := newResourceFactory(&specStubResource{
			timeouts: expectedTimeouts,
		})
		Convey("When createSchemaResourceTimeout is called", func() {
			timeouts, err := r.createSchemaResourceTimeout()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the timeouts should match the expected ones", func() {
				So(timeouts.Read, ShouldEqual, expectedTimeouts.Get)
				So(timeouts.Create, ShouldEqual, expectedTimeouts.Post)
				So(timeouts.Delete, ShouldEqual, expectedTimeouts.Delete)
				So(timeouts.Update, ShouldEqual, expectedTimeouts.Put)
			})
		})
	})
}

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

func TestCreateTerraformResourceSchema(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, _ := testCreateResourceFactory(t, idProperty, stringProperty)
		Convey("When createResourceSchema is called", func() {
			schema, err := r.createTerraformResourceSchema()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the schema returned should not contain the ID property as schema already has a reserved ID field to store the unique identifier", func() {
				So(schema, ShouldNotContainKey, idProperty.Name)
			})
			Convey("And the schema returned should contain the resource properties", func() {
				So(schema, ShouldContainKey, stringProperty.Name)
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
				So(err.Error(), ShouldEqual, "[resource='resourceName'] POST /v1/resource failed: [resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200 201 202] ()")
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

	Convey("Given a resource factory with an empty OpenAPI resource", t, func() {
		r := resourceFactory{}
		Convey("When create is called with empty data and a empty client", func() {
			client := &clientOpenAPIStub{}
			err := r.create(nil, client)
			Convey("Then the error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a resource factory that has an asynchronous create operation (post) but the polling operation fails for some reason", t, func() {
		expectedReturnCode := 202
		testSchema := newTestSchema(idProperty, stringProperty)
		resourceData := testSchema.getResourceData(t)
		specResource := newSpecStubResourceWithOperations("resourceName", "/v1/resource", false, testSchema.getSchemaDefinition(), &specResourceOperation{responses: specResponses{expectedReturnCode: &specResponse{isPollingEnabled: true}}}, &specResourceOperation{}, &specResourceOperation{}, &specResourceOperation{})
		r := resourceFactory{
			openAPIResource: specResource,
			defaultTimeout:  time.Duration(0 * time.Second),
		}
		Convey("When create is called with resource data and a client", func() {
			client := &clientOpenAPIStub{
				returnHTTPCode: expectedReturnCode,
				responsePayload: map[string]interface{}{
					idProperty.Name:     "someID",
					stringProperty.Name: "someExtraValueThatProvesResponseDataIsPersisted",
				},
			}
			err := r.create(resourceData, client)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "polling mechanism failed after POST /v1/resource call with response status code (202): error waiting for resource to reach a completion status ([]) [valid pending statuses ([])]: error on retrieving resource 'resourceName' (someID) when waiting: [resource='resourceName'] HTTP Response Status Code 202 not matching expected one [200] ()")
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

	Convey("Given a resource factory configured with a resource that is a subresource", t, func() {

		someOtherProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")

		// Pretending the data has already been populated with the parent property
		testSchema := newTestSchema(idProperty, someOtherProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		r := newResourceFactory(&SpecV2Resource{
			Path: "/v1/cdns/{id}/firewall",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"some_string_prop": spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: []string{"string"},
							},
						},
					},
				},
			},
		})
		Convey("When readRemote is called with resource data (containing the expected state property values) and a provider client configured to return some response payload", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					someOtherProperty.Name: "someOtherStringValue",
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

	Convey("Given a resource factory with an empty OpenAPI resource", t, func() {
		r := resourceFactory{}
		Convey("When create is called with empty data and a empty client", func() {
			client := &clientOpenAPIStub{}
			err := r.read(nil, client)
			Convey("Then the error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestReadRemote(t *testing.T) {

	Convey("Given a resource factory", t, func() {
		r := newResourceFactory(&specStubResource{name: "resourceName"})
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
				So(err.Error(), ShouldEqual, "[resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] ()")
			})
		})

		Convey("When readRemote is called with an instance id, a providerClient and a parent ID", func() {
			expectedID := "someID"
			expectedParentID := "someParentID"
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     "someID",
					stringProperty.Name: "someOtherStringValue",
				},
			}
			response, err := r.readRemote(expectedID, client, expectedParentID)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the provider client should have been called with the right argument values", func() {
				So(client.idReceived, ShouldEqual, expectedID)
				So(len(client.parentIDsReceived), ShouldEqual, 1)
				So(client.parentIDsReceived[0], ShouldEqual, expectedParentID)
			})
			Convey("And the values of the keys should match the values that came in the response", func() {
				So(response[idProperty.Name], ShouldEqual, client.responsePayload[idProperty.Name])
				So(response[stringProperty.Name], ShouldEqual, client.responsePayload[stringProperty.Name])
			})
		})
	})
}

func TestUpdate(t *testing.T) {
	Convey("Given a resource factory containing some properties including an immutable property", t, func() {
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
		Convey("When update is called with resource data and a client configured to return an error when update is called", func() {
			updateError := fmt.Errorf("some error when deleting")
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name: idProperty.Default,
				},
				error: updateError,
			}
			err := r.update(resourceData, client)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error returned should be the error returned by the client update operation", func() {
				So(err, ShouldEqual, updateError)
			})
		})
	})

	Convey("Given a resource factory containing some properties", t, func() {
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty, stringProperty)
		Convey("When update is called with resource data and a client returns a non expected http code when reading remote", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someExtraValueThatProvesResponseDataIsPersisted",
				},
				returnHTTPCode: http.StatusInternalServerError,
			}
			err := r.update(resourceData, client)
			Convey("And the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] UPDATE /v1/resource/id failed: [resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200 202] ()")
			})
		})
		Convey("When update is called with resource data and a client returns a non expected error", func() {
			expectedError := "some error returned by the PUT operation"
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someExtraValueThatProvesResponseDataIsPersisted",
				},
				error: fmt.Errorf(expectedError),
			}
			err := r.update(resourceData, client)
			Convey("And the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, expectedError)
			})
		})
	})

	Convey("Given a resource factory with no update operation configured", t, func() {
		specResource := newSpecStubResource("resourceName", "/v1/resource", false, nil)
		r := newResourceFactory(specResource)
		Convey("When update is called with resource data and a client", func() {
			client := &clientOpenAPIStub{}
			err := r.update(nil, client)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] resource does not support PUT operation, check the swagger file exposed on '/v1/resource'")
			})
		})
	})

	Convey("Given a resource factory with an empty OpenAPI resource", t, func() {
		r := resourceFactory{}
		Convey("When create is called with empty data and a empty client", func() {
			client := &clientOpenAPIStub{}
			err := r.update(nil, client)
			Convey("Then the error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a resource factory that has an asynchronous create operation (put) but the polling operation fails for some reason", t, func() {
		expectedReturnCode := 202
		testSchema := newTestSchema(idProperty, stringProperty)
		resourceData := testSchema.getResourceData(t)
		specResource := newSpecStubResourceWithOperations("resourceName", "/v1/resource", false, testSchema.getSchemaDefinition(), &specResourceOperation{}, &specResourceOperation{responses: specResponses{expectedReturnCode: &specResponse{isPollingEnabled: true}}}, &specResourceOperation{}, &specResourceOperation{})
		r := resourceFactory{
			openAPIResource: specResource,
			defaultTimeout:  time.Duration(0 * time.Second),
		}
		Convey("When create is called with resource data and a client", func() {
			client := &clientOpenAPIStub{
				returnHTTPCode: expectedReturnCode,
				responsePayload: map[string]interface{}{
					idProperty.Name:     "someID",
					stringProperty.Name: "someExtraValueThatProvesResponseDataIsPersisted",
				},
			}
			err := r.update(resourceData, client)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "polling mechanism failed after PUT /v1/resource call with response status code (202): error waiting for resource to reach a completion status ([]) [valid pending statuses ([])]: error on retrieving resource 'resourceName' () when waiting: [resource='resourceName'] HTTP Response Status Code 202 not matching expected one [200] ()")
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
				So(err.Error(), ShouldEqual, "[resource='resourceName'] DELETE /v1/resource/id failed: [resource='resourceName'] HTTP Response Status Code 500 not matching expected one [204 200 202] ()")
			})
		})

		Convey("When update is called with resource data and a client returns a 404 status code; hence the resource effectively no longer exists", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
				returnHTTPCode:  http.StatusNotFound,
			}
			err := r.delete(resourceData, client)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a resource factory with no delete operation configured", t, func() {
		specResource := newSpecStubResource("resourceName", "/v1/resource", false, nil)
		r := newResourceFactory(specResource)
		Convey("When delete is called with resource data and a client", func() {
			client := &clientOpenAPIStub{}
			err := r.delete(nil, client)
			Convey("Then the expectedValue returned should be true", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And resourceData should be populated with the values returned by the API including the ID", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] resource does not support DELETE operation, check the swagger file exposed on '/v1/resource'")
			})
		})
	})

	Convey("Given a resource factory with an empty OpenAPI resource", t, func() {
		r := resourceFactory{}
		Convey("When delete is called with empty data and a empty client", func() {
			client := &clientOpenAPIStub{}
			err := r.delete(nil, client)
			Convey("Then the error should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a resource factory that has an asynchronous create operation (delete) but the polling operation fails for some reason", t, func() {
		expectedReturnCode := 202
		testSchema := newTestSchema(idProperty, stringProperty)
		resourceData := testSchema.getResourceData(t)
		specResource := newSpecStubResourceWithOperations("resourceName", "/v1/resource", false, testSchema.getSchemaDefinition(), &specResourceOperation{}, &specResourceOperation{}, &specResourceOperation{}, &specResourceOperation{responses: specResponses{expectedReturnCode: &specResponse{isPollingEnabled: true}}})
		r := resourceFactory{
			openAPIResource: specResource,
			defaultTimeout:  time.Duration(0 * time.Second),
		}
		Convey("When create is called with resource data and a client", func() {
			client := &clientOpenAPIStub{
				returnHTTPCode: expectedReturnCode,
				responsePayload: map[string]interface{}{
					idProperty.Name:     "someID",
					stringProperty.Name: "someExtraValueThatProvesResponseDataIsPersisted",
				},
			}
			err := r.delete(resourceData, client)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "polling mechanism failed after DELETE /v1/resource call with response status code (202): error waiting for resource to reach a completion status ([destroyed]) [valid pending statuses ([])]: error on retrieving resource 'resourceName' () when waiting: [resource='resourceName'] HTTP Response Status Code 202 not matching expected one [200] ()")
			})
		})
	})
}

func TestImporter(t *testing.T) {
	Convey("Given a resource factory configured with a root resource (and the already populated id property value provided by the user)", t, func() {
		importedIDProperty := idProperty
		r, resourceData := testCreateResourceFactoryWithID(t, importedIDProperty, stringProperty)
		Convey("When importer is called", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someOtherStringValue",
				},
			}
			resourceImporter := r.importer()
			Convey("Then the resource importer returned should Not be nil", func() {
				So(resourceImporter, ShouldNotBeNil)
			})
			Convey("And when the resourceImporter State method is invoked with data resource and the provider client", func() {
				data, err := resourceImporter.State(resourceData, client)
				Convey("Then the err returned should be nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("And the data list returned should have one item", func() {
					So(len(data), ShouldEqual, 1)
				})
				Convey("And the data returned should contained the expected resource ID", func() {
					So(data[0].Id(), ShouldEqual, idProperty.Default)
				})
				Convey("And the data returned should contained the imported id field with the right value", func() {
					So(data[0].Get(importedIDProperty.Name), ShouldEqual, importedIDProperty.Default)
				})
				Convey("And the data returned should contained the imported string field with the right value returned from the API", func() {
					So(data[0].Get(stringProperty.Name), ShouldEqual, client.responsePayload[stringProperty.Name])
				})
			})
		})
	})
	Convey("Given a resource factory configured with a sub-resource (and the already populated id property value provided by the user with the correct format)", t, func() {
		expectedParentID := "32"
		expectedResourceInstanceID := "159"
		expectedParentPropertyName := "cdns_v1_id"

		importedIDValue := fmt.Sprintf("%s/%s", expectedParentID, expectedResourceInstanceID)
		importedIDProperty := newStringSchemaDefinitionProperty("id", "", true, true, false, false, false, true, false, false, importedIDValue)
		expectedParentProperty := newStringSchemaDefinitionProperty(expectedParentPropertyName, "", true, true, false, false, false, true, false, false, "")
		r, resourceData := testCreateSubResourceFactory(t, "/v1/cdns/{id}/firewall", []string{"cdns_v1"}, []string{expectedParentPropertyName}, "cdns_v1", importedIDProperty, stringProperty, expectedParentProperty)

		Convey("When importer is called", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someOtherStringValue",
				},
			}
			resourceImporter := r.importer()
			Convey("Then the resource importer returned should Not be nil", func() {
				So(resourceImporter, ShouldNotBeNil)
			})
			Convey("And when the resourceImporter State method is invoked with the provider client and resource data for one item", func() {
				data, err := resourceImporter.State(resourceData, client)
				Convey("Then the err returned should be nil", func() {
					So(err, ShouldBeNil)
				})
				Convey("And the data list returned should have one item", func() {
					So(len(data), ShouldEqual, 1)
				})
				Convey("And the data returned should contain the parent id field with the right value", func() {
					So(data[0].Get(expectedParentPropertyName), ShouldEqual, expectedParentID)
				})
				Convey("And the data returned should contain the expected resource ID", func() {
					So(data[0].Id(), ShouldEqual, expectedResourceInstanceID)
				})
				Convey("And the data returned should contain the imported string field with the right value returned from the API", func() {
					So(data[0].Get(stringProperty.Name), ShouldEqual, client.responsePayload[stringProperty.Name])
				})
			})
		})
	})

	Convey("Given a resource factory configured with a sub-resource (and the already populated id property value provided by the user with incorrect format)", t, func() {
		expectedParentPropertyName := "cdns_v1_id"

		importedIDValue := "someStringThatDoesNotMatchTheExpectedSubResourceIDFormat"
		importedIDProperty := newStringSchemaDefinitionProperty("id", "", true, true, false, false, false, true, false, false, importedIDValue)
		expectedParentProperty := newStringSchemaDefinitionProperty(expectedParentPropertyName, "", true, true, false, false, false, true, false, false, "")
		r, resourceData := testCreateSubResourceFactory(t, "/v1/cdns/{id}/firewall", []string{"cdns_v1"}, []string{expectedParentPropertyName}, "cdns_v1", importedIDProperty, stringProperty, expectedParentProperty)

		Convey("When importer is called", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someOtherStringValue",
				},
			}
			resourceImporter := r.importer()
			Convey("Then the resource importer returned should Not be nil", func() {
				So(resourceImporter, ShouldNotBeNil)
			})
			Convey("And when the resourceImporter State method is invoked with data resource and the provider client", func() {
				_, err := resourceImporter.State(resourceData, client)
				Convey("Then the err returned should be the expected one", func() {
					So(err.Error(), ShouldEqual, "can not import a subresource without providing all the parent IDs (1) and the instance ID")
				})
			})
		})
	})

	Convey("Given a resource factory configured with a sub-resource (and the already populated id property value contains more IDs than the resource number of parent properties)", t, func() {
		expectedParentPropertyName := "cdns_v1_id"
		importedIDValue := "/extraID/1234/23564"
		importedIDProperty := newStringSchemaDefinitionProperty("id", "", true, true, false, false, false, true, false, false, importedIDValue)
		expectedParentProperty := newStringSchemaDefinitionProperty(expectedParentPropertyName, "", true, true, false, false, false, true, false, false, "")
		r, resourceData := testCreateSubResourceFactory(t, "/v1/cdns/{id}/firewall", []string{"cdns_v1"}, []string{expectedParentPropertyName}, "cdns_v1", importedIDProperty, stringProperty, expectedParentProperty)

		Convey("When importer is called", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someOtherStringValue",
				},
			}
			resourceImporter := r.importer()
			Convey("Then the resource importer returned should Not be nil", func() {
				So(resourceImporter, ShouldNotBeNil)
			})
			Convey("And when the resourceImporter State method is invoked with data resource and the provider client", func() {
				_, err := resourceImporter.State(resourceData, client)
				Convey("Then the err returned should be the expected one", func() {
					So(err.Error(), ShouldEqual, "the number of parent IDs provided 3 is greater than the expected number of parent IDs 1")
				})
			})
		})
	})

	Convey("Given a resource factory configured with a sub-resource (and the already populated id property value contains less IDs than the resource number of parent properties)", t, func() {
		importedIDValue := "1234/5647" // missing one of the parent ids
		importedIDProperty := newStringSchemaDefinitionProperty("id", "", true, true, false, false, false, true, false, false, importedIDValue)
		expectedParentProperty := newStringSchemaDefinitionProperty("cdns_v1_firewalls_v1", "", true, true, false, false, false, true, false, false, "")
		r, resourceData := testCreateSubResourceFactory(t, "/v1/cdns/{id}/firewalls/{id}/rules", []string{"cdns_v1", "cdns_v1_firewalls_v1"}, []string{"cdns_v1_id", "cdns_v1_firewalls_v1_id"}, "cdns_v1_firewalls_v1", importedIDProperty, stringProperty, expectedParentProperty)

		Convey("When importer is called", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someOtherStringValue",
				},
			}
			resourceImporter := r.importer()
			Convey("Then the resource importer returned should Not be nil", func() {
				So(resourceImporter, ShouldNotBeNil)
			})
			Convey("And when the resourceImporter State method is invoked with data resource and the provider client", func() {
				_, err := resourceImporter.State(resourceData, client)
				Convey("Then the err returned should be the expected one", func() {
					So(err.Error(), ShouldEqual, "can not import a subresource without all the parent ids, expected 2 and got 1 parent IDs")
				})
			})
		})
	})

	Convey("Given a resource factory configured with a sub-resource (and the already populated id property value contains the right IDs but the resource for some reason is not configured correctly and does not have configured the parent id property)", t, func() {
		expectedParentPropertyName := "cdns_v1_id"
		importedIDValue := "1234/5678"
		importedIDProperty := newStringSchemaDefinitionProperty("id", "", true, true, false, false, false, true, false, false, importedIDValue)
		r, resourceData := testCreateSubResourceFactory(t, "/v1/cdns/{id}/firewall", []string{"cdns_v1"}, []string{expectedParentPropertyName}, "cdns_v1", importedIDProperty, stringProperty)

		Convey("When importer is called", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					stringProperty.Name: "someOtherStringValue",
				},
			}
			resourceImporter := r.importer()
			Convey("Then the resource importer returned should Not be nil", func() {
				So(resourceImporter, ShouldNotBeNil)
			})
			Convey("And when the resourceImporter State method is invoked with data resource and the provider client", func() {
				_, err := resourceImporter.State(resourceData, client)
				Convey("Then the err returned should be the expected one", func() {
					So(err.Error(), ShouldEqual, "could not find ID value in the state file for subresource parent property 'cdns_v1_id'")
				})
			})
		})
	})
}

func TestHandlePollingIfConfigured(t *testing.T) {
	Convey("Given a resource factory configured with a resource which has a schema definition containing a status property", t, func() {
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty, stringProperty, statusProperty)
		Convey("When handlePollingIfConfigured is called with an operation that has a response defined for the API response status code passed in and polling is enabled AND the API returns a status that matches the target", func() {
			targetState := "deployed"
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     idProperty.Default,
					stringProperty.Name: stringProperty.Default,
					statusProperty.Name: targetState,
				},
				returnHTTPCode: http.StatusOK,
			}

			responsePayload := map[string]interface{}{}

			responseStatusCode := http.StatusAccepted
			operation := &specResourceOperation{
				responses: map[int]*specResponse{
					responseStatusCode: {
						isPollingEnabled:    true,
						pollPendingStatuses: []string{"pending"},
						pollTargetStatuses:  []string{targetState},
					},
				},
			}
			err := r.handlePollingIfConfigured(&responsePayload, resourceData, client, operation, responseStatusCode, schema.TimeoutCreate)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the remote data should be the payload returned by the API", func() {
				So(responsePayload[idProperty.Name], ShouldEqual, client.responsePayload[idProperty.Name])
				So(responsePayload[stringProperty.Name], ShouldEqual, client.responsePayload[stringProperty.Name])
				So(responsePayload[statusProperty.Name], ShouldEqual, client.responsePayload[statusProperty.Name])
			})
		})

		Convey("When handlePollingIfConfigured is called with an operation that has a response defined for the API response status code passed in and polling is enabled AND the responsePayload is nil (meaning we are handling a DELETE operation)", func() {
			targetState := "deployed"
			client := &clientOpenAPIStub{
				returnHTTPCode: http.StatusNotFound,
			}
			responsePayload := map[string]interface{}{}

			responseStatusCode := http.StatusAccepted
			operation := &specResourceOperation{
				responses: map[int]*specResponse{
					responseStatusCode: {
						isPollingEnabled:    true,
						pollPendingStatuses: []string{"pending"},
						pollTargetStatuses:  []string{targetState},
					},
				},
			}
			err := r.handlePollingIfConfigured(nil, resourceData, client, operation, responseStatusCode, schema.TimeoutCreate)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the remote data should be the payload returned by the API", func() {
				So(responsePayload[idProperty.Name], ShouldEqual, client.responsePayload[idProperty.Name])
				So(responsePayload[stringProperty.Name], ShouldEqual, client.responsePayload[stringProperty.Name])
				So(responsePayload[statusProperty.Name], ShouldEqual, client.responsePayload[statusProperty.Name])
			})
		})

		Convey("When handlePollingIfConfigured is called with a response status code that DOES NOT any of the operation's reponse definitions", func() {
			client := &clientOpenAPIStub{}
			responseStatusCode := http.StatusAccepted
			operation := &specResourceOperation{
				responses: map[int]*specResponse{},
			}
			err := r.handlePollingIfConfigured(nil, resourceData, client, operation, responseStatusCode, schema.TimeoutCreate)
			Convey("Then the err  should be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When handlePollingIfConfigured is called with a response status code that DOES math one of the operation responses BUT polling is not enabled for that response", func() {
			client := &clientOpenAPIStub{}
			responseStatusCode := http.StatusAccepted
			operation := &specResourceOperation{
				responses: map[int]*specResponse{
					responseStatusCode: {
						isPollingEnabled:    false,
						pollPendingStatuses: []string{"pending"},
						pollTargetStatuses:  []string{"deployed"},
					},
				},
			}
			err := r.handlePollingIfConfigured(nil, resourceData, client, operation, responseStatusCode, schema.TimeoutCreate)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a resource factory that has an asynchronous create operation (post) but the polling operation fails for some reason", t, func() {
		expectedReturnCode := http.StatusAccepted
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty, stringProperty, statusProperty)
		Convey("When create is called with resource data and a client", func() {
			operation := &specResourceOperation{
				responses: map[int]*specResponse{
					expectedReturnCode: {
						isPollingEnabled:    true,
						pollPendingStatuses: []string{"pending"},
						pollTargetStatuses:  []string{"deployed"},
					},
				},
			}
			client := &clientOpenAPIStub{
				returnHTTPCode: expectedReturnCode,
				responsePayload: map[string]interface{}{
					idProperty.Name:     "someID",
					stringProperty.Name: "someExtraValueThatProvesResponseDataIsPersisted",
				},
				error: fmt.Errorf("some error"),
			}
			err := r.handlePollingIfConfigured(nil, resourceData, client, operation, expectedReturnCode, schema.TimeoutCreate)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "error waiting for resource to reach a completion status ([destroyed]) [valid pending statuses ([pending])]: error on retrieving resource 'resourceName' (id) when waiting: some error")
			})
		})
	})

}

func TestResourceStateRefreshFunc(t *testing.T) {
	Convey("Given a resource factory configured with a resource which has a schema definition containing a status property", t, func() {
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty, stringProperty, statusProperty)
		Convey("When resourceStateRefreshFunc is called with an update resource data and an open api client and the returned function (stateRefreshFunc) is invoked", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     idProperty.Default,
					stringProperty.Name: stringProperty.Default,
					statusProperty.Name: statusProperty.Default,
				},
			}
			stateRefreshFunc := r.resourceStateRefreshFunc(resourceData, client)
			remoteData, newStatus, err := stateRefreshFunc()
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the new status should match the one returned by the API", func() {
				So(newStatus, ShouldEqual, client.responsePayload[statusProperty.Name])
			})
			Convey("And the remote data should be the payload returned by the API", func() {
				So(remoteData.(map[string]interface{})[idProperty.Name], ShouldEqual, client.responsePayload[idProperty.Name])
				So(remoteData.(map[string]interface{})[stringProperty.Name], ShouldEqual, client.responsePayload[stringProperty.Name])
				So(remoteData.(map[string]interface{})[statusProperty.Name], ShouldEqual, client.responsePayload[statusProperty.Name])
			})

		})
		Convey("When resourceStateRefreshFunc is called with an update resource data and an open api client that returns 404 not found", func() {
			client := &clientOpenAPIStub{
				returnHTTPCode: http.StatusNotFound,
			}
			stateRefreshFunc := r.resourceStateRefreshFunc(resourceData, client)
			_, newStatus, err := stateRefreshFunc()
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the new status should be the internal hardcoded status 'destroyed' as a response with 404 status code is not expected to have a body", func() {
				So(newStatus, ShouldEqual, defaultDestroyStatus)
			})
		})

		Convey("When resourceStateRefreshFunc is called with an update resource data and an open api client that returns an error", func() {
			expectedError := "some error"
			client := &clientOpenAPIStub{
				error: errors.New(expectedError),
			}
			stateRefreshFunc := r.resourceStateRefreshFunc(resourceData, client)
			remoteData, newStatus, err := stateRefreshFunc()
			Convey("Then the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be the expected one", func() {
				So(err.Error(), ShouldEqual, fmt.Sprintf("error on retrieving resource 'resourceName' (id) when waiting: %s", expectedError))
			})
			Convey("And the remoteData should be empty", func() {
				So(remoteData, ShouldBeNil)
			})
			Convey("And the new status should be empty", func() {
				So(newStatus, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resource factory configured with a resource which has a schema definition missing a status property", t, func() {
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty, stringProperty)
		Convey("When resourceStateRefreshFunc is called with an update resource data and an open api client that returns an error", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     idProperty.Default,
					stringProperty.Name: stringProperty.Default,
				},
			}
			stateRefreshFunc := r.resourceStateRefreshFunc(resourceData, client)
			remoteData, newStatus, err := stateRefreshFunc()
			Convey("Then the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be the expected one", func() {
				So(err.Error(), ShouldEqual, "error occurred while retrieving status identifier value from payload for resource 'resourceName' (id): could not find any status property. Please make sure the resource schema definition has either one property named 'status' or one property is marked with IsStatusIdentifier set to true")
			})
			Convey("And the remoteData should be empty", func() {
				So(remoteData, ShouldBeNil)
			})
			Convey("And the new status should be empty", func() {
				So(newStatus, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resource factory configured with a resource which has a schema definition with a status property but the responsePayload is missing the status property", t, func() {
		r, resourceData := testCreateResourceFactoryWithID(t, idProperty, stringProperty, statusProperty)
		Convey("When resourceStateRefreshFunc is called with an update resource data and an open api client that returns an error", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{
					idProperty.Name:     idProperty.Default,
					stringProperty.Name: stringProperty.Default,
				},
			}
			stateRefreshFunc := r.resourceStateRefreshFunc(resourceData, client)
			remoteData, newStatus, err := stateRefreshFunc()
			Convey("Then the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be the expected one", func() {
				So(err.Error(), ShouldEqual, "error occurred while retrieving status identifier value from payload for resource 'resourceName' (id): payload does not match resouce schema, could not find the status field: [status]")
			})
			Convey("And the remoteData should be empty", func() {
				So(remoteData, ShouldBeNil)
			})
			Convey("And the new status should be empty", func() {
				So(newStatus, ShouldBeEmpty)
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

func TestCreatePayloadFromLocalStateData(t *testing.T) {

	Convey("Given a resource factory initialized with a spec resource with schema definitions for each of the supported property types (string, int, number, bool, slice of primitive, slice of objects, object and object with nested objects and a parent property)", t, func() {

		// - Object property configuration
		// object_property {
		//	 origin_port = 80
		//	 protocol = "http"
		// }
		objectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		objectDefault := map[string]interface{}{
			"origin_port": objectSchemaDefinition.Properties[0].Default,
			"protocol":    objectSchemaDefinition.Properties[1].Default,
		}
		objectProperty := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, false, objectDefault, objectSchemaDefinition)

		// - Object property with nested objects configuration
		// property_with_nested_object {
		//	id = "id",
		//	nested_object {
		//		origin_port = 80
		//		protocol = "http"
		//	}
		//}
		propertyWithNestedObjectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				idProperty,
				objectProperty,
			},
		}
		// Tag(NestedStructsWorkaround)
		// Note: This is the workaround needed to support properties with nested structs. The current Terraform sdk version
		// does not support this now, hence the suggestion from the Terraform maintainer was to use a list of map[string]interface{}
		// with the list containing just one element. The below represents the internal representation of the terraform state
		// for an object property that contains other objects
		propertyWithNestedObjectDefault := []map[string]interface{}{
			{
				"id":              propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
				"object_property": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
			},
		}
		propertyWithNestedObject := newObjectSchemaDefinitionPropertyWithDefaults("property_with_nested_object", "", true, false, false, propertyWithNestedObjectDefault, propertyWithNestedObjectSchemaDefinition)

		// - Array of objects property configuration
		// slice_object_property [
		//   {
		//	   origin_port = 80
		//     protocol = "http"
		//   }
		// ]
		arrayObjectDefault := []map[string]interface{}{
			{
				"origin_port": 80,
				"protocol":    "http",
			},
		}
		sliceObjectProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, arrayObjectDefault, typeObject, objectSchemaDefinition)

		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("parentProperty", "", true, false, "http")
		parentProperty.IsParentProperty = true

		r, resourceData := testCreateResourceFactory(t, idProperty, computedProperty, stringProperty, intProperty, numberProperty, boolProperty, slicePrimitiveProperty, sliceObjectProperty, objectProperty, propertyWithNestedObject, parentProperty)

		Convey("When createPayloadFromLocalStateData is called with a terraform resource data", func() {
			payload := r.createPayloadFromLocalStateData(resourceData)
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should not include the following keys as they are either an identifier or read only (computed) properties", func() {
				So(payload, ShouldNotContainKey, idProperty.Name)
				So(payload, ShouldNotContainKey, computedProperty.Name)
				So(payload, ShouldNotContainKey, parentProperty.Name)
			})
			Convey("And then payload returned should include the following keys ", func() {
				So(payload, ShouldContainKey, stringProperty.Name)
				So(payload, ShouldContainKey, intProperty.Name)
				So(payload, ShouldContainKey, numberProperty.Name)
				So(payload, ShouldContainKey, boolProperty.Name)
				So(payload, ShouldContainKey, slicePrimitiveProperty.Name)
				So(payload, ShouldContainKey, sliceObjectProperty.Name)
				So(payload, ShouldContainKey, objectProperty.Name)
				So(payload, ShouldContainKey, propertyWithNestedObject.Name)
			})
			Convey("And then payload returned should contain the expected data values for the string property", func() {
				So(payload[stringProperty.Name], ShouldEqual, stringProperty.Default)
			})
			Convey("And then payload returned should contain the expected data values for the int property", func() {
				So(payload[intProperty.Name], ShouldEqual, intProperty.Default)
			})
			Convey("And then payload returned should contain the expected data values for the number property", func() {
				So(payload[numberProperty.Name], ShouldEqual, numberProperty.Default)
			})
			Convey("And then payload returned should contain the expected data values for the bool property", func() {
				So(payload[boolProperty.Name], ShouldEqual, boolProperty.Default)
			})
			Convey("And then payload returned should contain the expected data values for the slice of primitive property", func() {
				So(payload[slicePrimitiveProperty.Name], ShouldContain, slicePrimitiveProperty.Default.([]string)[0])
			})
			Convey("And then payload returned should contain the expected data values for the slice of objects property", func() {
				arrayObject := payload[sliceObjectProperty.Name].([]interface{})
				object := arrayObject[0].(map[string]interface{})
				So(object["origin_port"], ShouldEqual, arrayObjectDefault[0]["origin_port"])
				So(object["protocol"], ShouldEqual, arrayObjectDefault[0]["protocol"])
			})
			Convey("And then payload returned should cotnain the expected data values for the object properties", func() {
				object := payload[objectProperty.Name].(map[string]interface{})
				So(object[objectProperty.SpecSchemaDefinition.Properties[0].Name], ShouldEqual, objectProperty.SpecSchemaDefinition.Properties[0].Default.(int))
				So(object[objectProperty.SpecSchemaDefinition.Properties[1].Name], ShouldEqual, objectProperty.SpecSchemaDefinition.Properties[1].Default)
			})
			Convey("And then payload returned should contain the expected data values for the object property with nested object", func() {
				topLevel := payload[propertyWithNestedObject.Name].(map[string]interface{})
				So(topLevel, ShouldContainKey, objectProperty.Name)
				So(topLevel, ShouldContainKey, idProperty.Name)
				So(topLevel[idProperty.Name], ShouldEqual, propertyWithNestedObjectSchemaDefinition.Properties[0].Default)
				nestedLevel := topLevel[objectProperty.Name].(map[string]interface{})
				So(nestedLevel["origin_port"], ShouldEqual, propertyWithNestedObjectSchemaDefinition.Properties[1].Default.(map[string]interface{})["origin_port"])
				So(nestedLevel["protocol"], ShouldEqual, propertyWithNestedObjectSchemaDefinition.Properties[1].Default.(map[string]interface{})["protocol"])
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

func TestGetPropertyPayload(t *testing.T) {
	Convey("Given a resource factory"+
		"When getPropertyPayload is called with a nil property"+
		"Then it panics", t, func() {
		input := map[string]interface{}{}
		dataValue := struct{}{}
		resourceFactory := resourceFactory{}
		So(func() { resourceFactory.getPropertyPayload(input, nil, dataValue) }, ShouldPanic)
	})

	Convey("Given a resource factory"+
		"When getPropertyPayload is called with a nil datavalue"+
		"Then it returns an error", t, func() {
		input := map[string]interface{}{}
		resourceFactory := resourceFactory{}
		So(resourceFactory.getPropertyPayload(input, &specSchemaDefinitionProperty{Name: "buu"}, nil).Error(), ShouldEqual, `property 'buu' has a nil state dataValue`)
	})

	Convey("Given a resource factory"+
		"When it is called with  non-nil property and value for dataValue which cannot be cast to []interface{}"+
		"Then it panics", t, func() {
		input := map[string]interface{}{}
		dataValue := []bool{}
		property := &specSchemaDefinitionProperty{}
		resourceFactory := resourceFactory{}
		So(func() { resourceFactory.getPropertyPayload(input, property, dataValue) }, ShouldPanic)
	})

	Convey("Given the function handleSliceOrArray"+
		"When it is called with an empty slice for dataValue"+
		"Then it should not return an error", t, func() {
		input := map[string]interface{}{}
		dataValue := []interface{}{}
		property := &specSchemaDefinitionProperty{}
		resourceFactory := resourceFactory{}
		e := resourceFactory.getPropertyPayload(input, property, dataValue)
		So(e, ShouldBeNil)
	})

	Convey("Given a resource factory initialized with a spec resource with a schema definition containing a string property", t, func() {
		// Use case - string property (terraform configuration pseudo representation below):
		// string_property = "some value"
		r, resourceData := testCreateResourceFactory(t, stringProperty)
		Convey("When getPropertyPayload is called with an empty map, the string property in the resource schema and it's corresponding terraform resourceData state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(stringProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, stringProperty, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should have the string property", func() {
				So(payload, ShouldContainKey, stringProperty.Name)
			})
			Convey("And then payload returned should have the data value from the state file", func() {
				So(payload[stringProperty.Name], ShouldEqual, stringProperty.Default)
			})
		})
	})

	Convey("Given a resource factory initialized with a schema definition containing an int property", t, func() {
		// Use case - int property (terraform configuration pseudo representation below):
		// int_property = 1234
		r, resourceData := testCreateResourceFactory(t, intProperty)
		Convey("When getPropertyPayload is called with an empty map, the int property in the resource schema  and it's corresponding terraform resourceData state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(intProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, intProperty, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should have the integer property", func() {
				So(payload, ShouldContainKey, intProperty.Name)
			})
			Convey("And then payload returned should have the data value from the state file", func() {
				So(payload[intProperty.Name], ShouldEqual, intProperty.Default)
			})
		})
	})

	Convey("Given a resource factory initialized with a schema definition containing a number property", t, func() {
		// Use case - number property (terraform configuration pseudo representation below):
		// number_property = 1.1234
		r, resourceData := testCreateResourceFactory(t, numberProperty)
		Convey("When getPropertyPayload is called with an empty map, the number property in the resource schema and it's corresponding terraform resourceData state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(numberProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, numberProperty, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should have the number property", func() {
				So(payload, ShouldContainKey, numberProperty.Name)
			})
			Convey("And then payload returned should have the data value from the state file", func() {
				So(payload[numberProperty.Name], ShouldEqual, numberProperty.Default)
			})
		})
	})

	Convey("Given a resource factory initialized with a schema definition containing a bool property", t, func() {
		// Use case - bool property (terraform configuration pseudo representation below):
		// bool_property = true
		r, resourceData := testCreateResourceFactory(t, boolProperty)
		Convey("When getPropertyPayload is called with an empty map, the bool property in the resource schema and it's corresponding terraform resourceData state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(boolProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, boolProperty, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should have the bool property", func() {
				So(payload, ShouldContainKey, boolProperty.Name)
			})
			Convey("And then payload returned should have the data value from the state file", func() {
				So(payload[boolProperty.Name], ShouldEqual, boolProperty.Default)
			})
		})
	})

	Convey("Given a resource factory initialized with a schema definition containing an object property", t, func() {
		// Use case - object property (terraform configuration pseudo representation below):
		// object_property {
		//	 origin_port = 80
		//	 protocol = "http"
		// }
		objectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		objectDefault := map[string]interface{}{
			"origin_port": 80,
			"protocol":    "http",
		}
		objectProperty := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, false, objectDefault, objectSchemaDefinition)
		r, resourceData := testCreateResourceFactory(t, objectProperty)
		Convey("When getPropertyPayload is called with an empty map, the object property in the resource schema and it's state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(objectProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, objectProperty, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should have the object property", func() {
				So(payload, ShouldContainKey, objectProperty.Name)
			})
			Convey("And then payload returned should have the data value from the state file", func() {
				So(payload[objectProperty.Name].(map[string]interface{})[objectProperty.SpecSchemaDefinition.Properties[0].Name], ShouldEqual, objectProperty.SpecSchemaDefinition.Properties[0].Default.(int))
				So(payload[objectProperty.Name].(map[string]interface{})[objectProperty.SpecSchemaDefinition.Properties[1].Name], ShouldEqual, objectProperty.SpecSchemaDefinition.Properties[1].Default)
			})
		})
	})

	Convey("Given a resource factory initialized with a schema definition containing an array ob objects property", t, func() {
		// Use case - object property (terraform configuration pseudo representation below):
		// slice_object_property [
		//{
		//	 origin_port = 80
		//	 protocol = "http"
		// }
		// ]
		objectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, 80),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		objectDefault := map[string]interface{}{
			"origin_port": 80,
			"protocol":    "http",
		}
		arrayObjectDefault := []map[string]interface{}{
			objectDefault,
		}
		sliceObjectProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, false, arrayObjectDefault, typeObject, objectSchemaDefinition)
		r, resourceData := testCreateResourceFactory(t, sliceObjectProperty)
		Convey("When getPropertyPayload is called with an empty map, the array of objects property in the resource schema and it's state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(sliceObjectProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, sliceObjectProperty, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should have the object property", func() {
				So(payload, ShouldContainKey, sliceObjectProperty.Name)
			})
			Convey("And then payload returned should have the data value from the state file properly formatter with the right types", func() {
				// For some reason the data values in the terraform state file are all strings
				So(payload[sliceObjectProperty.Name].([]interface{})[0].(map[string]interface{})[sliceObjectProperty.SpecSchemaDefinition.Properties[0].Name], ShouldEqual, sliceObjectProperty.SpecSchemaDefinition.Properties[0].Default.(int))
				So(payload[sliceObjectProperty.Name].([]interface{})[0].(map[string]interface{})[sliceObjectProperty.SpecSchemaDefinition.Properties[1].Name], ShouldEqual, sliceObjectProperty.SpecSchemaDefinition.Properties[1].Default)
			})
		})
	})

	Convey("Given a resource factory initialized with a schema definition containing a slice of strings property", t, func() {
		// Use case - slice of srings (terraform configuration pseudo representation below):
		// slice_property = ["some_value"]
		r, resourceData := testCreateResourceFactory(t, slicePrimitiveProperty)
		Convey("When getPropertyPayload is called with an empty map, the slice of strings property in the resource schema and it's state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(slicePrimitiveProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, slicePrimitiveProperty, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("And then payload returned should have the object property", func() {
				So(payload, ShouldContainKey, slicePrimitiveProperty.Name)
			})
			Convey("And then payload returned should have the data value from the state file", func() {
				So(payload[slicePrimitiveProperty.Name].([]interface{})[0], ShouldEqual, slicePrimitiveProperty.Default.([]string)[0])
			})
		})
	})

	Convey("Given a resource factory initialized with a spec resource with a property schema definition containing one nested struct", t, func() {
		// Use case - object with nested objects (terraform configuration pseudo representation below):
		// property_with_nested_object {
		//	id = "id",
		//	nested_object {
		//		origin_port = 80
		//		protocol = "http"
		//	}
		//}
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
		// Tag(NestedStructsWorkaround)
		// Note: This is the workaround needed to support properties with nested structs. The current Terraform sdk version
		// does not support this now, hence the suggestion from the Terraform maintainer was to use a list of map[string]interface{}
		// with the list containing just one element. The below represents the internal representation of the terraform state
		// for an object property that contains other objects
		propertyWithNestedObjectDefault := []map[string]interface{}{
			{
				"id":            propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
				"nested_object": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
			},
		}
		expectedPropertyWithNestedObjectName := "property_with_nested_object"
		propertyWithNestedObject := newObjectSchemaDefinitionPropertyWithDefaults(expectedPropertyWithNestedObjectName, "", true, false, false, propertyWithNestedObjectDefault, propertyWithNestedObjectSchemaDefinition)
		r, resourceData := testCreateResourceFactory(t, propertyWithNestedObject)
		Convey("When getPropertyPayload is called a slice with >1 dataValue, it complains", func() {
			err := r.getPropertyPayload(map[string]interface{}{}, propertyWithNestedObject, []interface{}{"foo", "bar", "baz"})
			So(err.Error(), ShouldEqual, "something is really wrong here...an object property with nested objects should have exactly one elem in the terraform state list")

		})
		Convey("When getPropertyPayload is called a slice with <1 dataValue, it complains", func() {
			err := r.getPropertyPayload(map[string]interface{}{}, propertyWithNestedObject, []interface{}{})
			So(err.Error(), ShouldEqual, "something is really wrong here...an object property with nested objects should have exactly one elem in the terraform state list")

		})

		Convey("When getPropertyPayload is called with an empty map, the property with nested object in the resource schema and it's corresponding terraform resourceData state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(propertyWithNestedObject.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, propertyWithNestedObject, dataValue)
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the map returned should not be empty", func() {
				So(payload, ShouldNotBeEmpty)
			})
			Convey("Then the map returned should contain the 'property_with_nested_object' property", func() {
				So(payload, ShouldContainKey, expectedPropertyWithNestedObjectName)
			})
			topLevel := payload[expectedPropertyWithNestedObjectName].(map[string]interface{})
			Convey("And then payload returned should have the id property with the right value", func() {
				So(topLevel, ShouldContainKey, idProperty.Name)
				So(topLevel[idProperty.Name], ShouldEqual, propertyWithNestedObjectSchemaDefinition.Properties[0].Default)
			})
			Convey("And then payload returned should have the nested_object object property with the right value", func() {
				So(topLevel, ShouldContainKey, nestedObject.Name)
				nestedLevel := topLevel[nestedObject.Name].(map[string]interface{})
				So(nestedLevel["origin_port"], ShouldEqual, propertyWithNestedObjectSchemaDefinition.Properties[1].Default.(map[string]interface{})["origin_port"])
				So(nestedLevel["protocol"], ShouldEqual, propertyWithNestedObjectSchemaDefinition.Properties[1].Default.(map[string]interface{})["protocol"])
			})
		})
	})

	Convey("Given a resource factory initialized with a spec resource with a property schema definition containing one nested struct but having terraform property names that are not valid within the resource definition", t, func() {
		// Use case -  crappy path while getting properties paylod for properties which do not exists in nested objects
		nestedObjectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, "http"),
			},
		}
		nestedObjectNoTFCompliantName := map[string]interface{}{
			"badprotocoldoesntexist": nestedObjectSchemaDefinition.Properties[0].Default,
		}
		nestedObjectNotTFCompliant := newObjectSchemaDefinitionPropertyWithDefaults("nested_object_not_compliant", "", true, false, false, nestedObjectNoTFCompliantName, nestedObjectSchemaDefinition)
		propertyWithNestedObjectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				idProperty,
				nestedObjectNotTFCompliant,
			},
		}
		propertyWithNestedObjectDefault := []map[string]interface{}{
			{
				"id":                          propertyWithNestedObjectSchemaDefinition.Properties[0].Default,
				"nested_object_not_compliant": propertyWithNestedObjectSchemaDefinition.Properties[1].Default,
			},
		}
		expectedPropertyWithNestedObjectName := "property_with_nested_object"
		propertyWithNestedObject := newObjectSchemaDefinitionPropertyWithDefaults(expectedPropertyWithNestedObjectName, "", true, false, false, propertyWithNestedObjectDefault, propertyWithNestedObjectSchemaDefinition)
		r, resourceData := testCreateResourceFactory(t, propertyWithNestedObject)
		Convey("When getPropertyPayload is called with an empty map, the property with nested object in the resource schema and it's corresponding terraform resourceData state data value", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(propertyWithNestedObject.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, propertyWithNestedObject, dataValue)
			Convey("Then the error should not be nil", func() {
				So(err.Error(), ShouldEqual, "property with terraform name 'badprotocoldoesntexist' not existing in resource schema definition")
			})
			Convey("Then the map returned should be empty", func() {
				So(payload, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resource factory initialized with a spec resource with a property schema definition containing a slice of objects with an invalid slice name definition", t, func() {
		// Use case -  crappy path while getting properties paylod for properties which do not exists in slices
		objectSchemaDefinition := &specSchemaDefinition{}
		arrayObjectDefault := []map[string]interface{}{
			{"protocol": "http"},
		}
		sliceObjectProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property_doesn_not_exists", "", true, false, false, arrayObjectDefault, typeObject, objectSchemaDefinition)

		r, resourceData := testCreateResourceFactory(t, sliceObjectProperty)

		Convey("When getPropertyPayload is called with an empty map, the property slice of objects in the resource schema are not found", func() {
			payload := map[string]interface{}{}
			dataValue, _ := resourceData.GetOkExists(sliceObjectProperty.getTerraformCompliantPropertyName())
			err := r.getPropertyPayload(payload, sliceObjectProperty, dataValue)
			Convey("Then the error should not be nil", func() {
				So(err.Error(), ShouldEqual, "property 'slice_object_property_doesn_not_exists' has a nil state dataValue")
			})
			Convey("Then the map returned should be empty", func() {
				So(payload, ShouldBeEmpty)
			})
		})
	})
}

func TestGetStatusValueFromPayload(t *testing.T) {
	Convey("Given a swagger schema definition that has an status property that is not an object", t, func() {
		specResource := newSpecStubResource(
			"resourceName",
			"/v1/resource",
			false,
			&specSchemaDefinition{
				Properties: specSchemaDefinitionProperties{
					&specSchemaDefinitionProperty{
						Name:     statusDefaultPropertyName,
						Type:     typeString,
						ReadOnly: true,
					},
				},
			})
		r := resourceFactory{
			openAPIResource: specResource,
		}
		Convey("When getStatusValueFromPayload method is called with a payload that also has a 'status' field in the root level", func() {
			expectedStatusValue := "someValue"
			payload := map[string]interface{}{
				statusDefaultPropertyName: expectedStatusValue,
			}
			statusField, err := r.getStatusValueFromPayload(payload)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should contain the name of the property 'status'", func() {
				So(statusField, ShouldEqual, expectedStatusValue)
			})
		})

		Convey("When getStatusValueFromPayload method is called with a payload that does not have status field", func() {
			payload := map[string]interface{}{
				"someOtherPropertyName": "arggg",
			}
			_, err := r.getStatusValueFromPayload(payload)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message should be", func() {
				So(err.Error(), ShouldEqual, "payload does not match resouce schema, could not find the status field: [status]")
			})
		})

		Convey("When getStatusValueFromPayload method is called with a payload that has a status field but the value is not supported", func() {
			payload := map[string]interface{}{
				statusDefaultPropertyName: 12, // this value is not supported, only strings and maps (for nested properties within an object) are supported
			}
			_, err := r.getStatusValueFromPayload(payload)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the error message should be", func() {
				So(err.Error(), ShouldEqual, "status property value '[status]' does not have a supported type [string/map]")
			})
		})
	})

	Convey("Given a swagger schema definition that has an status property that IS an object", t, func() {
		expectedStatusProperty := "some-other-property-holding-status"
		specResource := newSpecStubResource(
			"resourceName",
			"/v1/resource",
			false,
			&specSchemaDefinition{
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
			})
		r := resourceFactory{
			openAPIResource: specResource,
		}
		Convey("When getStatusValueFromPayload method is called with a payload that has an status object property inside which there's an status property", func() {
			expectedStatusValue := "someStatusValue"
			payload := map[string]interface{}{
				statusDefaultPropertyName: map[string]interface{}{
					expectedStatusProperty: expectedStatusValue,
				},
			}
			statusField, err := r.getStatusValueFromPayload(payload)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should contain the name of the property 'status'", func() {
				So(statusField, ShouldEqual, expectedStatusValue)
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
func testCreateResourceFactoryWithID(t *testing.T, idSchemaDefinitionProperty *specSchemaDefinitionProperty, schemaDefinitionProperties ...*specSchemaDefinitionProperty) (resourceFactory, *schema.ResourceData) {
	schemaDefinitionProperties = append(schemaDefinitionProperties, idSchemaDefinitionProperty)
	resourceFactory, resourceData := testCreateResourceFactory(t, schemaDefinitionProperties...)
	resourceData.SetId(idSchemaDefinitionProperty.Default.(string))
	return resourceFactory, resourceData
}

// testCreateResourceFactory configures the resourceData with some properties.
func testCreateResourceFactory(t *testing.T, schemaDefinitionProperties ...*specSchemaDefinitionProperty) (resourceFactory, *schema.ResourceData) {
	testSchema := newTestSchema(schemaDefinitionProperties...)
	resourceData := testSchema.getResourceData(t)
	specResource := newSpecStubResourceWithOperations("resourceName", "/v1/resource", false, testSchema.getSchemaDefinition(), &specResourceOperation{}, &specResourceOperation{}, &specResourceOperation{}, &specResourceOperation{})
	return newResourceFactory(specResource), resourceData
}

func testCreateSubResourceFactory(t *testing.T, path string, parentResourceNames, parentPropertyNames []string, fullParentResourceName string, idSchemaDefinitionProperty *specSchemaDefinitionProperty, schemaDefinitionProperties ...*specSchemaDefinitionProperty) (resourceFactory, *schema.ResourceData) {
	testSchema := newTestSchema(schemaDefinitionProperties...)
	resourceData := testSchema.getResourceData(t)
	resourceData.SetId(idSchemaDefinitionProperty.Default.(string))
	specResource := newSpecStubResourceWithOperations("subResourceName", path, false, testSchema.getSchemaDefinition(), &specResourceOperation{}, &specResourceOperation{}, &specResourceOperation{}, &specResourceOperation{})
	specResource.parentResourceNames = parentResourceNames
	specResource.parentPropertyNames = parentPropertyNames
	specResource.fullParentResourceName = fullParentResourceName
	return newResourceFactory(specResource), resourceData
}
