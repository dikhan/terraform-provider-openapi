package openapi

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	. "github.com/smartystreets/goconvey/convey"
)

type fakeServiceSchemaPropertyConfiguration struct {
}

func (fakeServiceSchemaPropertyConfiguration) GetDefaultValue() (string, error) {
	return "whatever default value", nil
}
func (fakeServiceSchemaPropertyConfiguration) ExecuteCommand() error {
	return nil
}

type fakeServiceConfiguration struct {
	getSwaggerURL func() string
}

func (c fakeServiceConfiguration) GetSwaggerURL() string {
	return c.getSwaggerURL()
}
func (fakeServiceConfiguration) GetPluginVersion() string {
	return "whatever plugin version"
}
func (fakeServiceConfiguration) IsInsecureSkipVerifyEnabled() bool {
	return false
}
func (fakeServiceConfiguration) GetSchemaPropertyConfiguration(schemaPropertyName string) ServiceSchemaPropertyConfiguration {
	return fakeServiceSchemaPropertyConfiguration{}
}
func (fakeServiceConfiguration) Validate(runningPluginVersion string) error {
	return nil
}

func Test_create_and_use_provider_from_json(t *testing.T) {
	Convey("Given an API server", t, func() {
		apiServerBehaviors := map[string]http.HandlerFunc{}
		apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("apiServer request>>>>", r.URL, r.Method)
			apiServerBehaviors[r.Method](w, r)
		}))

		Convey("And given the URL for a swagger document describing the API", func() {
			apiHost := apiServer.URL[7:]
			fmt.Println("apiHost>>>>", apiHost)

			swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(fmt.Sprintf(swaggerTemplate, apiHost)))
			}))

			fmt.Println("swaggerServer URL>>>>", swaggerServer.URL)

			Convey("A provider can be built from that swagger", func() {
				provider, e := createSchemaProviderFromServiceConfiguration(&ProviderOpenAPI{ProviderName: "bob"}, fakeServiceConfiguration{
					getSwaggerURL: func() string {
						return swaggerServer.URL
					},
				})
				So(e, ShouldBeNil)

				Convey("And the resulting provider should reflect the resource definitions from that swagger", func() {
					So(schema.TypeString, ShouldEqual, provider.ResourcesMap["bob_bottles"].Schema["name"].Type)
					So(schema.TypeInt, ShouldEqual, provider.ResourcesMap["bob_bottles"].Schema["vintage"].Type)
					So(schema.TypeInt, ShouldEqual, provider.ResourcesMap["bob_bottles"].Schema["rating"].Type)
					So(schema.TypeMap, ShouldEqual, provider.ResourcesMap["bob_bottles"].Schema["anotherbottle"].Type)
					So(schema.TypeString, ShouldEqual, provider.ResourcesMap["bob_bottles"].Schema["anotherbottle"].Elem.(*schema.Resource).Schema["name"].Type)
				})

				instanceInfo := &terraform.InstanceInfo{Type: "bob_bottles"}

				Convey("And calling ImportState before Configure will panic", func() {
					assert.Panics(t, func() { provider.ImportState(instanceInfo, "whatever") }, "ImportState panics if Configure hasn't been called first")
				})

				Convey("But ImportState works fine if Configure is called first", func() {
					e := provider.Configure(&terraform.ResourceConfig{})
					So(e, ShouldBeNil)

					var receivedGetToURI string
					var receivedBodyInGetRequest string
					apiServerBehaviors[http.MethodGet] = func(w http.ResponseWriter, r *http.Request) {
						receivedGetToURI = r.RequestURI
						bs, e := ioutil.ReadAll(r.Body)
						require.NoError(t, e)
						receivedBodyInGetRequest = string(bs)
						w.Write([]byte(`{"id":1337,"name":"Bottle #1337","rating":17,"vintage":1977,"anotherbottle":{"id":"nestedid1","name":"nestedname1"}}`))
					}

					var instanceStates []*terraform.InstanceState
					var importStateError error
					instanceStates, importStateError = provider.ImportState(instanceInfo, "1337")
					So(importStateError, ShouldBeNil)

					Convey("And the API server should receive the appropriate request", func() {
						So(receivedGetToURI, ShouldEqual, "/bottles/1337")
						So(receivedBodyInGetRequest, ShouldBeEmpty)
					})

					var initialInstanceState *terraform.InstanceState
					initialInstanceState = instanceStates[0]

					Convey("And the instance state returned should reflect the content of the API server's response", func() {
						So(1, ShouldEqual, len(instanceStates))
						So("1337", ShouldEqual, initialInstanceState.ID)
						So("Bottle #1337", ShouldEqual, initialInstanceState.Attributes["name"])
						So("17", ShouldEqual, initialInstanceState.Attributes["rating"])
						So("1977", ShouldEqual, initialInstanceState.Attributes["vintage"])
						So("nestedid1", ShouldEqual, initialInstanceState.Attributes["anotherbottle.id"])
						So("nestedname1", ShouldEqual, initialInstanceState.Attributes["anotherbottle.name"])
					})

					Convey("And changes can then be made to the resource by calling Apply", func() {

						var receivedPutToURI string
						var receivedPutBody string
						apiServerBehaviors[http.MethodPut] = func(w http.ResponseWriter, r *http.Request) {
							receivedPutToURI = r.RequestURI
							bs, e := ioutil.ReadAll(r.Body)
							require.NoError(t, e)
							receivedPutBody = string(bs)
							w.Write([]byte(`{"id":1337,"name":"leet bottle ftw","rating":17,"vintage":1977,"anotherbottle":{"id":"updatednested1","name":"updatednestedname1"}}`))
						}

						updatedInstanceState, updateError := provider.Apply(instanceInfo, initialInstanceState, &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{"name": {Old: "whatever", New: "whatever"}}})
						So(updateError, ShouldBeNil)

						Convey("And the API server should receive the appropriate request", func() {
							So("/bottles/1337", ShouldEqual, receivedPutToURI)
							So(receivedPutBody, ShouldEqual, `{"anotherbottle":{"id":"nestedid1","name":"nestedname1"},"name":"whatever","rating":17,"vintage":1977}`)
						})

						Convey("And the instance state returned should reflect the content of the API server's response", func() {
							So("1337", ShouldEqual, updatedInstanceState.ID)
							So("leet bottle ftw", ShouldEqual, updatedInstanceState.Attributes["name"])
							So("17", ShouldEqual, updatedInstanceState.Attributes["rating"])
							So("1977", ShouldEqual, updatedInstanceState.Attributes["vintage"])
							So("updatednested1", ShouldEqual, updatedInstanceState.Attributes["anotherbottle.id"])
							So("updatednestedname1", ShouldEqual, updatedInstanceState.Attributes["anotherbottle.name"])
						})
					})

					Convey("And the resouce can be deleted", func() {
						var receivedDeleteToURI string
						var receivedDeleteBody string

						apiServerBehaviors[http.MethodDelete] = func(w http.ResponseWriter, r *http.Request) {
							receivedDeleteToURI = r.RequestURI
							bs, e := ioutil.ReadAll(r.Body)
							require.NoError(t, e)
							receivedDeleteBody = string(bs)
							w.Write([]byte(`{}`))
						}

						deletedInstanceState, deleteError := provider.Apply(instanceInfo, initialInstanceState, &terraform.InstanceDiff{Destroy: true})
						So(deleteError, ShouldBeNil)

						Convey("And the API server should receive the appropriate request", func() {
							So("/bottles/1337", ShouldEqual, receivedDeleteToURI)
							So(receivedDeleteBody, ShouldBeEmpty)
						})

						So(deletedInstanceState, ShouldBeNil)
					})
				})
			})
		})
	})
}

