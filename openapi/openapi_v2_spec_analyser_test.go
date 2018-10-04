package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestIsMultiRegionResource(t *testing.T) {
	Convey("Given a specV2Analyser and a resource root has a POST operation containing the x-terraform-resource-host with a parametrized host containing region variable", t, func() {
		serviceProviderName := "serviceProviderName"
		r := specV2Analyser{}
		resourceRoot := &spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfResourceURL: fmt.Sprintf("some.api.${%s}.domain.com", serviceProviderName),
						},
					},
				},
			},
		}
		Convey("When isMultiRegionResource method is called with a resourceRoot pathItem and a set of extensions where one matches the region for which the above 's-terraform-resource-host' extension is for", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, serviceProviderName), "uswest,useast")
			isMultiRegion, regions, err := r.isMultiRegionResource(resourceRoot, rootLevelExtensions)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("And the map returned should contain uswest", func() {
				So(regions, ShouldContainKey, "uswest")
				So(regions["uswest"], ShouldEqual, "some.api.uswest.domain.com")
			})
			Convey("And the map returned should contain useast", func() {
				So(regions, ShouldContainKey, "useast")
				So(regions["useast"], ShouldEqual, "some.api.useast.domain.com")
			})
		})

		Convey("When isMultiRegionResource method is called with a set of extensions where NONE matches the region for which the above 's-terraform-resource-host' extension is for", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, "someOtherServiceProvider"), "rst, dub")
			isMultiRegion, regions, err := r.isMultiRegionResource(resourceRoot, rootLevelExtensions)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the err returned should contain the following error message", func() {
				So(err.Error(), ShouldEqual, "missing matching 'serviceProviderName' root level region extension 'x-terraform-resource-regions-serviceProviderName'")
			})
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
			Convey("And the regions map returned should be empty", func() {
				So(regions, ShouldBeEmpty)
			})
		})

		Convey("When isMultiRegionResource method is called with a set of extensions where one matches the region for which the above 's-terraform-resource-host' extension is for BUT the values are not comma separated", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, serviceProviderName), "uswest useast")
			isMultiRegion, regions, err := r.isMultiRegionResource(resourceRoot, rootLevelExtensions)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("And the map returned should contain uswest", func() {
				So(regions, ShouldContainKey, "uswestuseast")
				So(regions["uswestuseast"], ShouldEqual, "some.api.uswestuseast.domain.com")
			})
		})

		Convey("When isMultiRegionResource method is called with a set of extensions where one matches the region for which the above 's-terraform-resource-host' extension is for BUT the values are comma separated with spaces", func() {
			rootLevelExtensions := spec.Extensions{}
			rootLevelExtensions.Add(fmt.Sprintf(extTfResourceRegionsFmt, serviceProviderName), "uswest, useast")
			isMultiRegion, regions, err := r.isMultiRegionResource(resourceRoot, rootLevelExtensions)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be true", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("And the map returned should contain uswest", func() {
				So(regions, ShouldContainKey, "uswest")
				So(regions["uswest"], ShouldEqual, "some.api.uswest.domain.com")
			})
			Convey("And the map returned should contain useast", func() {
				So(regions, ShouldContainKey, "useast")
				So(regions["useast"], ShouldEqual, "some.api.useast.domain.com")
			})
		})
	})
}

func TestResourceInstanceRegex(t *testing.T) {
	Convey("Given an specV2Analyser", t, func() {
		a := specV2Analyser{}
		Convey("When resourceInstanceRegex method is called", func() {
			regex, err := a.resourceInstanceRegex()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the regex should not be nil", func() {
				So(regex, ShouldNotBeNil)
			})
		})
	})
}

func TestResourceInstanceEndPoint(t *testing.T) {
	Convey("Given an specV2Analyser", t, func() {
		a := specV2Analyser{}
		Convey("When isResourceInstanceEndPoint method is called with a valid resource path such as '/resource/{id}'", func() {
			resourceInstance, err := a.isResourceInstanceEndPoint("/resource/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(resourceInstance, ShouldBeTrue)
			})
		})
		Convey("When isResourceInstanceEndPoint method is called with a long path such as '/very/long/path/{id}'", func() {
			resourceInstance, err := a.isResourceInstanceEndPoint("/very/long/path/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(resourceInstance, ShouldBeTrue)
			})
		})
		Convey("When isResourceInstanceEndPoint method is called with a path that has path parameters '/resource/{name}/subresource/{id}'", func() {
			resourceInstance, err := a.isResourceInstanceEndPoint("/resource/{name}/subresource/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(resourceInstance, ShouldBeTrue)
			})
		})
		Convey("When isResourceInstanceEndPoint method is called with an invalid resource path such as '/resource/not/instance/path' not conforming with the expected pattern '/resource/{id}'", func() {
			resourceInstance, err := a.isResourceInstanceEndPoint("/resource/not/valid/instance/path")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be false", func() {
				So(resourceInstance, ShouldBeFalse)
			})
		})
	})
}

