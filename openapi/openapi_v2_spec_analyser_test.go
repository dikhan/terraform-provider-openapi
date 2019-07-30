package openapi

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSpecV2Analyser(t *testing.T) {
	Convey("Given a openAPIDocumentURL and a swagger doc object", t, func() {
		openAPIDocumentURL := ""
		d := &loads.Document{}
		Convey("When specV2Analyser method is constructed", func() {
			specV2Analyser := &specV2Analyser{
				openAPIDocumentURL: openAPIDocumentURL,
				d:                  d,
			}
			Convey("Then the specV2Analyser should comply with SpecAnalyser interface", func() {
				var _ SpecAnalyser = specV2Analyser
			})
		})
	})
}

func Test_pathExists(t *testing.T) {
	Convey("Given a blank specV2Analyser", t, func() {
		a := &specV2Analyser{}
		Convey("When pathExists is called", func() {
			Convey("Then it panics", func() {
				So(func() { a.pathExists("whatever") }, ShouldPanic)
			})
		})
	})

	Convey("Given a specV2Analyser with a blank d", t, func() {
		a := &specV2Analyser{d: &loads.Document{}}
		Convey("When pathExists is called", func() {
			Convey("Then it panics", func() {
				So(func() { a.pathExists("whatever") }, ShouldPanic)
			})
		})
	})

	Convey("Given a specV2Analyser initialized from a swagger doc with no paths", t, func() {
		swaggerDoc := `swagger: "2.0"`
		a := initAPISpecAnalyser(swaggerDoc)
		Convey("When pathExists is called", func() {
			Convey("Then it panics", func() {
				So(func() { a.pathExists("whatever") }, ShouldPanic)
			})
		})
	})

	Convey("Given a specV2Analyser initialized from a swagger doc with a path with a trailing slash", t, func() {
		swaggerDoc := `swagger: "2.0"
paths:
 /users/{id}/:
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

		a := initAPISpecAnalyser(swaggerDoc)
		Convey("When pathExists is called with a path without the trailing slash", func() {
			b, i := a.pathExists("/users/{id}")
			Convey("Then it returns true and the PathItem Operation is not nil", func() {
				So(b, ShouldBeTrue)
				So(i.Get, ShouldNotBeNil)
			})
		})
		Convey("When pathExists is called with a path with the trailing slash", func() {
			b, i := a.pathExists("/users/{id}/")
			Convey("Then it returns true and the PathItem Operation is not nil", func() {
				So(b, ShouldBeTrue)
				So(i.Get, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a specV2Analyser initialized from a swagger doc with a path without a trailing slash", t, func() {
		swaggerDoc := `swagger: "2.0"
paths:
 /abusers/{id}:
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

		a := initAPISpecAnalyser(swaggerDoc)
		Convey("When pathExists is called with a path not listed", func() {
			b, i := a.pathExists("whatever")
			Convey("Then it should return false and a non-nil PathItem", func() {
				So(b, ShouldBeFalse)
				So(i, ShouldNotBeNil)
			})
		})
		Convey("When pathExists is called with a path with no trailing slash that is listed", func() {
			b, i := a.pathExists("/abusers/{id}")
			Convey("Then it returns true and the PathItem Operation is not nil", func() {
				So(b, ShouldBeTrue)
				So(i.Get, ShouldNotBeNil)
			})
		})
		Convey("When pathExists is called with a path that is listed but with a trailing slash", func() {
			b, i := a.pathExists("/abusers/{id}/")
			Convey("Then it returns false and the PathItem is not nil", func() {
				So(b, ShouldBeFalse)
				So(i, ShouldNotBeNil)
			})
		})
	})

}

func Test_getBodyParameterBodySchema(t *testing.T) {
	Convey("Given a specV2Analyser", t, func() {
		specV2Analyser := &specV2Analyser{}
		Convey("When getBodyParameterBodySchema is called with an Operation with OperationProps with a Parameter with an In:body ParamProp and a Schema ParamProp with some properties", func() {
			resourceRootPostOperation := &spec.Operation{}
			schema := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"id": {},
					},
				},
			}
			param := spec.Parameter{ParamProps: spec.ParamProps{In: "body", Schema: schema}}
			resourceRootPostOperation.Parameters = []spec.Parameter{param}
			schema, err := specV2Analyser.getBodyParameterBodySchema(resourceRootPostOperation)
			Convey("Then the schema returned should not be empty", func() {
				So(schema, ShouldNotBeNil)
			})
			Convey("And the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("When getBodyParameterBodySchema is called with a nil arg", func() {
			_, err := specV2Analyser.getBodyParameterBodySchema(nil)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "resource root operation does not have a POST operation")
			})
		})
		Convey("When getBodyParameterBodySchema is called with an empty Operation (no params)", func() {
			resourceRootPostOperation := &spec.Operation{}
			_, err := specV2Analyser.getBodyParameterBodySchema(resourceRootPostOperation)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "resource root operation missing the body parameter")
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
			_, err := specV2Analyser.getBodyParameterBodySchema(operation)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "operation contains multiple 'body' parameters")
			})
		})
		Convey("When getBodyParameterBodySchema is called with an Operation with OperationProps with a Parameter with an In:body ParamProp and NO Schema ParamProp", func() {
			resourceRootPostOperation := &spec.Operation{}
			param := spec.Parameter{ParamProps: spec.ParamProps{In: "body"}}
			resourceRootPostOperation.Parameters = []spec.Parameter{param}
			_, err := specV2Analyser.getBodyParameterBodySchema(resourceRootPostOperation)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "resource root operation missing the schema for the POST operation body parameter")
			})
		})
		Convey("When getBodyParameterBodySchema is called with an Operation with OperationProps with a Parameter with an In:body ParamProp and and a schema with a ref not expanded", func() {
			resourceRootPostOperation := &spec.Operation{}
			ref := spec.MustCreateRef("#/definitions/Users")
			s := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Ref: spec.Ref(ref),
				},
			}
			param := spec.Parameter{ParamProps: spec.ParamProps{In: "body", Schema: s}}
			resourceRootPostOperation.Parameters = []spec.Parameter{param}
			schema, err := specV2Analyser.getBodyParameterBodySchema(resourceRootPostOperation)
			Convey("Then the schema returned should be empty", func() {
				So(schema, ShouldBeNil)
			})
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "the operation ref was not expanded properly, check that the ref is valid (no cycles, bogus, etc)")
			})
		})
		Convey("When getBodyParameterBodySchema is called with an Operation with OperationProps with a Parameter with an In:body ParamProp and a Schema ParamProp with NO properties", func() {
			resourceRootPostOperation := &spec.Operation{}
			schema := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{},
				},
			}
			param := spec.Parameter{ParamProps: spec.ParamProps{In: "body", Schema: schema}}
			resourceRootPostOperation.Parameters = []spec.Parameter{param}
			schema, err := specV2Analyser.getBodyParameterBodySchema(resourceRootPostOperation)
			Convey("Then the schema returned should be empty", func() {
				So(schema, ShouldBeNil)
			})
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "POST operation contains an schema with no properties")
			})
		})
	})
}

