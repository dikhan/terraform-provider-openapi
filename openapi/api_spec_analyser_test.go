package openapi

import (
	"encoding/json"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestResourceInstanceRegex(t *testing.T) {
	Convey("Given an apiSpecAnalyser", t, func() {
		a := apiSpecAnalyser{}
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
	Convey("Given an apiSpecAnalyser", t, func() {
		a := apiSpecAnalyser{}
		Convey("When resourceInstanceRegex method is called with a valid resource path such as '/resource/{id}'", func() {
			resourceInstance, err := a.isResourceInstanceEndPoint("/resource/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(resourceInstance, ShouldBeTrue)
			})
		})
		Convey("When resourceInstanceRegex method is called with an invalid resource path such as '/resource/not/instance/path' not conforming with the expected pattern '/resource/{id}'", func() {
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

func TestFindMatchingRootPath(t *testing.T) {
	Convey("Given an apiSpecAnalyser", t, func() {
		a := apiSpecAnalyser{}
		Convey("When findMatchingResourceRootPath method is called with a valid resource path such as '/users/{id}' and paths containing that path with trailing slash", func() {
			paths := map[string]spec.PathItem{}
			pathItem := spec.PathItem{}
			pathItem.Post = &spec.Operation{}
			paths["/users/"] = pathItem
			resourceRootPath, err := a.findMatchingResourceRootPath("/users/{id}", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be '/users/'", func() {
				So(resourceRootPath, ShouldEqual, "/users/")
			})
		})
		Convey("When findMatchingResourceRootPath method is called with a valid resource path such as '/users/{id}' and paths containing that path without a trailing slash", func() {
			paths := map[string]spec.PathItem{}
			pathItem := spec.PathItem{}
			pathItem.Post = &spec.Operation{}
			paths["/users"] = pathItem
			resourceRootPath, err := a.findMatchingResourceRootPath("/users/{id}", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be '/users'", func() {
				So(resourceRootPath, ShouldEqual, "/users")
			})
		})
		Convey("When findMatchingResourceRootPath method is called with a valid resource path that is versioned such as '/v1/users/{id}' and paths containing that resource with version", func() {
			paths := map[string]spec.PathItem{}
			pathItem := spec.PathItem{}
			pathItem.Post = &spec.Operation{}
			paths["/v1/users"] = pathItem
			resourceRootPath, err := a.findMatchingResourceRootPath("/v1/users/{id}", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be '/v1/users'", func() {
				So(resourceRootPath, ShouldEqual, "/v1/users")
			})
		})
		Convey("When findMatchingResourceRootPath method is called with an invalid resource path such as '/resource/not/instance/path'", func() {
			paths := map[string]spec.PathItem{}
			resourceRootPath, err := a.findMatchingResourceRootPath("/resource/not/instance/path", paths) // instnace paths are of form */{id}
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be empty", func() {
				So(resourceRootPath, ShouldBeEmpty)
			})
		})
	})
}

func TestGetResourceName(t *testing.T) {
	Convey("Given an apiSpecAnalyser", t, func() {
		a := apiSpecAnalyser{}
		Convey("When getResourceName method is called with a valid resource instance path such as '/users/{id}'", func() {
			resourceName, err := a.getResourceName("/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be 'users'", func() {
				So(resourceName, ShouldEqual, "users")
			})
		})

		Convey("When getResourceName method is called with an invalid resource instance path such as '/resource/not/instance/path'", func() {
			_, err := a.getResourceName("'/resource/not/instance/path'")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When getResourceName method is called with a valid resource instance path that is versioned such as '/v1/users/{id}'", func() {
			resourceName, err := a.getResourceName("/v1/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be 'users_v1'", func() {
				So(resourceName, ShouldEqual, "users_v1")
			})
		})

		Convey("When getResourceName method is called with a valid resource instance path that is versioned but long such as '/v1/something/users/{id}'", func() {
			resourceName, err := a.getResourceName("/v1/something/users/{id}")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should still be 'users_v1'", func() {
				So(resourceName, ShouldEqual, "users_v1")
			})
		})
	})
}