func TestGetPayloadDefName(t *testing.T) {
	Convey("Given an specV2Analyser", t, func() {
		a := specV2Analyser{}

		// Local Reference use cases
		Convey("When getPayloadDefName method is called with a valid internal definition path", func() {
			defName, err := a.getPayloadDefName("#/definitions/ContentDeliveryNetworkV1")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(defName, ShouldEqual, "ContentDeliveryNetworkV1")
			})
		})

		Convey("When getPayloadDefName method is called with a URL (not supported)", func() {
			_, err := a.getPayloadDefName("http://path/to/your/resource.json#myElement")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		// Remote Reference use cases
		Convey("When getPayloadDefName method is called with an element of the document located on the same server (not supported)", func() {
			_, err := a.getPayloadDefName("document.json#/myElement")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When getPayloadDefName method is called with an element of the document located in the parent folder (not supported)", func() {
			_, err := a.getPayloadDefName("../document.json#/myElement")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When getPayloadDefName method is called with an specific element of the document stored on the different server (not supported)", func() {
			_, err := a.getPayloadDefName("http://path/to/your/resource.json#myElement")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		// URL Reference use case
		Convey("When getPayloadDefName method is called with an element of the document located in another folder (not supported)", func() {
			_, err := a.getPayloadDefName("../another-folder/document.json#/myElement")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When getPayloadDefName method is called with document on the different server, which uses the same protocol (not supported)", func() {
			_, err := a.getPayloadDefName("//anotherserver.com/files/example.json")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestGetResourcePayloadSchemaRef(t *testing.T) {
	Convey("Given an specV2Analyser", t, func() {
		a := specV2Analyser{}

		Convey("When getResourcePayloadSchemaRef method is called with an operation that has a reference to the resource schema definition", func() {
			expectedRef := "#/definitions/ContentDeliveryNetworkV1"
			ref := spec.MustCreateRef(expectedRef)
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{
						{
							ParamProps: spec.ParamProps{
								In:   "body",
								Name: "body",
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref,
									},
								},
							},
						},
					},
				},
			}
			returnedRef, err := a.getResourcePayloadSchemaRef(operation)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(returnedRef, ShouldEqual, expectedRef)
			})
		})

		Convey("When getResourcePayloadSchemaRef method is called with an operation that does not have parameters", func() {
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{},
				},
			}
			_, err := a.getResourcePayloadSchemaRef(operation)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "operation does not have parameters defined")
			})
		})
		Convey("When getResourcePayloadSchemaRef method is called with an operation that is missing required body parameter ", func() {
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{
						{
							ParamProps: spec.ParamProps{
								In:   "header",
								Name: "some-header",
							},
						},
					},
				},
			}
			_, err := a.getResourcePayloadSchemaRef(operation)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "operation is missing required 'body' type parameter")
			})
		})
		Convey("When getResourcePayloadSchemaRef method is called with an operation that has multiple body parameters", func() {
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{
						{
							ParamProps: spec.ParamProps{
								In:   "body",
								Name: "first body",
							},
						},
						{
							ParamProps: spec.ParamProps{
								In:   "body",
								Name: "second body",
							},
						},
					},
				},
			}
			_, err := a.getResourcePayloadSchemaRef(operation)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "operation contains multiple 'body' parameters")
			})
		})
		Convey("When getResourcePayloadSchemaRef method is called with an operation that is missing a the schema field", func() {
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{
						{
							ParamProps: spec.ParamProps{
								In:   "body",
								Name: "body",
								//Schema: No schema
							},
						},
					},
				},
			}
			_, err := a.getResourcePayloadSchemaRef(operation)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "operation is missing the ref to the schema definition")
			})
		})
		Convey("When getResourcePayloadSchemaRef method is called with an operation that has a schema but the ref is not populated", func() {
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{
						{
							ParamProps: spec.ParamProps{
								In:     "body",
								Name:   "body",
								Schema: &spec.Schema{},
							},
						},
					},
				},
			}
			_, err := a.getResourcePayloadSchemaRef(operation)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "operation has an invalid schema definition ref empty")
			})
		})
	})
}

