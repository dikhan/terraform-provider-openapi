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

func Test_create_and_use_provider_from_yaml_swagger(t *testing.T) {
	apiServerBehaviors := map[string]http.HandlerFunc{}
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("apiServer request>>>>", r.URL, r.Method)
		apiServerBehaviors[r.Method](w, r)
	}))

	apiHost := apiServer.URL[7:]
	fmt.Println("apiHost>>>>", apiHost)

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerReturned := fmt.Sprintf(cdnSwaggerYAMLTemplate, apiHost)
		fmt.Println("swaggerReturned>>>>", swaggerReturned)
		w.Write([]byte(swaggerReturned))
	}))

	fmt.Println("swaggerServer URL>>>>", swaggerServer.URL)
	provider, e := createSchemaProviderFromServiceConfiguration(&ProviderOpenAPI{ProviderName: "bob"}, fakeServiceConfiguration{
		getSwaggerURL: func() string {
			return swaggerServer.URL
		},
	})
	assert.NoError(t, e)

	assert.NotNil(t, provider.ResourcesMap["bob_cdn_v1"])
	//TODO: add'l assertions about provider

	instanceInfo := &terraform.InstanceInfo{Type: "bob_cdn_v1"}

	apiServerBehaviors[http.MethodGet] = func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(">>> GET")
		assert.Equal(t, "/v1/cdns/1337", r.RequestURI)
		bs, e := ioutil.ReadAll(r.Body)
		require.NoError(t, e)
		fmt.Println("GET request body >>>", string(bs))
		apiResponse := `{"id":1337,"label":"CDN #1337","ips":[],"hostnames":[]}`
		w.Write([]byte(apiResponse))
	}

	assert.NoError(t, provider.Configure(&terraform.ResourceConfig{}))

	instanceStates, importStateError := provider.ImportState(instanceInfo, "1337")
	assert.NoError(t, importStateError)
	assert.NotNil(t, instanceStates)
}