func Test_ImportState_panics_if_swagger_defines_put_without_response_status_codes(t *testing.T) {
	Convey("Given a provider built from a swagger crafted with an empty method block", t, func() {
		provider, e := createSchemaProviderFromServiceConfiguration(&ProviderOpenAPI{ProviderName: "bob"}, fakeServiceConfiguration{
			getSwaggerURL: func() string {
				swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					swaggerToCauseAPanicByDefiningPUTWithoutStatusCodes := strings.Replace(fmt.Sprintf(swaggerTemplate, "whatever.api.host"), bottlePut, `"put": {},`, 1)
					w.Write([]byte(swaggerToCauseAPanicByDefiningPUTWithoutStatusCodes))
				}))
				return swaggerServer.URL
			},
		})
		require.NoError(t, e)
		require.NotNil(t, provider)
		require.NoError(t, provider.Configure(&terraform.ResourceConfig{}))

		Convey("When ImportState is called, it will panic", func() {
			assert.Panics(t, func() { provider.ImportState(&terraform.InstanceInfo{Type: "bob_bottles"}, "1337") })
		})
	})
}

func TestOpenAPIProvider(t *testing.T) {
	Convey("Given a provider name missing the service configuration", t, func() {
		providerName := "nonExistingProvider"
		Convey("When getServiceConfiguration method is called", func() {
			p := ProviderOpenAPI{ProviderName: providerName}
			tfProvider, err := p.CreateSchemaProvider()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(err.Error(), ShouldContainSubstring, "plugin init error")
			})
			Convey("Then the schema provider returned should also be nil", func() {
				So(tfProvider, ShouldBeNil)
			})
		})
	})

	Convey("Given a provider name with service configuration but there is an error with the OpenAPI spec analyser", t, func() {
		providerName := "providerName"
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		attemptedSwaggerURL := s.URL + "/swagger.yaml"
		os.Setenv(fmt.Sprintf(otfVarSwaggerURL, providerName), attemptedSwaggerURL)
		os.Setenv(otfVarInsecureSkipVerify, "false")
		Convey("When getServiceConfiguration method is called", func() {
			p := ProviderOpenAPI{ProviderName: providerName}
			tfProvider, err := p.CreateSchemaProvider()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message returned should be", func() {
				So(err.Error(), ShouldEqual, "plugin OpenAPI spec analyser error: failed to retrieve the OpenAPI document from '"+attemptedSwaggerURL+`' - error = could not access document at "`+attemptedSwaggerURL+`" [404 Not Found] `)
			})
			Convey("Then the schema provider returned should also be nil", func() {
				So(tfProvider, ShouldBeNil)
			})
		})
	})
}