func TestGetResourcePayloadSchemaDef(t *testing.T) {
	Convey("Given an specV2Analyser", t, func() {
		swaggerContent := `swagger: "2.0"
definitions:
  Users:
    type: "object"
    required:
      - name
    properties:
      id:
        type: "string"
        readOnly: true`

		a := initAPISpecAnalyser(swaggerContent)
		Convey("When getResourcePayloadSchemaDef method is called with an operation containing a valid ref: '#/definitions/Users'", func() {
			expectedRef := "#/definitions/Users"
			ref := spec.MustCreateRef(expectedRef)
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{
						{
							ParamProps: spec.ParamProps{
								In:   "body",
								Name: "body",
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref,
									},
								},
							},
						},
					},
				},
			}
			resourcePayloadSchemaDef, err := a.getResourcePayloadSchemaDef(operation)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be a valid schema def", func() {
				So(len(resourcePayloadSchemaDef.Type), ShouldEqual, 1)
				So(resourcePayloadSchemaDef.Type, ShouldContain, "object")
			})
		})

	})

	Convey("Given an specV2Analyser", t, func() {
		a := specV2Analyser{}
		Convey("When getResourcePayloadSchemaDef method is called with an operation that is missing the parameters section", func() {
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{},
			}
			_, err := a.getResourcePayloadSchemaDef(operation)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "operation does not have parameters defined")
			})
		})
	})

	Convey("Given an apiSpecAnalyser", t, func() {
		swaggerContent := `swagger: "2.0"
definitions:
  OtherDef:
    type: "object"
    required:
      - name
    properties:
      id:
        type: "string"
        readOnly: true`

		a := initAPISpecAnalyser(swaggerContent)
		Convey("When getResourcePayloadSchemaDef method is called with an operation that is missing the definition the ref is pointing at", func() {
			ref := spec.MustCreateRef("#/definitions/Users")
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Parameters: []spec.Parameter{
						{
							ParamProps: spec.ParamProps{
								In:   "body",
								Name: "body",
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Ref: ref,
									},
								},
							},
						},
					},
				},
			}
			_, err := a.getResourcePayloadSchemaDef(operation)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "missing schema definition in the swagger file with the supplied ref '#/definitions/Users'")
			})
		})
	})
}

func TestFindMatchingResourceRootPath(t *testing.T) {

	Convey("Given an apiSpecAnalyser with a valid resource path such as '/users/{id}' and missing resource root path", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When findMatchingResourceRootPath method is called ", func() {
			_, err := a.findMatchingResourceRootPath("/users/{id}")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource instance path '/users/{id}' missing resource root path")
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a valid resource path such as '/users/{id}' and root path with trailing slash", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users/:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When findMatchingResourceRootPath method is called ", func() {
			resourceRootPath, err := a.findMatchingResourceRootPath("/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be '/users/'", func() {
				So(resourceRootPath, ShouldEqual, "/users/")
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a valid resource path such as '/users/{id}' and root path with without slash", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When findMatchingResourceRootPath method is called ", func() {
			resourceRootPath, err := a.findMatchingResourceRootPath("/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be '/users'", func() {
				So(resourceRootPath, ShouldEqual, "/users")
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a valid resource path that is versioned such as '/v1/users/{id}' and root path containing version", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /v1/users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"
  /v1/users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When findMatchingResourceRootPath method is called ", func() {
			resourceRootPath, err := a.findMatchingResourceRootPath("/v1/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be '/v1/users'", func() {
				So(resourceRootPath, ShouldEqual, "/v1/users")
			})
		})
	})
}

func TestPostIsPresent(t *testing.T) {

	Convey("Given an specV2Analyser with a path '/users' that has a post operation", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When postDefined method is called'", func() {
			postIsPresent := a.postDefined("/users")
			Convey("Then the value returned should be true", func() {
				So(postIsPresent, ShouldBeTrue)
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a path '/users' that DOES NOT have a post operation", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When postDefined method is called'", func() {
			postIsPresent := a.postDefined("/users")
			Convey("Then the value returned should be false", func() {
				So(postIsPresent, ShouldBeFalse)
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a path '/users'", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When postDefined method is called wigh a non existing path'", func() {
			postIsPresent := a.postDefined("/nonExistingPath")
			Convey("Then the value returned should be false", func() {
				So(postIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestValidateResourceSchemaDefinition(t *testing.T) {
	Convey("Given an specV2Analyser", t, func() {
		a := specV2Analyser{}
		Convey("When validateResourceSchemaDefinition method is called with a valid schema definition containing a property ID'", func() {
			schema := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"id": spec.Schema{},
					},
				},
			}
			err := a.validateResourceSchemaDefinition(schema)
			Convey("Then error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("When validateResourceSchemaDefinition method is called with a valid schema definition missing an ID property but a different property acts as unique identifier'", func() {
			schema := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"name": spec.Schema{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfID: true,
								},
							},
						},
					},
				},
			}
			err := a.validateResourceSchemaDefinition(schema)
			Convey("Then error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("When validateResourceSchemaDefinition method is called with a valid schema definition with both a property that name 'id' and a different property with the 'x-terraform-id' extension'", func() {
			schema := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"id": spec.Schema{},
						"name": spec.Schema{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfID: true,
								},
							},
						},
					},
				},
			}
			err := a.validateResourceSchemaDefinition(schema)
			Convey("Then error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("When validateResourceSchemaDefinition method is called with a NON valid schema definition due to missing unique identifier'", func() {
			schema := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"name": spec.Schema{},
					},
				},
			}
			err := a.validateResourceSchemaDefinition(schema)
			Convey("Then error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource schema is missing a property that uniquely identifies the resource, either a property named 'id' or a property with the extension 'x-terraform-id' set to true")
			})
		})
	})
}

