package openapi

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

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
				So(err.Error(), ShouldEqual, "[resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] ()")
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

		Convey("When update is called with resource data and a client returns a non expected http code when reading remote", func() {
			client := &clientOpenAPIStub{
				responsePayload: map[string]interface{}{},
				returnHTTPCode:  http.StatusInternalServerError,
			}
			err := r.update(resourceData, client)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should be", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] ()")
			})
		})
	})

	Convey("Given a resource factory with no delete operation configured", t, func() {
		specResource := newSpecStubResource("resourceName", "/v1/resource", false, nil)
		r := newResourceFactory(specResource)
		Convey("When update is called with resource data and a client", func() {
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
}

func TestImporter(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		r, resourceData := testCreateResourceFactory(t, idProperty, stringProperty)
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
				Convey("And the data returned should contained the imported id field with the right value", func() {
					So(data[0].Get(idProperty.Name), ShouldEqual, idProperty.Default)
				})
				Convey("And the data returned should contained the imported string field with the right value returned from the API", func() {
					So(data[0].Get(stringProperty.Name), ShouldEqual, client.responsePayload[stringProperty.Name])
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
			Convey("Then the err returned should be nil", func() {
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
				So(err.Error(), ShouldEqual, fmt.Sprintf("error on retrieving resource '/v1/resource' (id) when waiting: %s", expectedError))
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
				So(err.Error(), ShouldEqual, "error occurred while retrieving status identifier value from payload for resource '/v1/resource' (id): could not find any status property. Please make sure the resource schema definition has either one property named 'status' or one property is marked with IsStatusIdentifier set to true")
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
				So(err.Error(), ShouldEqual, "error occurred while retrieving status identifier value from payload for resource '/v1/resource' (id): payload does not match resouce schema, could not find the status field: [status]")
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
		r, resourceData := testCreateResourceFactory(t, idProperty, computedProperty, stringProperty, intProperty, numberProperty, boolProperty, slicePrimitiveProperty)
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
				So(payload, ShouldContainKey, slicePrimitiveProperty.Name)
			})
			Convey("And then payload key values should match the values stored in the terraform resource data", func() {
				So(payload[stringProperty.Name], ShouldEqual, stringProperty.Default)
				So(payload[intProperty.Name], ShouldEqual, intProperty.Default)
				So(payload[numberProperty.Name], ShouldEqual, numberProperty.Default)
				So(payload[boolProperty.Name], ShouldEqual, boolProperty.Default)
				So(payload[slicePrimitiveProperty.Name], ShouldContain, slicePrimitiveProperty.Default.([]string)[0])
			})
		})
	})

	Convey("Given a resource factory initialized with a spec resource with some schema definition containing an array of objects", t, func() {
		objectSchemaDefinition := &specSchemaDefinition{
			Properties: specSchemaDefinitionProperties{
				newIntSchemaDefinitionPropertyWithDefaults("origin_port", "", true, false, nil),
				newStringSchemaDefinitionPropertyWithDefaults("protocol", "", true, false, nil),
			},
		}
		arrayObjectDefault := []map[string]interface{}{
			{
				"origin_port": 80,
				"protocol":    "http",
			},
		}
		sliceObjectProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, arrayObjectDefault, typeObject, objectSchemaDefinition)
		r, resourceData := testCreateResourceFactory(t, idProperty, computedProperty, stringProperty, slicePrimitiveProperty, sliceObjectProperty)
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
				So(payload, ShouldContainKey, slicePrimitiveProperty.Name)
				So(payload, ShouldContainKey, sliceObjectProperty.Name)
			})
			Convey("And then payload key values should match the values stored in the terraform resource data", func() {
				So(payload[stringProperty.Name], ShouldEqual, stringProperty.Default)
				So(payload[slicePrimitiveProperty.Name], ShouldContain, slicePrimitiveProperty.Default.([]string)[0])
				So(payload[sliceObjectProperty.Name].([]interface{})[0].(map[string]interface{})["origin_port"], ShouldEqual, arrayObjectDefault[0]["origin_port"])
				So(payload[sliceObjectProperty.Name].([]interface{})[0].(map[string]interface{})["protocol"], ShouldEqual, arrayObjectDefault[0]["protocol"])
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
	Convey("Given a resource factory initialized with a spec resource with some schema definition", t, func() {
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
		objectProperty := newObjectSchemaDefinitionPropertyWithDefaults("object_property", "", true, false, objectDefault, objectSchemaDefinition)
		sliceObjectProperty := newListSchemaDefinitionPropertyWithDefaults("slice_object_property", "", true, false, arrayObjectDefault, typeObject, objectSchemaDefinition)
		r, resourceData := testCreateResourceFactory(t, idProperty, computedProperty, stringProperty, intProperty, numberProperty, boolProperty, slicePrimitiveProperty, objectProperty, sliceObjectProperty)
		Convey("When createPayloadFromLocalStateData is called with an empty map, the string property in the resource schema and it's state data value", func() {
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
		Convey("When createPayloadFromLocalStateData is called with an empty map, the int property in the resource schema and it's state data value", func() {
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
		Convey("When createPayloadFromLocalStateData is called with an empty map, the number property in the resource schema and it's state data value", func() {
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
		Convey("When createPayloadFromLocalStateData is called with an empty map, the bool property in the resource schema and it's state data value", func() {
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

		Convey("When createPayloadFromLocalStateData is called with an empty map, the object property in the resource schema and it's state data value", func() {
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

		Convey("When createPayloadFromLocalStateData is called with an empty map, the array of objects property in the resource schema and it's state data value", func() {
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

		Convey("When createPayloadFromLocalStateData is called with an empty map, the slice of strings property in the resource schema and it's state data value", func() {
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
				So(err.Error(), ShouldEqual, "could not find schema definition property name nonExistingKey in the resource data: property with name 'nonExistingKey' not existing in resource schema definition")
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
