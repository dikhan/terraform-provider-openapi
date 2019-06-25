package openapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform/terraform"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

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
        "summary": "create bottle",
        "description": "creates a bottle",
        "operationId": "bottle#create",
        "parameters": [
          {
            "name": "payload",
            "in": "body",
            "description": "BottlePayload is the type used to create bottles",
            "required": true,
            "schema": {
              "$ref": "#/definitions/BottlePayload"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "Created",
            "schema": {
              "$ref": "#/definitions/bottle"
            }
          },
          "400": {
            "description": "Bad Request",
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
      },
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
            "description": "OK",
            "schema": {
              "$ref": "#\/definitions\/bottle"
            }
          },
          "404": {
            "description": "Not Found"
          }
        }
      },
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
            "description": "OK",
            "schema": {
              "$ref": "#\/definitions\/bottle"
            }
          },
          "404": {
            "description": "Not Found"
          }
        }
      }
    },
    "/swagger/swagger.json": {
      "get": {
        "operationId": "Spec#/swagger/swagger.json",
        "responses": {
          "200": {
            "description": "File downloaded",
            "schema": {
              "type": "file"
            }
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
          "description": "Unique bottle ID",
          "example": "Enim sapiente expedita sit.",
          "readOnly": true
        },
        "name": {
          "type": "string",
          "description": "Name of bottle",
          "example": "x",
          "minLength": 1
        },
        "rating": {
          "type": "integer",
          "description": "Rating of bottle",
          "example": 4,
          "minimum": 1,
          "maximum": 5
        },
        "vintage": {
          "type": "integer",
          "description": "Vintage of bottle",
          "example": 2653,
          "minimum": 1900
        }
      },
      "description": "BottlePayload is the type used to create bottles",
      "example": {
        "id": "Enim sapiente expedita sit.",
        "name": "x",
        "rating": 4,
        "vintage": 2653
      },
      "required": [
        "name",
        "vintage",
        "rating"
      ]
    },
    "bottle": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "description": "Unique bottle ID",
          "example": "Voluptates non excepturi.",
          "readOnly": true
        },
        "name": {
          "type": "string",
          "description": "Name of bottle",
          "example": "krt",
          "minLength": 1
        },
        "rating": {
          "type": "integer",
          "description": "Rating of bottle",
          "example": 3,
          "minimum": 1,
          "maximum": 5
        },
        "nestedftw": {
          "type": "object",
		  "properties": {
        	"name": {
        	  "type": "string",
        	  "description": "Name of bottle",
        	  "example": "x",
        	  "minLength": 1
        	},
		  },
        },
        "vintage": {
          "type": "integer",
          "description": "Vintage of bottle",
          "example": 1932,
          "minimum": 1900
        }
      },
      "description": "bottle media type (default view)",
      "example": {
        "id": "Voluptates non excepturi.",
        "name": "krt",
        "rating": 3,
        "vintage": 1932
      },
      "required": [
        "id",
        "name",
        "vintage",
        "rating"
      ]
    },
    "error": {
      "type": "object",
      "properties": {
        "code": {
          "type": "string",
          "description": "an application-specific error code, expressed as a string value.",
          "example": "invalid_value"
        },
        "detail": {
          "type": "string",
          "description": "a human-readable explanation specific to this occurrence of the problem.",
          "example": "Value of ID must be an integer"
        },
        "id": {
          "type": "string",
          "description": "a unique identifier for this particular occurrence of the problem.",
          "example": "3F1FKVRR"
        },
        "meta": {
          "type": "object",
          "description": "a meta object containing non-standard meta-information about the error.",
          "example": {
            "timestamp": 1458609066
          },
          "additionalProperties": true
        },
        "status": {
          "type": "string",
          "description": "the HTTP status code applicable to this problem, expressed as a string value.",
          "example": "400"
        }
      },
      "description": "Error response media type (default view)",
      "example": {
        "code": "invalid_value",
        "detail": "Value of ID must be an integer",
        "id": "3F1FKVRR",
        "meta": {
          "timestamp": 1458609066
        },
        "status": "400"
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

const swaggerToCauseAPanicByDefiningPUTWithoutStatusCodes = `{
  "swagger": "2.0",
  "host": "whatever-host",
  "consumes": [
    "application\/json"
  ],
  "produces": [
    "application\/json"
  ],
  "paths": {
    "/bottles/": {
      "post": {
        "summary": "create bottle",
        "description": "creates a bottle",
        "operationId": "bottle#create",
        "parameters": [
          {
            "name": "payload",
            "in": "body",
            "description": "BottlePayload is the type used to create bottles",
            "required": true,
            "schema": {
              "$ref": "#/definitions/BottlePayload"
            }
          }
        ],
        "responses": {
          "201": {
            "description": "Created",
            "schema": {
              "$ref": "#/definitions/bottle"
            }
          },
          "400": {
            "description": "Bad Request",
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
      "put": {},
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
            "description": "OK",
            "schema": {
              "$ref": "#\/definitions\/bottle"
            }
          },
          "404": {
            "description": "Not Found"
          }
        }
      }
    },
    "/swagger/swagger.json": {
      "get": {
        "operationId": "Spec#/swagger/swagger.json",
        "responses": {
          "200": {
            "description": "File downloaded",
            "schema": {
              "type": "file"
            }
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
          "description": "Unique bottle ID",
          "example": "Enim sapiente expedita sit.",
          "readOnly": true
        },
        "name": {
          "type": "string",
          "description": "Name of bottle",
          "example": "x",
          "minLength": 1
        },
        "rating": {
          "type": "integer",
          "description": "Rating of bottle",
          "example": 4,
          "minimum": 1,
          "maximum": 5
        },
        "vintage": {
          "type": "integer",
          "description": "Vintage of bottle",
          "example": 2653,
          "minimum": 1900
        }
      },
      "description": "BottlePayload is the type used to create bottles",
      "example": {
        "id": "Enim sapiente expedita sit.",
        "name": "x",
        "rating": 4,
        "vintage": 2653
      },
      "required": [
        "name",
        "vintage",
        "rating"
      ]
    },
    "bottle": {
      "type": "object",
      "properties": {
        "id": {
          "type": "string",
          "description": "Unique bottle ID",
          "example": "Voluptates non excepturi.",
          "readOnly": true
        },
        "name": {
          "type": "string",
          "description": "Name of bottle",
          "example": "krt",
          "minLength": 1
        },
        "rating": {
          "type": "integer",
          "description": "Rating of bottle",
          "example": 3,
          "minimum": 1,
          "maximum": 5
        },
        "vintage": {
          "type": "integer",
          "description": "Vintage of bottle",
          "example": 1932,
          "minimum": 1900
        }
      },
      "description": "bottle media type (default view)",
      "example": {
        "id": "Voluptates non excepturi.",
        "name": "krt",
        "rating": 3,
        "vintage": 1932
      },
      "required": [
        "id",
        "name",
        "vintage",
        "rating"
      ]
    },
    "error": {
      "type": "object",
      "properties": {
        "code": {
          "type": "string",
          "description": "an application-specific error code, expressed as a string value.",
          "example": "invalid_value"
        },
        "detail": {
          "type": "string",
          "description": "a human-readable explanation specific to this occurrence of the problem.",
          "example": "Value of ID must be an integer"
        },
        "id": {
          "type": "string",
          "description": "a unique identifier for this particular occurrence of the problem.",
          "example": "3F1FKVRR"
        },
        "meta": {
          "type": "object",
          "description": "a meta object containing non-standard meta-information about the error.",
          "example": {
            "timestamp": 1458609066
          },
          "additionalProperties": true
        },
        "status": {
          "type": "string",
          "description": "the HTTP status code applicable to this problem, expressed as a string value.",
          "example": "400"
        }
      },
      "description": "Error response media type (default view)",
      "example": {
        "code": "invalid_value",
        "detail": "Value of ID must be an integer",
        "id": "3F1FKVRR",
        "meta": {
          "timestamp": 1458609066
        },
        "status": "400"
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
	o := &ProviderOpenAPI{ProviderName: "bob"}

	provider, e := o.createSchemaProviderFromServiceConfiguration(fakeServiceConfiguration{
		getSwaggerURL: func() string {
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("apiServer request>>>>", r.URL, r.Method)
				switch r.Method {
				case http.MethodGet:
					w.Write([]byte(`{"id":1337,"name":"Bottle #1337","rating":17,"vintage":1977}`))
				case http.MethodPut:
					w.Write([]byte(`{"id":1337,"name":"leet bottle ftw","rating":17,"vintage":1977}`))
				case http.MethodDelete:
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
	assert.NotNil(t, provider)

	assert.Equal(t, 1, len(provider.Schema))
	assert.Equal(t, 1, len(provider.ResourcesMap))

	instanceInfo := &terraform.InstanceInfo{Type: "bob_bottles"}
	assert.Panics(t, func() { provider.ImportState(instanceInfo, "my fancy id") }, "ImportState panics if Configure hasn't been called first")

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

	updatedInstanceState, updateError := provider.Apply(instanceInfo, initialInstanceState, &terraform.InstanceDiff{Attributes: map[string]*terraform.ResourceAttrDiff{"name": {Old: "whatever", New: "whatever"}}})
	assert.NoError(t, updateError)
	assert.NotNil(t, updatedInstanceState)
	assert.Equal(t, "1337", updatedInstanceState.ID)
	assert.Equal(t, "leet bottle ftw", updatedInstanceState.Attributes["name"])
	assert.Equal(t, "17", updatedInstanceState.Attributes["rating"])
	assert.Equal(t, "1977", updatedInstanceState.Attributes["vintage"])

	deletedInstanceState, deleteError := provider.Apply(instanceInfo, initialInstanceState, &terraform.InstanceDiff{Destroy: true})
	assert.NoError(t, deleteError)
	assert.Nil(t, deletedInstanceState)
}

func Test_ImportState_panics_if_swagger_defines_put_without_response_status_codes(t *testing.T) {
	o := &ProviderOpenAPI{ProviderName: "bob"}

	provider, e := o.createSchemaProviderFromServiceConfiguration(fakeServiceConfiguration{
		getSwaggerURL: func() string {
			swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