func TestValidateRootPath(t *testing.T) {
	Convey("Given an specV2Analyser with a terraform compliant root path", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/nonExisting"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When validateResourceSchemaDefinition method is called with '/users/{id}'", func() {
			_, _, _, err := a.validateRootPath("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root path '/users' POST operation validation error: missing schema definition in the swagger file with the supplied ref '#/definitions/nonExisting'")
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a resource instance path such as '/users/{id}' that its root path '/users' DOES NOT expose a POST operation", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When validateResourceSchemaDefinition method is called with '/users/{id}'", func() {
			_, _, _, err := a.validateRootPath("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root path '/users' missing required POST operation")
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a resource instance path such as '/users/{id}' that its root path '/users' is missing the reference to the schea definition", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"
definitions:
  Users:
    type: "object"
    required:
      - name
    properties:
      id:
        type: "string"
        readOnly: true
      name:
        type: "string"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When validateResourceSchemaDefinition method is called with '/users/{id}'", func() {
			resourceRootPath, _, _, err := a.validateRootPath("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(resourceRootPath, ShouldContainSubstring, "/users")
			})
		})
	})
}

func TestValidateInstancePath(t *testing.T) {
	Convey("Given an specV2Analyser with a terraform compliant instance path", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The user id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When validateInstancePath method is called with '/users/{id}'", func() {
			err := a.validateInstancePath("/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an specV2Analyser with an instance path that is missing the get operation", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users/{id}:
    put:
      parameters:
      - name: "id"
        in: "path"
        description: "The user id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When validateInstancePath method is called with '/users/{id}'", func() {
			err := a.validateInstancePath("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource instance path '/users/{id}' missing required GET operation")
			})
		})
	})

	Convey("Given an specV2Analyser", t, func() {
		a := specV2Analyser{}
		Convey("When validateInstancePath method is called with a non instance path", func() {
			err := a.validateInstancePath("/users")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "path '/users' is not a resource instance path")
			})
		})
	})
}

func TestIsEndPointTerraformResourceCompliant(t *testing.T) {
	Convey("Given an specV2Analyser with a fully terraform compliant resource Users", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"
definitions:
  Users:
    type: "object"
    required:
      - name
    properties:
      id:
        type: "string"
        readOnly: true
      name:
        type: "string"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When isEndPointFullyTerraformResourceCompliant method is called ", func() {
			resourceRootPath, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be", func() {
				So(resourceRootPath, ShouldEqual, "/users")
			})
		})
	})

	// This is the ideal case where the resource exposes al CRUD operations
	Convey("Given an specV2Analyser with an resource instance path such as '/users/{id}' that has a GET/PUT/DELETE operations exposed and the corresponding resource root path '/users' exposes a POST operation", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/Users"
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"
    put:
      parameters:
      - name: "id"
        in: "path"
        type: "string"
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/Users"
    delete:
      parameters:
      - name: "id"
        in: "path"
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"
definitions:
  Users:
    type: "object"
    required:
      - name
    properties:
      id:
        type: "string"
        readOnly: true
      name:
        type: "string"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When isEndPointFullyTerraformResourceCompliant method is called ", func() {
			resourceRootPath, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be", func() {
				So(resourceRootPath, ShouldEqual, "/users")
			})
		})
	})

	// This use case avoids resource duplicates as the root paths are filtered out
	Convey("Given an specV2Analyser", t, func() {
		swaggerContent := `swagger: "2.0"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When isEndPointFullyTerraformResourceCompliant method is called with a non resource instance path such as '/users'", func() {
			_, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "path '/users' is not a resource instance path")
			})
		})
	})

	Convey("Given an specV2Analyser with a resource that fails the instance path validation (no get operation defined)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users/{id}:
    put:
    delete:`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When isEndPointFullyTerraformResourceCompliant method is called ", func() {
			_, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource instance path '/users/{id}' missing required GET operation")
			})
		})
	})

	Convey("Given an specV2Analyser with a resource that fails the root path validation (no post operation defined)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When isEndPointFullyTerraformResourceCompliant method is called ", func() {
			_, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root path '/users' missing required POST operation")
			})
		})
	})

	Convey("Given an specV2Analyser with a resource that fails the schema validation (non existing ref)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/Users"
      responses:
        201:
          schema:
            $ref: "#/definitions/NonExistingDefinition"
  /users/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/Users"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When isEndPointFullyTerraformResourceCompliant method is called ", func() {
			_, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root path '/users' POST operation validation error: missing schema definition in the swagger file with the supplied ref '#/definitions/Users'")
			})
		})
	})
}

