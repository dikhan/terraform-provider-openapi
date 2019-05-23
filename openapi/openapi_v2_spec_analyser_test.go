package openapi

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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

func TestNewSpecAnalyserV2(t *testing.T) {
	Convey("Given a valid swagger doc where a definition has a ref to an external definition hosted somewhere else (in this case file system)", t, func() {
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
		externalRefFile := initAPISpecFile(externalJSON)
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

	Convey("Given a valid swagger doc where a definition has a ref to an external definition hosted somewhere else (in this an HTTP server)", t, func() {
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

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			fmt.Fprintln(w, externalJSON)
		}))
		defer ts.Close()

		var swaggerJSON = createSwaggerWithExternalRef(ts.URL +"/")

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

	Convey("Given an specV2Analyser and an operation that has an embedded schema in the body parameter (this mimics what getResourcePayloadSchemaDef will get when the spec expansion has been performed)", t, func() {
		a := specV2Analyser{}
		operation := &spec.Operation{
			OperationProps: spec.OperationProps{
				Parameters: []spec.Parameter{
					{
						ParamProps: spec.ParamProps{
							In:   "body",
							Name: "body",
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Ref:      spec.Ref{},
									Required: []string{"name"},
									Type:     spec.StringOrArray{"object"},
									Properties: map[string]spec.Schema{
										"id": spec.Schema{
											SwaggerSchemaProps: spec.SwaggerSchemaProps{
												ReadOnly: true,
											},
											SchemaProps: spec.SchemaProps{
												Type: spec.StringOrArray{"string"},
											},
										},
										"name": spec.Schema{
											SwaggerSchemaProps: spec.SwaggerSchemaProps{},
											SchemaProps: spec.SchemaProps{
												Type: spec.StringOrArray{"string"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Convey("When getResourcePayloadSchemaDef method is called with an operation containing an expanded schema'", func() {
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
	swagger := json.RawMessage([]byte(swaggerContent))
	d, _ := loads.Analyzed(swagger, "2.0")
	return specV2Analyser{d: d}
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