func TestNewSpecAnalyserV2(t *testing.T) {

	Convey("Given a valid swagger doc where a definition has a ref to an external definition hosted somewhere else (in this case file system)", t, func() {
		externalRefFile := initAPISpecFile(createExternalSwaggerContent())
		defer os.Remove(externalRefFile.Name())

		var swaggerJSON = createSwaggerWithExternalRef(externalRefFile.Name())

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())
		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specAnalyserV2 struct should not be nil", func() {
				So(specAnalyserV2, ShouldNotBeNil)
			})
			Convey("And the new doc should contain the definition ref expanded with the right required fields", func() {
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Required[0], ShouldEqual, "name")
			})
			Convey("And the new doc should contain the definition ref expanded with the right required properties", func() {
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Properties, ShouldContainKey, "id")
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Properties, ShouldContainKey, "name")

			})
			Convey("And the ref should be empty", func() {
				ref := specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Ref.Ref
				So(ref.GetURL(), ShouldBeNil)
			})
		})
	})

	Convey("Given a valid swagger doc where a definition has a ref to an external definition hosted somewhere else (in this case an HTTP server)", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, createExternalSwaggerContent())
		}))
		defer ts.Close()

		var swaggerJSON = createSwaggerWithExternalRef(ts.URL + "/")

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())

		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specAnalyserV2 struct should not be nil", func() {
				So(specAnalyserV2, ShouldNotBeNil)
			})
			Convey("And the new doc should contain the definition ref expanded with the right required fields", func() {
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Required[0], ShouldEqual, "name")
			})
			Convey("And the new doc should contain the definition ref expanded with the right required properties", func() {
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Properties, ShouldContainKey, "id")
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Properties, ShouldContainKey, "name")

			})
			Convey("And the ref should be empty", func() {
				ref := specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Ref.Ref
				So(ref.GetURL(), ShouldBeNil)
			})
		})
	})

	Convey("Given a valid swagger doc where a definition has a ref to an external definition hosted somewhere else (in this case an HTTP server)", t, func() {
		var swaggerJSON = createSwaggerWithExternalRef("myscheme://authority<\"hi\">/foo")

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())

		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldContainSubstring, "error = invalid character 'h' after object key:value pair")
			})
			Convey("AND the specAnalyserV2 struct should  be nil", func() {
				So(specAnalyserV2, ShouldBeNil)
			})
		})
	})

	Convey("Given a valid swagger doc where a definition has a ref to an external definition hosted somewhere else that is unavailable (in this case an HTTP server)", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, createExternalSwaggerContent())
		}))
		defer ts.Close()

		var swaggerJSON = createSwaggerWithExternalRef(ts.URL + "badbadpath")

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())

		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be the expected error", func() {
				So(err.Error(), ShouldContainSubstring, "error = read .: is a directory")
			})
			Convey("AND the specAnalyserV2 struct should be nil", func() {
				So(specAnalyserV2, ShouldBeNil)
			})
		})
	})

	Convey("Given a swagger doc with circular refs", t, func() {
		var externalJSON1 = `{
 "definitions":{
    "OtherKindOfAThing":{
       "$ref":"%s#/definitions/OtherKindOfAThing"
    },
    "ContentDeliveryNetwork":{
       "type":"object",
       "required": [
         "name"
       ],
       "properties":{
          "id":{
             "type":"string",
             "readOnly": true,
          },
          "name":{
             "type":"string"
          }
       }
    }
 }
}`
		externalRefFile1 := initAPISpecFile(externalJSON1)
		defer os.Remove(externalRefFile1.Name())

		var externalJSON2 = `{
 "definitions":{
    "ContentDeliveryNetwork":{
       "$ref":"%s#/definitions/ContentDeliveryNetwork"
    },
    "OtherKindOfAThing":{
       "type":"object",
       "required": [
         "name"
       ],
       "properties":{
          "id":{
             "type":"string",
             "readOnly": true,
          },
          "name":{
             "type":"string"
          }
       }
    }
 }
}`
		externalRefFile2 := initAPISpecFile(externalJSON2)
		defer os.Remove(externalRefFile2.Name())

		var swaggerJSON = fmt.Sprintf(`{
  "swagger":"2.0",
  "paths":{
     "/v1/cdns":{
        "post":{
           "summary":"Create cdn",
           "parameters":[
              {
                 "in":"body",
                 "name":"body",
                 "description":"Created CDN",
                 "schema":{
                    "$ref":"#/definitions/ContentDeliveryNetwork",
                    "$ref":"#/definitions/OtherKindOfAThing"
                 }
              }
           ]
        }
     }
  },
  "definitions":{
     "ContentDeliveryNetwork":{
        "$ref":"%s#/definitions/ContentDeliveryNetwork"
     },
     "OtherKindOfAThing":{
        "$ref":"%s#/definitions/OtherKindOfAThing"
     }
  }
}`, externalRefFile1.Name(), externalRefFile2.Name())

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())
		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specAnalyserV2 struct should not be nil", func() {
				So(specAnalyserV2, ShouldNotBeNil)
			})
			Convey("And the new doc should contain the definition ref expanded with the right required fields", func() {
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Required[0], ShouldEqual, "name")
				So(specAnalyserV2.d.Spec().Definitions["OtherKindOfAThing"].SchemaProps.Required[0], ShouldEqual, "name")
			})
			Convey("And the new doc should contain the definition ref expanded with the right required properties", func() {
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Properties, ShouldContainKey, "id")
				So(specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Properties, ShouldContainKey, "name")
				So(specAnalyserV2.d.Spec().Definitions["OtherKindOfAThing"].SchemaProps.Properties, ShouldContainKey, "id")
				So(specAnalyserV2.d.Spec().Definitions["OtherKindOfAThing"].SchemaProps.Properties, ShouldContainKey, "name")
			})
			Convey("And the ref should be empty", func() {
				ref1 := specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Ref.Ref
				So(ref1.GetURL(), ShouldBeNil)
				ref2 := specAnalyserV2.d.Spec().Definitions["OtherKindOfAThing"].SchemaProps.Ref.Ref
				So(ref2.GetURL(), ShouldBeNil)
			})
		})
	})

	Convey("Given a swagger doc with a circular ref (ref points to itself)", t, func() {
		var swaggerJSON = fmt.Sprintf(`{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
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
      }
   },
   "definitions":{
      "ContentDeliveryNetwork":{
         "$ref":"#/definitions/ContentDeliveryNetwork"
      }
   }
}`)
		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())
		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specAnalyserV2 struct should not be nil", func() {
				So(specAnalyserV2, ShouldNotBeNil)
			})
			Convey("And the new doc should contain the definition ref expanded with the right required fields", func() {
				So(specAnalyserV2.d.Spec().Definitions, ShouldContainKey, "ContentDeliveryNetwork")
			})
			Convey("And the ref should NOT be empty as per the go-openapi library documentation", func() {
				// As per the go-openapi documentation (https://github.com/go-openapi/spec/blob/master/expander.go#L314):
				// this means there is a cycle in the recursion tree: return the Ref
				// - circular refs cannot be expanded. We leave them as ref.
				// - denormalization means that a new local file ref is set relative to the original basePath
				ref1 := specAnalyserV2.d.Spec().Definitions["ContentDeliveryNetwork"].SchemaProps.Ref.Ref
				So(ref1.GetURL().String(), ShouldEqual, "#/definitions/ContentDeliveryNetwork")
			})
		})
	})

	Convey("Given a swagger doc with a ref to a definition that does not exists", t, func() {
		var swaggerJSON = fmt.Sprintf(`{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
            "summary":"Create cdn",
            "parameters":[
               {
                  "in":"body",
                  "name":"body",
                  "description":"Created CDN",
                  "schema":{
                     "$ref":"#/definitions/NonExistingDef"
                  }
               }
            ]
         }
      }
   }
}`)
		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())
		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldContainSubstring, "error = object has no key \"NonExistingDef\"")
			})
			Convey("AND the specAnalyserV2 struct should be nil", func() {
				So(specAnalyserV2, ShouldBeNil)
			})
		})
	})

	Convey("Given a swagger doc with a ref to a definition is wrongly formatted (no empty string)", t, func() {
		var swaggerJSON = fmt.Sprintf(`{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
            "summary":"Create cdn",
            "parameters":[
               {
                  "in":"body",
                  "name":"body",
                  "description":"Created CDN",
                  "schema":{
                     "$ref":
                  }
               }
            ]
         }
      }
   }
}`)
		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())
		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldContainSubstring, "error = invalid character '}' looking for beginning of value")
			})
			Convey("AND the specAnalyserV2 struct should be nil", func() {
				So(specAnalyserV2, ShouldBeNil)
			})
		})
	})

	Convey("Given an swagger doc with a ref to a nonexistent file", t, func() {
		var swaggerJSON = createSwaggerWithExternalRef("nosuchfile.json")

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())
		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned should be not nil", func() {
				So(err.Error(), ShouldContainSubstring, "failed to expand the OpenAPI document from ")
				So(err.Error(), ShouldContainSubstring, " - error = open nosuchfile.json: no such file or directory")
			})
			Convey("AND the specAnalyserV2 struct should be nil", func() {
				So(specAnalyserV2, ShouldBeNil)
			})
		})
	})

	Convey("When newSpecAnalyserV2 method is called with an empty string for openAPIDocumentFilename", t, func() {
		specAnalyserV2, err := newSpecAnalyserV2("")
		Convey("Then the error returned should be not nil", func() {
			So(err.Error(), ShouldEqual, "open api document filename argument empty, please provide the url of the OpenAPI document")
		})
		Convey("AND the specAnalyserV2 struct should be nil", func() {
			So(specAnalyserV2, ShouldBeNil)
		})
	})

	Convey("When newSpecAnalyserV2 method is called with a bogus value openAPIDocumentFilename", t, func() {
		specAnalyserV2, err := newSpecAnalyserV2("nosuchthing")
		Convey("Then the error returned should be not nil", func() {
			So(err.Error(), ShouldEqual, "failed to retrieve the OpenAPI document from 'nosuchthing' - error = open nosuchthing: no such file or directory")
		})
		Convey("AND the specAnalyserV2 struct should be nil", func() {
			So(specAnalyserV2, ShouldBeNil)
		})
	})

}