func TestPostIsPresent(t *testing.T) {
	Convey("Given an apiSpecAnalyser", t, func() {
		a := apiSpecAnalyser{}
		Convey("When postIsPresent method is called with a path '/users' that has a post operation'", func() {
			paths := map[string]spec.PathItem{
				"/users": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			}
			postIsPresent := a.postIsPresent("/users", paths)
			Convey("Then the value returned should be true", func() {
				So(postIsPresent, ShouldBeTrue)
			})
		})

		Convey("When postIsPresent method is called with a path '/users' that DOES NOT have a post operation'", func() {
			paths := map[string]spec.PathItem{
				"/users": {
					PathItemProps: spec.PathItemProps{},
				},
			}
			postIsPresent := a.postIsPresent("/users", paths)
			Convey("Then the value returned should be false", func() {
				So(postIsPresent, ShouldBeFalse)
			})
		})
	})
}

func TestIsEndPointTerraformResourceCompliant(t *testing.T) {
	Convey("Given an apiSpecAnalyser", t, func() {
		a := apiSpecAnalyser{}
		// This use case covers the bare minimum operations (GET/POST) that a resource has to expose to be considered a terraform resource
		Convey("When TestIsEndPointTerraformResourceCompliant method is called with an resource instance path such as '/users/{id}' that has a GET operation and the corresponding resource root path '/users' exposes a POST operation", func() {
			paths := map[string]spec.PathItem{
				"/users": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/users/{id}": {
					PathItemProps: spec.PathItemProps{
						Get: &spec.Operation{},
					},
				},
			}
			isEndPointTerraformResourceCompliant, err := a.isEndPointTerraformResourceCompliant("/users/{id}", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(isEndPointTerraformResourceCompliant, ShouldBeTrue)
			})
		})

		// This is the ideal case where the resource exposes al CRUD operations
		Convey("When TestIsEndPointTerraformResourceCompliant method is called with an resource instance path such as '/users/{id}' that has a GET/PUT/DELETE operations exposed and the corresponding resource root path '/users' exposes a POST operation", func() {
			paths := map[string]spec.PathItem{
				"/users": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/users/{id}": {
					PathItemProps: spec.PathItemProps{
						Get:    &spec.Operation{},
						Delete: &spec.Operation{},
						Put:    &spec.Operation{},
					},
				},
			}
			isEndPointTerraformResourceCompliant, err := a.isEndPointTerraformResourceCompliant("/users/{id}", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(isEndPointTerraformResourceCompliant, ShouldBeTrue)
			})
		})

		// This use case avoids resource duplicates as the root paths are filtered out
		Convey("When TestIsEndPointTerraformResourceCompliant method is called with a non resource instance path such as '/users'", func() {
			paths := map[string]spec.PathItem{
				"/users": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/users/{id}": {
					PathItemProps: spec.PathItemProps{
						Get: &spec.Operation{},
					},
				},
			}
			isEndPointTerraformResourceCompliant, err := a.isEndPointTerraformResourceCompliant("/users", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be false", func() {
				So(isEndPointTerraformResourceCompliant, ShouldBeFalse)
			})
		})

		Convey("When TestIsEndPointTerraformResourceCompliant method is called with a resource instance path such as '/monitors/{id}' that DOES NOT expose a GET operation", func() {
			paths := map[string]spec.PathItem{
				"/monitors": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/monitors/{id}": {
					PathItemProps: spec.PathItemProps{},
				},
			}
			isEndPointTerraformResourceCompliant, err := a.isEndPointTerraformResourceCompliant("/monitors/{id}", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be false as a resource to be considered compliant has to expose a POST operation on the root path (in this case /monitors) and a GET operation in the instance resource path (in this case /monitors/{id})", func() {
				So(isEndPointTerraformResourceCompliant, ShouldBeFalse)
			})
		})

		Convey("When TestIsEndPointTerraformResourceCompliant method is called with a resource instance path such as '/monitors/{id}' that its root path '/monitors' DOES NOT expose a POST operation", func() {
			paths := map[string]spec.PathItem{
				"/monitors": {
					PathItemProps: spec.PathItemProps{},
				},
				"/monitors/{id}": {
					PathItemProps: spec.PathItemProps{
						Get: &spec.Operation{},
					},
				},
			}
			isEndPointTerraformResourceCompliant, err := a.isEndPointTerraformResourceCompliant("/monitors/{id}", paths)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be false as a resource to be considered compliant has to expose a POST operation on the root path (in this case /monitors) and a GET operation in the instance resource path (in this case /monitors/{id})", func() {
				So(isEndPointTerraformResourceCompliant, ShouldBeFalse)
			})
		})
	})
}

