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
	provider, e := createSchemaProviderFromServiceConfiguration(&ProviderOpenAPI{ProviderName: "bob"}, fakeServiceConfiguration{
		getSwaggerURL: func() string {
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("apiServer request>>>>", r.URL, r.Method)
				switch r.Method {
				case http.MethodGet:
					assert.Equal(t, "/bottles/1337", r.RequestURI)
					bs, e := ioutil.ReadAll(r.Body)
					require.NoError(t, e)
					assert.Empty(t, string(bs))
					w.Write([]byte(`{"id":1337,"name":"Bottle #1337","rating":17,"vintage":1977,"anotherbottle":{"id":"nestedid1","name":"nestedname1"}}`))
				case http.MethodPut:
					assert.Equal(t, "/bottles/1337", r.RequestURI)
					bs, e := ioutil.ReadAll(r.Body)
					require.NoError(t, e)
					assert.Equal(t, `{"anotherbottle":{"id":"nestedid1","name":"nestedname1"},"name":"whatever","rating":17,"vintage":1977}`, string(bs))
					w.Write([]byte(`{"id":1337,"name":"leet bottle ftw","rating":17,"vintage":1977,"anotherbottle":{"id":"updatednested1","name":"updatednestedname1"}}`))
				case http.MethodDelete:
					assert.Equal(t, "/bottles/1337", r.RequestURI)
					bs, e := ioutil.ReadAll(r.Body)
					require.NoError(t, e)
					assert.Empty(t, string(bs))
					w.Write([]byte(`{}`))
				}
			}))

			apiHost := apiServer.URL[7:]
			fmt.Println("apiHost>>>>", apiHost)

			swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(fmt.Sprintf(swaggerTemplate, apiHost)))
			}))

			fmt.Println("swaggerServer URL>>>>", swaggerServer.URL)
			return swaggerServer.URL
		},
	})
	assert.NoError(t, e)
	assert.Equal(t, schema.TypeString, provider.ResourcesMap["bob_bottles"].Schema["name"].Type)
	assert.Equal(t, schema.TypeInt, provider.ResourcesMap["bob_bottles"].Schema["vintage"].Type)
	assert.Equal(t, schema.TypeInt, provider.ResourcesMap["bob_bottles"].Schema["rating"].Type)
	assert.Equal(t, schema.TypeMap, provider.ResourcesMap["bob_bottles"].Schema["anotherbottle"].Type)
	assert.Equal(t, schema.TypeString, provider.ResourcesMap["bob_bottles"].Schema["anotherbottle"].Elem.(*schema.Resource).Schema["name"].Type)

	instanceInfo := &terraform.InstanceInfo{Type: "bob_bottles"}
	assert.Panics(t, func() { provider.ImportState(instanceInfo, "whatever") }, "ImportState panics if Configure hasn't been called first")

	assert.NoError(t, provider.Configure(&terraform.ResourceConfig{}))

	instanceStates, e := provider.ImportState(instanceInfo, "1337")
	assert.NoError(t, e)
	assert.NotNil(t, instanceStates)
	assert.Equal(t, 1, len(instanceStates))
	initialInstanceState := instanceStates[0]
	assert.Equal(t, "1337", initialInstanceState.ID)
	assert.Equal(t, "Bottle #1337", initialInstanceState.Attributes["name"])
	assert.Equal(t, "17", initialInstanceState.Attributes["rating"])
	assert.Equal(t, "1977", initialInstanceState.Attributes["vintage"])
	assert.Equal(t, "nestedid1", initialInstanceState.Attributes["anotherbottle.id"])
	assert.Equal(t, "nestedname1", initialInstanceState.Attributes["anotherbottle.name"])

	updatedInstanceState, updateError := provider.Apply(instanceInfo, initialInstanceState, &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{"name": {Old: "whatever", New: "whatever"}}})
	assert.NoError(t, updateError)
	assert.NotNil(t, updatedInstanceState)
	assert.Equal(t, "1337", updatedInstanceState.ID)
	assert.Equal(t, "leet bottle ftw", updatedInstanceState.Attributes["name"])
	assert.Equal(t, "17", updatedInstanceState.Attributes["rating"])
	assert.Equal(t, "1977", updatedInstanceState.Attributes["vintage"])
	assert.Equal(t, "updatednested1", updatedInstanceState.Attributes["anotherbottle.id"])
	assert.Equal(t, "updatednestedname1", updatedInstanceState.Attributes["anotherbottle.name"])

	deletedInstanceState, deleteError := provider.Apply(instanceInfo, initialInstanceState, &terraform.InstanceDiff{Destroy: true})
	assert.NoError(t, deleteError)
	assert.Nil(t, deletedInstanceState)
}

func Test_ImportState_panics_if_swagger_defines_put_without_response_status_codes(t *testing.T) {
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

	instanceInfo := &terraform.InstanceInfo{Type: "bob_bottles"}

	require.NoError(t, provider.Configure(&terraform.ResourceConfig{}))

	assert.Panics(t, func() { provider.ImportState(instanceInfo, "1337") })
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