func TestSpecV2AnalyserGetAllHeaderParameters(t *testing.T) {
	Convey("Given a specV2Analyser loaded with a resources that has a header parameter", t, func() {
		var swaggerJSON = `
{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
            "summary":"Create cdn",
            "parameters":[
               {
                  "in":"body",
                  "name":"body",
                  "description":"Created CDN",
                  "schema":{
                     "$ref":"#/definitions/ContentDeliveryNetwork"
                  }
               },
               {
                  "in":"header",
                  "name":"header_name",
                  "description":"some header to be passed in the POST request"
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
		r := initAPISpecAnalyser(swaggerJSON)
		Convey("When GetAllHeaderParameters method is called", func() {
			specHeaderParameters, err := r.GetAllHeaderParameters()
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the specHeaderParameters size should be one", func() {
				So(len(specHeaderParameters), ShouldEqual, 1)
			})
			Convey("Then the specBackedConfig returned should not be nil", func() {
				So(specHeaderParameters, ShouldContain, SpecHeaderParam{Name: "header_name"})
			})
		})
	})

	Convey("Given a specV2Analyser loaded with few resources that have header parameters", t, func() {
		var swaggerJSON = `
{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
            "summary":"Create cdn",
            "parameters":[
               {
                  "in":"body",
                  "name":"body",
                  "description":"Created CDN",
                  "schema":{
                     "$ref":"#/definitions/ContentDeliveryNetwork"
                  }
               },
               {
                  "in":"header",
                  "name":"header_name",
                  "description":"some header to be passed in the POST request"
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
      },
      "/v1/lbs":{
         "post":{
            "summary":"Create lb",
            "parameters":[
               {
                  "in":"body",
                  "name":"body",
                  "description":"Created LB",
                  "schema":{
                     "$ref":"#/definitions/LB"
                  }
               },
               {
                  "in":"header",
                  "name":"header_name",
                  "description":"some header to be passed in the POST request"
               }
            ]
         }
      },
      "/v1/lbs/{id}":{
         "get":{
            "summary":"Get lb by id"
         },
         "put":{
            "summary":"Updated lb"
         },
         "delete":{
            "summary":"Delete lb"
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
      },
      "LB":{
         "type":"object",
         "properties":{
            "id":{
               "type":"string"
            }
         }
      }
   }
}`
		r := initAPISpecAnalyser(swaggerJSON)
		Convey("When GetAllHeaderParameters method is called", func() {
			specHeaderParameters, err := r.GetAllHeaderParameters()
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the specHeaderParameters should have size one since the same header is present in multiple resources", func() {
				So(len(specHeaderParameters), ShouldEqual, 1)
			})
			Convey("Then the specBackedConfig returned should not be nil", func() {
				So(specHeaderParameters, ShouldContain, SpecHeaderParam{Name: "header_name"})
			})
		})
	})
}

func TestGetAPIBackendConfiguration(t *testing.T) {
	Convey("Given a specV2Analyser", t, func() {
		var swaggerJSON = `
{
   "swagger":"2.0"
}`
		r := initAPISpecAnalyser(swaggerJSON)
		r.openAPIDocumentURL = "http://hostname.com/swagger.json"
		Convey("When GetAPIBackendConfiguration method is called", func() {
			specBackedConfig, err := r.GetAPIBackendConfiguration()
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the specBackedConfig returned should not be nil", func() {
				So(specBackedConfig, ShouldNotBeNil)
			})
		})

	})
}

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
			Convey("And the list returned should contain uswest and useast", func() {
				So(regions, ShouldContain, "uswest")
				So(regions, ShouldContain, "useast")
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
			Convey("And the regions list returned should be empty", func() {
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
			Convey("And the list returned should contain uswestuseast", func() {
				So(regions, ShouldContain, "uswestuseast")
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
			Convey("And the list returned should contain uswest and useast", func() {
				So(regions, ShouldContain, "uswest")
				So(regions, ShouldContain, "useast")
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
        type: "string"`
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
	Convey("Given an specV2Analyser with a terraform compliant root path (and the schema has already been expanded)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          type: "object"
          required:
            - name
          properties:
            id:
              type: "string"
              readOnly: true
            name:
              type: "string"
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
			resourceRootPath, _, resourceRootPostSchemaDef, err := a.validateRootPath("/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resourceRootPath should be", func() {
				So(resourceRootPath, ShouldContainSubstring, "/users")
			})
			Convey("And the resourceRootPostSchemaDef should contain the expected properties", func() {
				So(resourceRootPostSchemaDef.Properties, ShouldContainKey, "id")
				So(resourceRootPostSchemaDef.Properties, ShouldContainKey, "name")
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a resource instance path such as '/users/{id}' that is missing the root path", t, func() {
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
			_, _, _, err := a.validateRootPath("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource instance path '/users/{id}' missing resource root path")
			})
		})
	})

	Convey("Given an apiSpecAnalyser with a resource instance path such as '/users/{id}' but the root is missing the 'body' parameter", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters: # no body parameter
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
			_, _, _, err := a.validateRootPath("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root path '/users' POST operation validation error: resource root operation missing the body parameter")
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
			_, _, _, err := a.validateRootPath("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root path '/users' missing required POST operation")
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
			_, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root path '/users' missing required POST operation")
			})
		})
	})

	Convey("Given an specV2Analyser with a resource that fails the schema validation (body schema is empty)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /users:
    post:
      parameters:
      - in: "body"
        name: "body"
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
			_, _, _, err := a.isEndPointFullyTerraformResourceCompliant("/users/{id}")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "resource root operation missing the schema for the POST operation body parameter")
			})
		})
	})
}