func TestGetTerraformCompliantResources(t *testing.T) {
	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource /v1/cdns and some non compliant paths", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /v1/cdns:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/ContentDeliveryNetwork"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetwork"
  /v1/cdns/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/ContentDeliveryNetwork"
    put:
      parameters:
      - name: "id"
        in: "path"
        type: "string"
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/ContentDeliveryNetwork"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetwork"
    delete:
      parameters:
      - name: "id"
        in: "path"
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"
  /non/compliant:
    post: # this path post operation is missing a reference to the schema definition (commented out)
      parameters:
      - in: "body"
        name: "body"
      #  schema:
      #    $ref: "#/definitions/NonCompliant"
      responses:
        201:
          schema:
            $ref: "#/definitions/NonCompliant"
  /non/compliant/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/NonCompliant"
definitions:
  ContentDeliveryNetwork:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
  NonCompliant:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When GetTerraformCompliantResources method is called ", func() {
			terraformCompliantResources, err := a.GetTerraformCompliantResources()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resources info map should only contain a resource called cdns_v1", func() {
				So(len(terraformCompliantResources), ShouldEqual, 1)
				So(terraformCompliantResources[0].getResourceName(), ShouldEqual, "cdns_v1")
			})
		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a non compliant terraform resource /v1/cdns because its missing the post operation", t, func() {
		var swaggerJSON = `
{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns/{id}":{
         "get":{
            "summary":"Get cdn by id"
         },
         "put":{
            "summary":"Updated cdn"
         },
         "delete":{
            "summary":"Delete cdn"
         }
      }
   },
   "definitions":{
      "ContentDeliveryNetwork":{
         "type":"object",
         "properties":{
            "id":{
               "type":"string"
            }
         }
      }
   }
}`
		a := initAPISpecAnalyser(swaggerJSON)
		Convey("When GetTerraformCompliantResources method is called ", func() {
			terraformCompliantResources, err := a.GetTerraformCompliantResources()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resources info map should contain a resource called cdns_v1", func() {
				So(terraformCompliantResources, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource that has the 'x-terraform-exclude-resource' with value true", t, func() {
		var swaggerJSON = `
{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
            "x-terraform-exclude-resource": true,
            "summary":"Create cdn",
            "parameters":[
               {
                  "in":"body",
                  "name":"body",
                  "description":"Created CDN",
                  "schema":{
                     "$ref":"#/definitions/ContentDeliveryNetwork"
                  }
               }
            ]
         }
      },
      "/v1/cdns/{id}":{
         "get":{
            "summary":"Get cdn by id"
         },
         "put":{
            "summary":"Updated cdn"
         },
         "delete":{
            "summary":"Delete cdn"
         }
      }
   },
   "definitions":{
      "ContentDeliveryNetwork":{
         "type":"object",
         "properties":{
            "id":{
               "type":"string"
            }
         }
      }
   }
}`
		a := initAPISpecAnalyser(swaggerJSON)
		Convey("When GetTerraformCompliantResources method is called ", func() {
			terraformCompliantResources, err := a.GetTerraformCompliantResources()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the terraformCompliantResources map should contain one resource with ignore flag set to true", func() {
				So(terraformCompliantResources[0].shouldIgnoreResource(), ShouldBeTrue)
			})
		})
	})
}

func initAPISpecAnalyser(swaggerContent string) specV2Analyser {
	swagger := json.RawMessage([]byte(swaggerContent))
	d, _ := loads.Analyzed(swagger, "2.0")
	return specV2Analyser{d: d}
}