func TestGetServiceConfiguration(t *testing.T) {
	Convey("Given a swagger url configured with environment variable and skip verify being false", t, func() {
		providerName := "providerName"
		expectedSwaggerURL := "http://www.domain.com/swagger.yaml"
		os.Setenv(fmt.Sprintf(otfVarSwaggerURL, providerName), expectedSwaggerURL)
		os.Setenv(otfVarInsecureSkipVerify, "false")
		Convey("When getServiceConfiguration method is called", func() {
			serviceConfiguration, err := getServiceConfiguration(providerName)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the service configuration swagger URL should be the expected one", func() {
				So(serviceConfiguration.GetSwaggerURL(), ShouldEqual, expectedSwaggerURL)
			})
			Convey("And the service configuration should be false", func() {
				So(serviceConfiguration.IsInsecureSkipVerifyEnabled(), ShouldBeFalse)
			})
		})
	})
}

const swaggerTemplate = `{
  "swagger": "2.0",
  "host": "%s",
  "consumes": [
    "application\/json"
  ],
  "produces": [
    "application\/json"
  ],
  "paths": {
    "/bottles/": {
      "post": {
        "operationId": "bottle#create",
        "parameters": [
          {
            "name": "payload",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/BottlePayload"
            }
          }
        ],
        "responses": {
          "201": {
            "schema": {
              "$ref": "#/definitions/bottle"
            }
          },
          "400": {
            "schema": {
              "$ref": "#/definitions/error"
            }
          },
          "500": {
            "description": "Internal Server Error"
          }
        }
      }
    },
    "/bottles/{id}": {
      "delete": {
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "description": "OK",
            "schema": {
              "$ref": "#\/definitions\/bottle"
            }
          },
          "404": {
            "description": "Not Found"
          }
        }
      },` + bottlePut + `
      "get": {
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "schema": {
              "$ref": "#\/definitions\/bottle"
            }
          },
          "404": {
            "description": "Not Found"
          }
        }
      }
    }
  },
  "definitions": {
    "BottlePayload": {
      "title": "BottlePayload",
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "readOnly": true
        },
        "name": {
          "type": "string",
          "minLength": 1
        },
        "rating": {
          "type": "integer",
          "minimum": 1,
          "maximum": 5
        },
        "vintage": {
          "type": "integer",
          "minimum": 1900
        },
		"anotherbottle": {
		  "type": "object",
		  "description": "another bottle within a bottle",
		  "properties": {
			"id": {
			  "type": "string",
			  "readOnly": true
			},
			"name": {
			  "type": "string",
			  "minLength": 1
			}
          }
		}
      },
      "required": [
        "name",
        "vintage",
        "rating",
		"anotherbottle"
      ]
    },
    "bottle": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "readOnly": true
        },
        "name": {
          "type": "string",
          "minLength": 1
        },
        "rating": {
          "type": "integer",
          "minimum": 1,
          "maximum": 5
        },
        "vintage": {
          "type": "integer",
          "minimum": 1900
        },
		"anotherbottle": {
		  "type": "object",
		  "properties": {
			"id": {
			  "type": "string",
			  "readOnly": true
			},
			"name": {
			  "type": "string",
			  "minLength": 1
			}
          }
		}
      },
      "required": [
        "id",
        "name",
        "vintage",
        "rating",
		"anotherbottle"
      ]
    },
    "error": {
      "type": "object",
      "properties": {
        "code": {
          "type": "string"
        },
        "id": {
          "type": "string"
        },
        "status": {
          "type": "string"
        }
      }
    }
  },
  "responses": {
    "InternalServerError": {
      "description": "Internal Server Error"
    },
    "NotFound": {
      "description": "Not Found"
    }
  }
}`

const bottlePut = `
      "put": {
        "parameters": [
          {
            "name": "id",
            "in": "path",
            "required": true,
            "type": "string"
          }
        ],
        "responses": {
          "200": {
            "schema": {
              "$ref": "#\/definitions\/bottle"
            }
          },
          "404": {
            "description": "Not Found"
          }
        }
      },
`