func getExpectedResource(terraformCompliantResources []SpecResource, expectedResourceName string) SpecResource {
	for _, r := range terraformCompliantResources {
		if r.getResourceName() == expectedResourceName {
			return r
		}
	}
	return nil
}

func TestValidateSubResourceTerraformCompliance(t *testing.T) {

	type testCasesDef []struct {
		name          string
		inputResource SpecV2Resource
		expectedError string
	}

	Convey("Given an specV2Analyser with a parent path (both the root and the instance paths)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /cdns:
  /cdns/{id}:
  /cdns/{id}/firewalls:
  /cdns/{id}/firewalls/{id}:`
		a := initAPISpecAnalyser(swaggerContent)
		testCases := testCasesDef{
			{name: "resource containing a subresource path where the parent path exists in the swagger file", inputResource: SpecV2Resource{Path: "/cdns/{id}/firewalls"}, expectedError: ""},
			{name: "resource containing a subresource path where the input resource path path params DO NOT match the parents", inputResource: SpecV2Resource{Path: "/cdns/{cdn_id}/firewalls"}, expectedError: "subresource with path '/cdns/{cdn_id}/firewalls' is missing parent path instance definition '/cdns/{cdn_id}'"},
			{name: "resource containing a subresource path (containing multiple parents) where the parent paths exist in the swagger file", inputResource: SpecV2Resource{Path: "/cdns/{id}/firewalls/{id}/rules"}, expectedError: ""},
			{name: "resource containing a subresource path (containing multiple parents) where one of the parent path DOES NOT exist in the swagger file", inputResource: SpecV2Resource{Path: "/notexisting/{id}/firewalls/{id}/rules"}, expectedError: "subresource with path '/notexisting/{id}/firewalls/{id}/rules' is missing parent path instance definition '/notexisting/{id}'"},
			{name: "resource containing a subresource path where the parent path DOES NOT exists in the swagger file", inputResource: SpecV2Resource{Path: "/resource/{id}/firewalls"}, expectedError: "subresource with path '/resource/{id}/firewalls' is missing parent path instance definition '/resource/{id}'"},
			{name: "resource that is not a subresource", inputResource: SpecV2Resource{Path: "/cdns"}, expectedError: ""},
		}

		for _, tc := range testCases {
			Convey(fmt.Sprintf("When validateSubResourceTerraformCompliance method is called with a %s", tc.name), func() {
				err := a.validateSubResourceTerraformCompliance(tc.inputResource)
				Convey("Then the error returned should be the expected one (if any)", func() {
					So(err == nil || err.Error() == tc.expectedError, ShouldBeTrue)
				})
			})
		}
	})

	Convey("Given an specV2Analyser with a parent path (both the root and the instance paths that use versioning)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /v1/cdns:
  /v1/cdns/{id}:
  /v1/cdns/{id}/v2/firewalls:
  /v1/cdns/{id}/v2/firewalls/{id}:`
		a := initAPISpecAnalyser(swaggerContent)
		testCases := testCasesDef{
			{name: "subresource path where the parent path exists in the swagger file", inputResource: SpecV2Resource{Path: "/v1/cdns/{id}/v2/firewalls"}, expectedError: ""},
			{name: "subresource path (containing multiple parents) where the parent paths exist in the swagger file", inputResource: SpecV2Resource{Path: "/v1/cdns/{id}/v2/firewalls/{id}/rules"}, expectedError: ""},
		}
		for _, tc := range testCases {
			Convey(fmt.Sprintf("When validateSubResourceTerraformCompliance method is called with a %s", tc.name), func() {
				err := a.validateSubResourceTerraformCompliance(tc.inputResource)
				Convey("Then the error returned should be the expected one (if any)", func() {
					So(err == nil || err.Error() == tc.expectedError, ShouldBeTrue)
				})
			})
		}
	})

	Convey("Given an specV2Analyser with a parent path (both the root and the instance paths with trailing paths)", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /cdns/:
  /cdns/{id}/:
  /cdns/{id}/firewalls/:
  /cdns/{id}/firewalls/{id}/:`
		a := initAPISpecAnalyser(swaggerContent)
		testCases := testCasesDef{
			{name: "1 level subresource path where the parent path exists in the swagger file", inputResource: SpecV2Resource{Path: "/cdns/{id}/firewalls"}, expectedError: ""},
			{name: "1 level subresource path with trailing / where the parent path exists in the swagger file", inputResource: SpecV2Resource{Path: "/cdns/{id}/firewalls/"}, expectedError: ""},
		}
		for _, tc := range testCases {
			Convey(fmt.Sprintf("When validateSubResourceTerraformCompliance method is called with a %s", tc.name), func() {
				err := a.validateSubResourceTerraformCompliance(tc.inputResource)
				Convey("Then the error returned should be the expected one (if any)", func() {
					So(err == nil || err.Error() == tc.expectedError, ShouldBeTrue)
				})
			})
		}
	})

	Convey("Given an specV2Analyser with a resource that is ignored", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /cdns:
    post:
      x-terraform-exclude-resource: true
  /cdns/{id}:`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When validateSubResourceTerraformCompliance method is called with a subresource path where the parent path exists in the swagger file", func() {
			inputResource := SpecV2Resource{Path: "/cdns/{id}/firewalls"}
			err := a.validateSubResourceTerraformCompliance(inputResource)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "subresource with path '/cdns/{id}/firewalls' contains a parent /cdns that is marked as ignored, therefore ignoring the subresource too")
			})
		})
	})

	Convey("Given an specV2Analyser with a resource that is missing the parent root path", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /cdns/{id}:`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When validateSubResourceTerraformCompliance method is called with a subresource path where the parent path DOES NOT exists in the swagger file", func() {
			inputResource := SpecV2Resource{Path: "/cdns/{id}/firewalls"}
			err := a.validateSubResourceTerraformCompliance(inputResource)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "subresource with path '/cdns/{id}/firewalls' is missing parent root path definition '/cdns'")
			})
		})
	})
}

func TestCreateMultiRegionResources(t *testing.T) {
	Convey("Given an specV2Analyser loaded with a swagger file containing a multiregion resource", t, func() {
		swaggerContent := `swagger: "2.0"
x-terraform-resource-regions-keyword: "rst1"
paths:
  /v1/cdns:
    post:
      x-terraform-resource-host: some.subdomain.${keyword}.domain.com
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
definitions:
  ContentDeliveryNetwork:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When createMultiRegionResources method is called with a map of regions and corresponding resolved URLs", func() {
			regions := []string{"rst1"}
			resourceRootPath := "/v1/cdns"
			pathRootItem := a.d.Spec().Paths.Paths["/v1/cdns"]
			pathItem := a.d.Spec().Paths.Paths["/v1/cdns/{id}"]
			resourcePayloadSchemaDef := a.d.Spec().Definitions["ContentDeliveryNetwork"]
			multiRegionResources, err := a.createMultiRegionResources(regions, resourceRootPath, pathRootItem, pathItem, &resourcePayloadSchemaDef)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the list resources return should only contain a resource called cdns_v1_rst1", func() {
				So(len(multiRegionResources), ShouldEqual, 1)
				So(multiRegionResources[0].getResourceName(), ShouldEqual, "cdns_v1_rst1")
			})
			cdnMultiRegionResource := multiRegionResources[0]
			Convey("And the host is correctly configured according to the swagger", func() {
				host, err := cdnMultiRegionResource.getHost()
				So(err, ShouldBeNil)
				So(host, ShouldEqual, "some.subdomain.rst1.domain.com")
			})
		})
		Convey("When createMultiRegionResources method is called with a map of regions and an empty resourceRootPath", func() {
			regions := []string{"rst1"}
			resourceRootPath := ""
			pathRootItem := a.d.Spec().Paths.Paths["/v1/cdns"]
			pathItem := a.d.Spec().Paths.Paths["/v1/cdns/{id}"]
			resourcePayloadSchemaDef := a.d.Spec().Definitions["ContentDeliveryNetwork"]
			multiRegionResources, err := a.createMultiRegionResources(regions, resourceRootPath, pathRootItem, pathItem, &resourcePayloadSchemaDef)
			Convey("Then the error returned should be as expected", func() {
				So(err.Error(), ShouldEqual, "failed to create a resource with region: path must not be empty")
			})
			Convey("And multiRegionResources should be nil", func() {
				So(multiRegionResources, ShouldBeNil)
			})
		})
		Convey("When createMultiRegionResources method is called with a list of regions containing empty strings", func() {
			regions := []string{""}
			resourceRootPath := "/v1/cdns"
			pathRootItem := a.d.Spec().Paths.Paths["/v1/cdns"]
			pathItem := a.d.Spec().Paths.Paths["/v1/cdns/{id}"]
			resourcePayloadSchemaDef := a.d.Spec().Definitions["ContentDeliveryNetwork"]
			multiRegionResources, err := a.createMultiRegionResources(regions, resourceRootPath, pathRootItem, pathItem, &resourcePayloadSchemaDef)
			Convey("Then the error returned should be as expected", func() {
				So(err.Error(), ShouldEqual, "multi region host for resource is not valid: region can not be empty for multiregion resources")
			})
			Convey("And multiRegionResources should be nil", func() {
				So(multiRegionResources, ShouldBeNil)
			})
		})
	})
}