func Test_create_and_use_provider_from_json_swagger(t *testing.T) {
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
				swaggerReturned := fmt.Sprintf(bottlesSwaggerJSONTemplate, apiHost)
				fmt.Println("swagger returned >>>>", swaggerReturned)
				w.Write([]byte(swaggerReturned))
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
						fmt.Println(">>>> GET")
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
					swaggerToCauseAPanicByDefiningPUTWithoutStatusCodes := strings.Replace(fmt.Sprintf(bottlesSwaggerJSONTemplate, "whatever.api.host"), bottlePut, `"put": {},`, 1)
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

const bottlesSwaggerJSONTemplate = `{
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

const cdnSwaggerYAMLTemplate = `swagger: "2.0"

info:
  description: "This service provider allows the creation of fake 'cdns' and 'lbs' resources"
  version: "1.0.0"
  title: "Dummy Service Provider generated using 'swaggercodegen' that has two resources 'cdns' and 'lbs' which are terraform compliant"
  contact:
    email: "apiteam@serviceprovider.io"
host: %s
#basePath: ""
tags:
- name: "cdn"
  description: "Operations about cdns"
  externalDocs:
    description: "Find out more about cdn api"
    url: "https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/swaggercodegen"
- name: "lb"
  description: "Operations about lbs"
  externalDocs:
    description: "Find out more about lb api"
    url: "https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/swaggercodegen"
- name: "monitor"
  description: "Operations about monitors"
  externalDocs:
    description: "Find out more about monitor api"
    url: "https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/swaggercodegen"
schemes:
- "http"

consumes:
- "application/json"
produces:
- "application/json"

security:
  - apikey_auth: []

# This make the provider multiregional, so API calls will be make against the specific region as per the value provided
# provided by the user according to the 'x-terraform-provider-regions' regions. If non is provided, the default value will
# be the first item in the 'x-terraform-provider-regions' list of strings. in the case below that will be 'rst1'
x-terraform-provider-regions: "rst1,dub1"

# This is legacy configuration that will be deprecated soon
x-terraform-resource-regions-monitor: "rst1,dub1"

paths:
  /swagger.json:
    get:
      summary: "Api discovery endpoint"
      operationId: "ApiDiscovery"
      responses:
        200:
          description: "successful operation"
  /version:
    get:
      summary: "Get api version"
      operationId: "getVersion"
      responses:
        200:
          description: "successful operation"

  ######################
  #### CDN Resource ####
  ######################

  /v1/cdns:
    post:
      x-terraform-resource-name: "cdn"
      tags:
      - "cdn"
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV1"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      - in: "header"
        x-terraform-header: x_request_id
        name: "X-Request-ID"
        type: "string"
        required: true
      x-terraform-resource-timeout: "30s"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
        default:
          description: "generic error response"
          schema:
            $ref: "#/definitions/Error"
      #security: For the sake of the example, this POST operation will use the global security schemes
      #  - apikey_auth: []
  /v1/cdns/{id}:
    get:
      tags:
      - "cdn"
      summary: "Get cdn by id"
      description: ""
      operationId: "ContentDeliveryNetworkGetV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      #x-terraform-resource-timeout: "30s" If a given operation does not have the 'x-terraform-resource-timeout' extension; the resource operation timeout will default to 10m (10 minutes)
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
        400:
          description: "Invalid cdn id supplied"
        404:
          description: "CDN not found"
      security:
        - apikey_auth: []
    put:
      tags:
      - "cdn"
      summary: "Updated cdn"
      operationId: "ContentDeliveryNetworkUpdateV1"
      parameters:
      - name: "id"
        in: "path"
        description: "cdn that needs to be updated"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Updated cdn object"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
        400:
          description: "Invalid cdn id supplied"
        404:
          description: "CDN not found"
      security:
        - apikey_auth: []
    delete:
      tags:
      - "cdn"
      summary: "Delete cdn"
      operationId: "ContentDeliveryNetworkDeleteV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn that needs to be deleted"
        required: true
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"
        400:
          $ref: "#/responses/Unauthorized"
        404:
          $ref: "#/responses/NotFound"
      security:
        - apikey_auth: []

  ######################
  ##### LB Resource ####
  ######################

  /v1/lbs:
    post:
      tags:
      - "lb"
      summary: "Create lb v1"
      operationId: "LBCreateV1"
      parameters:
      - in: "body"
        name: "body"
        description: "LB v1 payload object to be posted as part of the POST request"
        required: true
        schema:
          $ref: "#/definitions/LBV1"
      x-terraform-resource-timeout: "2s"
      responses:
        202: # Accepted
          x-terraform-resource-poll-enabled: true # [type (bool)] - this flags the response as trully async. Some resources might be async too but may require manual intervention from operators to complete the creation workflow. This flag will be used by the OpenAPI Service provider to detect whether the polling mechanism should be used or not. The flags below will only be applicable if this one is present with value 'true'
          x-terraform-resource-poll-completed-statuses: "deployed" # [type (string)] - Comma separated values with the states that will considered this resource creation done/completed
          x-terraform-resource-poll-pending-statuses: "deploy_pending,deploy_in_progress" # [type (string)] - Comma separated values with the states that are "allowed" and will continue trying
          description: "this operation is asynchronous, to check the status of the deployment call GET operation and check the status field returned in the payload"
          schema:
            $ref: "#/definitions/LBV1"
        default:
          description: "generic error response"
          schema:
            $ref: "#/definitions/Error"
  /v1/lbs/{id}:
    get:
      tags:
      - "lb"
      summary: "Get lb v1 by id"
      description: ""
      operationId: "LBGetV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The lb v1 id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/LBV1"
        400:
          description: "Invalid lb id supplied"
        404:
          description: "LB not found"
    put:
      tags:
      - "lb"
      summary: "Updated cdn"
      operationId: "LBUpdateV1"
      parameters:
      - name: "id"
        in: "path"
        description: "lb v1 that needs to be updated"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Updated cdn object"
        required: true
        schema:
          $ref: "#/definitions/LBV1"
      #      x-terraform-resource-timeout: "30s" If a given operation does not have the 'x-terraform-resource-timeout' extension; the resource operation timeout will default to 10m (10 minutes)
      responses:
        202: # Accepted
          x-terraform-resource-poll-enabled: true
          x-terraform-resource-poll-completed-statuses: "deployed"
          x-terraform-resource-poll-pending-statuses: "deploy_pending,deploy_in_progress"
          schema:
            $ref: "#/definitions/LBV1"
          description: "this operation is asynchronous, to check the status of the deployment call GET operation and check the status field returned in the payload"
        400:
          description: "Invalid lb id supplied"
        404:
          description: "LB v1 not found"
    delete:
      tags:
      - "lb"
      summary: "Delete lb v1"
      operationId: "LBDeleteV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The lb v1 that needs to be deleted"
        required: true
        type: "string"
      responses:
        202:
          description: "LB v1 deletion"
          x-terraform-resource-poll-enabled: true
          #x-terraform-resource-poll-completed-statuses: "destroyed-crazy-nusts!!!" #This extension is not needed in DELETE operations and will be ignored if present. This is due to the fact that when the resource is destroyed, it is expected that http GET calls made by the polling mechanism will get a NotFound response status code back wit no payload whatsoever. And the OpenAPI Terraform provider will internally know how to handle this particular cases without this extension being present.
          x-terraform-resource-poll-pending-statuses: "delete_pending,delete_in_progress"
        400:
          $ref: "#/responses/Unauthorized"
        404:
          $ref: "#/responses/NotFound"


  ############################
  ##### Monitors Multiregion Resource name based ####
  ############################

  # The monitor resource is not implemented in the backed, it only serves as an example on how the global host can be overridden
  # and how the resource can be configured with multi region setup

  /v1/monitors:
    post:
      tags:
      - "monitor"
      summary: "Create monitor v1"
      operationId: "MonitorV1"
      parameters:
      - in: "body"
        name: "body"
        description: "Monitor v1 payload object to be posted as part of the POST request"
        required: true
        schema:
          $ref: "#/definitions/MonitorV1"
      responses:
        200:
          description: "this operation is asynchronous, to check the status of the deployment call GET operation and check the status field returned in the payload"
          schema:
            $ref: "#/definitions/MonitorV1"
        default:
          description: "generic error response"
          schema:
            $ref: "#/definitions/Error"
  /v1/monitors/{id}:
    get:
      tags:
      - "monitor"
      summary: "Get monitor by id"
      description: ""
      operationId: "MonitorV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The monitor v1 id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/MonitorV1"
        400:
          description: "Invalid monitor id supplied"
        404:
          description: "Monitor not found"

  ############################
  ##### Monitors MultiRegion Resource ####
  ############################

  # The monitor resource is not implemented in the backed, it only serves as an example on how the resource not overriding
  # the global host configuration will use by default the multi-region host

  /v1/multiregionmonitors:
    post:
      tags:
      - "multi_region_monitor"
      summary: "Create monitor v1"
      operationId: "MonitorV1"
      parameters:
      - in: "body"
        name: "body"
        description: "Monitor v1 payload object to be posted as part of the POST request"
        required: true
        schema:
          $ref: "#/definitions/MonitorV1"
      responses:
        200:
          description: "this operation is asynchronous, to check the status of the deployment call GET operation and check the status field returned in the payload"
          schema:
            $ref: "#/definitions/MonitorV1"
        default:
          description: "generic error response"
          schema:
            $ref: "#/definitions/Error"
  /v1/multiregionmonitors/{id}:
    get:
      tags:
      - "multi_region_monitor"
      summary: "Get monitor by id"
      description: ""
      operationId: "MonitorV1"
      parameters:
      - name: "id"
        in: "path"
        description: "The monitor v1 id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/MonitorV1"
        400:
          description: "Invalid monitor id supplied"
        404:
          description: "Monitor not found"

securityDefinitions:
  apikey_auth: # basic apikey header auth, the header name will be the value used in the name property, in this case "Authorization", when calling the applicable resource API
    type: "apiKey"
    name: "Authorization"
    in: "header"
#  apikey_header_auth_bearer: // example of header auth using the bearer schema as per the specification
#    type: "apiKey"
#    in: "header"
#    x-terraform-authentication-scheme-bearer: true # this extension would make the auth use the Bearer schema, so
# there will be no need to specify the name property as internally the provider will take care of using the right header name
# following the bearer spec, hence it will use as header name "Authorization", and the token value will be prefixed with
# Bearer schema automatically without needing that input from the user, just the token
#  apikey_query_auth_bearer:  // example of query auth using the bearer schema as per the specification
#    type: "apiKey"
#    in: "query"
#    x-terraform-authentication-scheme-bearer: true # this extension would make the auth use the Bearer schema, so
# there will be no need to specify the name property as internally the provider will take care of using the right header name
# following the bearer spec, hence it will use as header name "Authorization", and the token value will be the one provided
# by the user
#  apikey_query_auth: // basic apikey query auth, the call to the API will attach to the URI the query param, e,g: http://hostname.com?Authorization="value provided by the user"
#    type: "apiKey"
#    name: "Authorization"
#    in: "query"

definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
      - ips
      - hostnames
    properties:
      id:
        type: "string"
        readOnly: true # This property will not be considered when creating a new resource, however, it is expected to
                       # to be returned from the api, and will be saved as computed value in the terraform state file
      label:
        type: "string"
        x-terraform-immutable: true
      ips:
        type: "array"
        x-terraform-force-new: true # when this value changes terraform will force the creation of a new resource
        items:
          type: "string"
      hostnames:
        type: "array"
        items:
          type: "string"
      exampleInt: # this name is not terraform name compliant; the provider will do a translation on the fly to make it terraform name compliant - the result will be example_int
        type: integer
      exampleNumber:
        type: number
        x-terraform-field-name: betterExampleNumberFieldName  # overriding exampleNumber with a different name 'betterExampleNumberFieldName'; the preferred name is not terraform compliant either so the provider will perform the name conversion automatically when translating the name into the provider resource configuration and when saving the field into the state file
      example_boolean:
        type: boolean
      optional_property: # this property is optional as far as input from user is concerned, if the API computes a value if the user does not provider it, see 'optional_computed' or 'optional_computed_with_default' property definitions.
        type: "string"
      computed: # the value of this computed property is not known at runtime (e,g: uuid, etc)
        type: "string"
        readOnly: true
      computed_with_default: # computed property that the default value is known at runtime
        type: "string"
        readOnly: true
        default: "computed value known at runtime" # this computed value happens to be known before hand, the default attribute is just for documentation purposes
      optional_computed: # optional property that the default value is NOT known at runtime
        type: "string"
        x-terraform-computed: true
      optional_computed_with_default: # this computed value happens to be known at runtime, so the service provider decides to document what the default value will be if the client does not provide a value
        type: "string"
        default: "some computed value known at runtime" # this default value, will effectively
      object_property:
        #type: object - type is optional for properties of object type that use $ref
        $ref: "#/definitions/ObjectProperty"
      arrayOfObjectsExample: # This is an example of an array of objects
        type: "array"
        items:
          type: "object"
          properties:
            protocol:
              type: string
            originPort:
              type: integer
              x-terraform-field-name: "origin_port"
      object_nested_scheme_property: # this also covers object within objects
        type: "object"
        x-terraform-computed: true
        properties:
          name:
            type: "string"
            readOnly: true
          object_property:
            type: "object" # nested properties required type equal object to be considered as object
            properties:
              account:
                type: string

  ObjectProperty:
    type: object
    required:
    - message
    - detailedMessage
    - exampleInt
    - exampleNumber
    - example_boolean
    properties:
      message:
        type: string
      detailedMessage:
        type: string
        x-terraform-field-name: "detailed_message"
      exampleInt:
        type: integer
      exampleNumber:
        type: number
      example_boolean:
        type: boolean

  LBV1:
    type: "object"
    required:
    - name
    - backends
    properties:
      id:
        type: "string"
        readOnly: true # This property will not be considered when creating a new resource, however, it is expected to
        # to be returned from the api, and will be saved as computed value in the terraform state file
      name:
        type: "string"
      backends:
        type: "array"
        items:
          type: "string"
      status:
#        x-terraform-field-status: true # identifies the field that should be used as status for async operations. This is handy when the field name is not status but some other name the service provider might have chosen and enables the provider to identify the field as the status field that will be used to track progress for the async operations
        description: lb resource status
        type: string
        readOnly: true
        enum: # this is just for documentation purposes and to let the consumer know what statues should be expected
        - deploy_pending
        - deploy_in_progress
        - deploy_failed
        - deployed
        - delete_pending
        - delete_in_progress
        - delete_failed
        - deleted
      timeToProcess: # time that the resource will take to be processed in seconds
        type: integer
        default: 60 # it will take two minute to process the resource operation (POST/PUT/READ/DELETE)
      simulate_failure: # allows user to set it to true and force an error on the API when the given operation (POST/PUT/READ/DELETE) is being performed
        type: boolean
      newStatus:
        $ref: "#/definitions/Status"

  Status:
    type: object
    readOnly: true
    x-terraform-field-status: true # identifies the field that should be used as status for async operations. This is handy when the field name is not status but some other name the service provider might have chosen and enables the provider to identify the field as the status field that will be used to track progress for the async operations
    properties:
      message:
        type: string
      status:
        type: string

  MonitorV1:
    type: "object"
    required:
    - name
    properties:
      id:
        type: "string"
        readOnly: true
      name:
        type: "string"

  # Schema for error response body
  Error:
    type: object
    required:
    - code
    - message
    properties:
      code:
        type: string
      message:
        type: string

# Descriptions of common responses
responses:
  NotFound:
    description: The specified resource was not found
    schema:
      $ref: "#/definitions/Error"
  Unauthorized:
    description: Unauthorized
    schema:
      $ref: "#/definitions/Error"
`