func TestGetResourcePayloadSchemaDef(t *testing.T) {
	Convey("Given an apiSpecAnalyser loaded with the following swagger", t, func() {
		var swaggerJSON = `{
  "swagger": "2.0",
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
  "definitions": {
    "ContentDeliveryNetwork": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string"
        }
      }
    }
  }
}`
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}

		Convey("When getResourcePayloadSchemaDef method is called with a root path '/v1/cdns' containing a valid ref: '#/definitions/ContentDeliveryNetwork'", func() {
			resourcePayloadSchemaDef, err := a.getResourcePayloadSchemaDef("/v1/cdns")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be a valid schema def", func() {
				So(len(resourcePayloadSchemaDef.Type), ShouldEqual, 1)
				So(resourcePayloadSchemaDef.Type, ShouldContain, "object")
			})
		})

		Convey("When getResourcePayloadSchemaDef method is called with an unknown root path", func() {
			_, err := a.getResourcePayloadSchemaDef("/non/existing/root/path")
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with the following swagger", t, func() {
		var swaggerJSON = `{
  "swagger": "2.0",
   "paths":{
      "/v1/cdns":{
      }
   }
}`
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}

		Convey("When getResourcePayloadSchemaDef method is called with a root path '/v1/cdns' that is missing the post operation", func() {
			_, err := a.getResourcePayloadSchemaDef("/v1/cdns")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with the following swagger", t, func() {
		var swaggerJSON = `{
  "swagger": "2.0",
   "paths":{
      "/v1/cdns":{
         "post":{
            "summary":"Create cdn"
         }
      }
   }
}`
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}

		Convey("When getResourcePayloadSchemaDef method is called with a root path '/v1/cdns' that is missing the parameters section", func() {
			_, err := a.getResourcePayloadSchemaDef("/v1/cdns")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestGetResourcesInfo(t *testing.T) {
	Convey("Given an apiSpecAnalyser loaded with a swagger file containing a compliant terraform resource /v1/cdns", t, func() {
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
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}
		Convey("When getResourcesInfo method is called ", func() {
			resourcesInfo, err := a.getResourcesInfo()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resources info map should contain a resource called cdns_v1", func() {
				So(len(resourcesInfo), ShouldEqual, 1)
				So(resourcesInfo, ShouldContainKey, "cdns_v1")
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with a swagger file containing a non compliant terraform resource /v1/cdns because its missing the post operation", t, func() {
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
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}
		Convey("When getResourcesInfo method is called ", func() {
			resourcesInfo, err := a.getResourcesInfo()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resources info map should contain a resource called cdns_v1", func() {
				So(resourcesInfo, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with a swagger file containing a non compliant terraform resource /v1/cdns/{id} because its missing the get operation", t, func() {
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
               }
            ]
         }
      },
      "/v1/cdns/{id}":{
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
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}
		Convey("When getResourcesInfo method is called ", func() {
			resourcesInfo, err := a.getResourcesInfo()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the resources info map should contain a resource called cdns_v1", func() {
				So(resourcesInfo, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with a swagger file containing a potential compliant terraform resource but the payload schema definition is missing", t, func() {
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
   }
}`
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}
		Convey("When getResourcesInfo method is called ", func() {
			_, err := a.getResourcesInfo()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with a swagger file containing a potential compliant terraform resource but the POST body is missing the reference to the schema definition", t, func() {
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
                  "description":"Created CDN"
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
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}
		Convey("When getResourcesInfo method is called ", func() {
			_, err := a.getResourcesInfo()
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with a swagger file containing a compliant terraform resource that has a POST operation with multiple parameters", t, func() {
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
                  "name":"X-Request-ID",
                  "type": "string",
				  "required": true
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
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}
		Convey("When getResourcesInfo method is called ", func() {
			_, err := a.getResourcesInfo()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an apiSpecAnalyser loaded with a swagger file containing a potential compliant terraform resource that has a POST operation with multiple 'body' type parameters", t, func() {
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
                  "in":"body",
                  "name":"body2",
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
		spec, _ := loads.Analyzed(json.RawMessage([]byte(swaggerJSON)), "2.0")
		a := apiSpecAnalyser{
			d: spec,
		}
		Convey("When getResourcesInfo method is called ", func() {
			_, err := a.getResourcesInfo()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