func TestGetTerraformCompliantResources(t *testing.T) {

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform subresource /v1/cdns/{id}/v1/firewalls but missing the parent resource resource description", t, func() {
		swaggerContent := `swagger: "2.0"
host: 127.0.0.1 
paths:

  ######################
  ## CDN sub-resource
  ######################

  /v1/cdns/{parent_id}/v1/firewalls:
    post:
      parameters:
      - name: "parent_id"
        in: "path"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
  /v1/cdns/{parent_id}/v1/firewalls/{id}:
    get:
      parameters:
      - name: "parent_id"
        in: "path"
        required: true
        type: "string"
      - name: "id"
        in: "path"
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

definitions:
  ContentDeliveryNetworkFirewallV1:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"`

		a := initAPISpecAnalyser(swaggerContent)
		Convey("When GetTerraformCompliantResources method is called ", func() {
			terraformCompliantResources, err := a.GetTerraformCompliantResources()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the list of resources returned should be empty since the subresource is not considered compliant if the parent is missing", func() {
				So(terraformCompliantResources, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform parent resource /v1/cdns that uses a preferred resource name and a terraform compatible subresource /v1/cdns/{id}/v1/firewalls", t, func() {
		swaggerContent := `swagger: "2.0"
host: 127.0.0.1 
paths:

  ######################
  ## CDN parent resource 
  ######################

  /v1/cdns:
    post:
      x-terraform-resource-name: "cdn"
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
  /v1/cdns/{cdn_id}:
    get:
      parameters:
      - name: "cdn_id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"

  ######################
  ## CDN sub-resource
  ######################

  /v1/cdns/{cdn_id}/v1/firewalls:
    post:
      x-terraform-resource-host: 178.168.3.4
      parameters:
      - name: "cdn_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Created CDN firewall"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
  /v1/cdns/{cdn_id}/v1/firewalls/{id}:
    get:
      parameters:
      - name: "cdn_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - name: "id"
        in: "path"
        description: "The cdn firewall id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
    delete:
      parameters: 
        - description: "The cdn id that contains the firewall to be fetched."
          in: path
          name: parent_id
          required: true
          type: string
        - description: "The cdn firewall id that needs to be fetched."
          in: path
          name: id
          required: true
          type: string
      responses: 
        204:
    put:
      x-terraform-resource-timeout: "300s"
      parameters:
      - name: "id"
        in: "path"
        description: "firewall that needs to be updated"
        required: true
        type: "string"
      - name: "parent_id"
        in: "path"
        description: "cdn which this firewall belongs to"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Updated firewall object"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        200:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

definitions:
  ContentDeliveryNetworkFirewallV1:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"`
		a := initAPISpecAnalyser(swaggerContent)
		Convey("When GetTerraformCompliantResources method is called ", func() {
			terraformCompliantResources, err := a.GetTerraformCompliantResources()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})

			cdnV1Resource := getExpectedResource(terraformCompliantResources, "cdn_v1")
			firewallV1Resource := getExpectedResource(terraformCompliantResources, "cdn_v1_firewalls_v1")

			Convey("And the resources info map should only contain a resource called both the parent cdns_v1 resource and the subresource cdns_v1_firewalls_v1", func() {
				So(len(terraformCompliantResources), ShouldEqual, 2)
				So(cdnV1Resource, ShouldNotBeNil)
				So(firewallV1Resource, ShouldNotBeNil)
			})

			Convey("And the firewall is a subresource which references the parent CDN resource", func() {
				subRes := firewallV1Resource.getParentResourceInfo()
				So(subRes, ShouldNotBeNil)
				So(subRes.parentResourceNames, ShouldResemble, []string{"cdn_v1"})
				So(subRes.fullParentResourceName, ShouldEqual, "cdn_v1")
				Convey("And the full resourcePath is resolved correctly, with the the cdn {parent_id} resolved as 42", func() {
					parentID := "42"
					resourcePath, err := firewallV1Resource.getResourcePath([]string{parentID})
					So(err, ShouldBeNil)
					So(resourcePath, ShouldEqual, "/v1/cdns/42/v1/firewalls")
				})
			})
			Convey("And the firewall resource operations are attached to the resource schema (GET,POST,PUT,DELETE) as stated in the YAML", func() {
				resOperation := firewallV1Resource.getResourceOperations()
				So(resOperation.Get.responses, ShouldContainKey, 200)
				So(resOperation.Post.responses, ShouldContainKey, 201)
				So(resOperation.Put.responses, ShouldContainKey, 200)
				So(resOperation.Delete.responses, ShouldContainKey, 204)
			})
			Convey("And each firewall operation exposed on the resource has its own timeout set", func() {
				timeoutSpec, err := firewallV1Resource.getTimeouts()
				So(err, ShouldBeNil)

				So(timeoutSpec.Put.String(), ShouldEqual, "5m0s")
				So(timeoutSpec.Get, ShouldBeNil)
				So(timeoutSpec.Post, ShouldBeNil)
				So(timeoutSpec.Delete, ShouldBeNil)
			})
			Convey("And the firewall host is correctly configured according to the swagger", func() {
				host, err := firewallV1Resource.getHost()
				So(err, ShouldBeNil)
				So(host, ShouldEqual, "178.168.3.4")
			})

			Convey("And the firewall resource schema contains 3 properties, 2 taken from the model and one added on the fly for the parent resource id", func() {
				actualResourceSchema, err := firewallV1Resource.getResourceSchema()
				So(err, ShouldBeNil)
				So(len(actualResourceSchema.Properties), ShouldEqual, 3)

				idExists, _ := assertPropertyExists(actualResourceSchema.Properties, "id")
				So(idExists, ShouldBeTrue)
				labelExists, _ := assertPropertyExists(actualResourceSchema.Properties, "label")
				So(labelExists, ShouldBeTrue)
				So(actualResourceSchema.Properties[2].Name, ShouldEqual, "cdn_v1_id") //property added on the fly: is a reference to the parent as Firewall is a sub resource
			})

		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource /v1/cdns that is multi region", t, func() {
		swaggerContent := `swagger: "2.0"
x-terraform-resource-regions-keyword: "sea1"
paths:
  /v1/cdns:
    post:
      x-terraform-resource-host: some.subdomain.${keyword}.domain.com
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
definitions:
  ContentDeliveryNetwork:
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
			Convey("And the resources info map should only contain a resource called cdns_v1_sea1", func() {
				So(len(terraformCompliantResources), ShouldEqual, 1)
				So(terraformCompliantResources[0].getResourceName(), ShouldEqual, "cdns_v1_sea1")
			})
			cndV1Resource := terraformCompliantResources[0]
			Convey("And the cndV1Resource should not be considered a subresource", func() {
				subRes := cndV1Resource.getParentResourceInfo()
				So(err, ShouldBeNil)
				So(subRes, ShouldBeNil)
			})
			Convey("And the resource operations are attached to the resource schema (GET,POST,PUT,DELETE) as stated in the YAML", func() {
				resOperation := cndV1Resource.getResourceOperations()
				So(resOperation.Get.responses, ShouldContainKey, 200)
				So(resOperation.Post.responses, ShouldContainKey, 201)
				So(resOperation.Put, ShouldBeNil)
				So(resOperation.Delete, ShouldBeNil)
			})
			Convey("And each operation exposed on the resource has a nil timeout", func() {
				timeoutSpec, err := cndV1Resource.getTimeouts()
				So(err, ShouldBeNil)
				So(timeoutSpec.Post, ShouldBeNil)
				So(timeoutSpec.Get, ShouldBeNil)
				So(timeoutSpec.Put, ShouldBeNil)
				So(timeoutSpec.Delete, ShouldBeNil)
			})
			Convey("And the host is correctly configured according to the swagger", func() {
				host, err := cndV1Resource.getHost()
				So(err, ShouldBeNil)
				So(host, ShouldEqual, "some.subdomain.sea1.domain.com")
			})

			Convey("And the resource schema contains the one property specified in the ContentDeliveryNetwork model definition", func() {
				actualResourceSchema, err := cndV1Resource.getResourceSchema()
				So(err, ShouldBeNil)
				So(len(actualResourceSchema.Properties), ShouldEqual, 1)
				So(actualResourceSchema.Properties[0].Name, ShouldEqual, "id")
			})
		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource /v1/cdns and some non compliant paths", t, func() {
		swaggerContent := `swagger: "2.0"
paths:
  /v1/cdns:
    post:
      x-terraform-resource-timeout: "5s"
      x-terraform-resource-host: some-host.com
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
			cndV1Resource := terraformCompliantResources[0]
			Convey("And the cndV1Resource should not be considered a subresource", func() {
				subRes := cndV1Resource.getParentResourceInfo()
				So(err, ShouldBeNil)
				So(subRes, ShouldBeNil)
			})
			Convey("And the resource operations are attached to the resource schema (GET,POST,PUT,DELETE) as stated in the YAML", func() {
				resOperation := cndV1Resource.getResourceOperations()
				So(resOperation.Get.responses, ShouldContainKey, 200)
				So(resOperation.Post.responses, ShouldContainKey, 201)
				So(resOperation.Put.responses, ShouldContainKey, 200)
				So(resOperation.Delete.responses, ShouldContainKey, 204)
			})
			Convey("And each operation exposed on the resource has it own timeout set", func() {
				timeoutSpec, err := cndV1Resource.getTimeouts()
				So(err, ShouldBeNil)
				So(timeoutSpec.Post.String(), ShouldEqual, "5s")
				So(timeoutSpec.Get, ShouldBeNil)
				So(timeoutSpec.Put, ShouldBeNil)
				So(timeoutSpec.Delete, ShouldBeNil)
			})
			Convey("And the host is correctly configured according to the swagger", func() {
				host, err := cndV1Resource.getHost()
				So(err, ShouldBeNil)
				So(host, ShouldEqual, "some-host.com")
			})

			Convey("And the resource schema contains the one property specified in the ContentDeliveryNetwork model definition", func() {
				actualResourceSchema, err := cndV1Resource.getResourceSchema()
				So(err, ShouldBeNil)
				So(len(actualResourceSchema.Properties), ShouldEqual, 1)
				So(actualResourceSchema.Properties[0].Name, ShouldEqual, "id")
			})
		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource /v1/cdns that has a property being an array of strings", t, func() {
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
definitions:
  ContentDeliveryNetwork:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      listeners:
        type: array
        items:
          type: "string"`
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
			Convey("And the resources schema should contain the right configuration", func() {
				resourceSchema, err := terraformCompliantResources[0].getResourceSchema()
				So(err, ShouldBeNil)
				Convey("And the resources schema should contain the id property", func() {
					exists, _ := assertPropertyExists(resourceSchema.Properties, "id")
					So(exists, ShouldBeTrue)
				})
				Convey("And the resources schema should contain the listeners property", func() {
					exists, idx := assertPropertyExists(resourceSchema.Properties, "listeners")
					So(exists, ShouldBeTrue)
					So(resourceSchema.Properties[idx].Type, ShouldEqual, typeList)
					So(resourceSchema.Properties[idx].ArrayItemsType, ShouldEqual, typeString)
				})
			})

		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource /v1/cdns that has a property being an array objects (using ref)", t, func() {
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
definitions:
  ContentDeliveryNetwork:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      listeners:
        type: array
        items:
          $ref: '#/definitions/Listener'
  Listener:
    type: object
    required:
      - protocol
    properties:
      protocol:
        type: string`
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
			Convey("And the resources schema should contain the right configuration", func() {
				resourceSchema, err := terraformCompliantResources[0].getResourceSchema()
				So(err, ShouldBeNil)
				Convey("And the resources schema should contain the id property", func() {
					exists, _ := assertPropertyExists(resourceSchema.Properties, "id")
					So(exists, ShouldBeTrue)
				})
				Convey("And the resources schema should contain the listeners property", func() {
					exists, idx := assertPropertyExists(resourceSchema.Properties, "listeners")
					So(exists, ShouldBeTrue)
					So(resourceSchema.Properties[idx].Type, ShouldEqual, typeList)
					So(resourceSchema.Properties[idx].ArrayItemsType, ShouldEqual, typeObject)
					So(resourceSchema.Properties[idx].SpecSchemaDefinition.Properties[0].Name, ShouldEqual, "protocol")
					So(resourceSchema.Properties[idx].SpecSchemaDefinition.Properties[0].Type, ShouldEqual, typeString)
				})
			})
		})
	})
	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource /v1/cdns that has a property being an array objects (using ref) (in this an HTTP server)", t, func() {
		var externalJSON = `{
	  "definitions":{
	     "ContentDeliveryNetwork":{
	        "type":"object",
	        "required": [
	          "name"
	        ],
	        "properties":{
	           "id":{
	              "type":"string",
	              "readOnly": true,
	           },
	           "name":{
	              "type":"string"
	           }
	        }
	     }
	  }
	}`

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, externalJSON)
		}))
		defer ts.Close()

		var swaggerJSON = createSwaggerWithExternalRef(ts.URL + "/")

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())

		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned by calling newSpecAnalyserV2 should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specAnalyserV2 struct should not be nil", func() {
				So(specAnalyserV2, ShouldNotBeNil)
			})

			specResources, err := specAnalyserV2.GetTerraformCompliantResources()
			Convey("Then the error returned by calling GetTerraformCompliantResources should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specResources slice should not be nil", func() {
				So(specResources, ShouldNotBeNil)
			})
			Convey("And the resources info map should only contain a resource called cdns_v1", func() {
				So(len(specResources), ShouldEqual, 1)
				So(specResources[0].getResourceName(), ShouldEqual, "cdns_v1")
			})

			Convey("And the resources schema should contain the right configuration", func() {
				resourceSchema, err := specResources[0].getResourceSchema()
				So(err, ShouldBeNil)
				Convey("And the resources schema should contain the id property", func() {
					exists, _ := assertPropertyExists(resourceSchema.Properties, "id")
					So(exists, ShouldBeTrue)
				})
				Convey("And the resources schema should contain the name property", func() {
					exists, _ := assertPropertyExists(resourceSchema.Properties, "name")
					So(exists, ShouldBeTrue)
				})
			})
		})
	})

	Convey("Given an specV2Analyser loaded with a swagger file containing a compliant terraform resource /v1/cdns that has a property being an array objects (nested configuration)", t, func() {
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
definitions:
  ContentDeliveryNetwork:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      listeners:
        type: array
        items:
          type: object
          required:
          - protocol
          properties:
            protocol:
              type: string`
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
			Convey("And the resources schema should contain the right configuration", func() {
				resourceSchema, err := terraformCompliantResources[0].getResourceSchema()
				So(err, ShouldBeNil)
				Convey("And the resources schema should contain the id property", func() {
					exists, _ := assertPropertyExists(resourceSchema.Properties, "id")
					So(exists, ShouldBeTrue)
				})
				Convey("And the resources schema should contain the listeners property", func() {
					exists, idx := assertPropertyExists(resourceSchema.Properties, "listeners")
					So(exists, ShouldBeTrue)
					So(resourceSchema.Properties[idx].Type, ShouldEqual, typeList)
					So(resourceSchema.Properties[idx].ArrayItemsType, ShouldEqual, typeObject)
					So(resourceSchema.Properties[idx].SpecSchemaDefinition.Properties[0].Name, ShouldEqual, "protocol")
					So(resourceSchema.Properties[idx].SpecSchemaDefinition.Properties[0].Type, ShouldEqual, typeString)
				})
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

	Convey("Given an specV2Analyser loaded with a swagger file containing a schema ref that is empty", t, func() {
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
                     "$ref":""
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
			Convey("And the terraformCompliantResources map should be empty since the resource ref is empty", func() {
				So(terraformCompliantResources, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a valid swagger doc where a definition has a ref to an external definition hosted somewhere else (in this case an HTTP server)", t, func() {
		var swaggerJSON = createSwaggerWithExternalRef("//not.a.user@%66%6f%6f.com/just/a/path/also")

		swaggerFile := initAPISpecFile(swaggerJSON)
		defer os.Remove(swaggerFile.Name())

		Convey("When newSpecAnalyserV2 method is called", func() {
			specAnalyserV2, err := newSpecAnalyserV2(swaggerFile.Name())
			Convey("Then the error returned by calling newSpecAnalyserV2 should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specAnalyserV2 struct should not be nil", func() {
				So(specAnalyserV2, ShouldNotBeNil)
			})

			specResources, err := specAnalyserV2.GetTerraformCompliantResources()
			Convey("Then the error returned by calling GetTerraformCompliantResources should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("AND the specResources slice should not be nil", func() {
				So(specResources, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a swagger doc that exposes a resource with not valid multi region configuration (x-terraform-resource-regions-serviceProviderName is missing", t, func() {
		var swaggerJSON = `
{
   "swagger":"2.0",
   "x-terraform-resource-regions-someOtherServiceProvider": "rst, dub",
   "paths":{
      "/v1/cdns":{
         "post":{
            "x-terraform-resource-host": "some.api.${serviceProviderName}.domain.com",
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
		Convey("When GetTerraformCompliantResources method is called", func() {
			r, err := a.GetTerraformCompliantResources()
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be empty", func() {
				So(r, ShouldBeEmpty)
			})
		})
	})
}

func assertPropertyExists(properties specSchemaDefinitionProperties, name string) (bool, int) {
	for idx, prop := range properties {
		if prop.Name == name {
			return true, idx
		}
	}
	return false, -1
}

func initAPISpecAnalyser(swaggerContent string) specV2Analyser {
	file := initAPISpecFile(swaggerContent)
	defer os.Remove(file.Name())
	specV2Analyser, err := newSpecAnalyserV2(file.Name())
	if err != nil {
		log.Panic("newSpecAnalyserV2 failed: ", err)
	}
	return *specV2Analyser
}

func createSwaggerWithExternalRef(filename string) string {
	return fmt.Sprintf(`{
   "swagger":"2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
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
         "$ref":"%s#/definitions/ContentDeliveryNetwork"
      }
   }
}`, filename)
}

func createExternalSwaggerContent() string {
	return `{
  "definitions":{
     "ContentDeliveryNetwork":{
        "type":"object",
        "required": [
          "name"
        ],
        "properties":{
           "id":{
              "type":"string",
              "readOnly": true,
           },
           "name":{
              "type":"string"
           }
        }
     }
  }
}`
}
