package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/dikhan/terraform-provider-openapi/v2/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAcc_ResourceMissingHeaderValue(t *testing.T) {
	apiCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: %s 
schemes:
- "http"
paths:
  /cdns:
    post:
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      - in: header
        type: string
        name: required_header_example
        required: true
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
  /cdns/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`
provider "openapi" {
  required_header_example = ""
}

resource "openapi_cdns" "my_cdn" {
  label = "some label"
}`)

	expectedValidationError, _ := regexp.Compile(".*failed to configure the API request for POST http://127.0.0.1:[\\d]+/cdns: required header 'required_header_example' is missing the value. Please make sure the property 'required_header_example' is configured with a value in the provider's terraform configuration.*")
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config:      tfFileContents,
				ExpectError: expectedValidationError,
			},
		},
	})
	assert.False(t, apiCalled)
}

func TestAcc_ResourceMissingSecurityDefinitionValueAPKIKeyHeader(t *testing.T) {
	apiCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: %s 
schemes:
- "http"
paths:
  /cdns:
    post:
      parameters:
      - in: "body"
        name: "body"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
      security:
      - some_not_global_sec_def: []
  /cdns/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
securityDefinitions:
  some_not_global_sec_def:
    type: apiKey
    name: some_not_global_sec_def
    in: header`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`
provider "openapi" {
  some_not_global_sec_def = ""
}
resource "openapi_cdns" "my_cdn" {
  label = "some label"
}`)

	expectedValidationError, _ := regexp.Compile(".*failed to configure the API request for POST http://127.0.0.1:[\\d]+/cdns: required security definition 'some_not_global_sec_def' is missing the value. Please make sure the property 'some_not_global_sec_def' is configured with a value in the provider's terraform configuration.*")
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config:      tfFileContents,
				ExpectError: expectedValidationError,
			},
		},
	})
	assert.False(t, apiCalled)
}

func TestAcc_ResourceMissingSecurityDefinitionValueAPKIKeyQuery(t *testing.T) {
	apiCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: %s 
schemes:
- "http"
paths:
  /cdns:
    post:
      parameters:
      - in: "body"
        name: "body"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
      security:
      - some_not_global_sec_def: []
  /cdns/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
securityDefinitions:
  some_not_global_sec_def:
    type: apiKey
    name: some_not_global_sec_def
    in: query`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`
provider "openapi" {
  some_not_global_sec_def = ""
}
resource "openapi_cdns" "my_cdn" {
  label = "some label"
}`)

	expectedValidationError, _ := regexp.Compile(".*failed to configure the API request for POST http://127.0.0.1:[\\d]+/cdns: required security definition 'some_not_global_sec_def' is missing the value. Please make sure the property 'some_not_global_sec_def' is configured with a value in the provider's terraform configuration.*")
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config:      tfFileContents,
				ExpectError: expectedValidationError,
			},
		},
	})
	assert.False(t, apiCalled)
}

func TestAcc_ResourceMissingSecurityDefinitionValueRefreshToken(t *testing.T) {
	apiCalled := false
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: %s 
schemes:
- "http"
paths:
  /cdns:
    post:
      parameters:
      - in: "body"
        name: "body"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
      security:
      - some_not_global_sec_def: []
  /cdns/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
securityDefinitions:
  some_not_global_sec_def:
    type: apiKey
    name: some_not_global_sec_def
    in: header
    x-terraform-refresh-token-url: "http://some-refresh-token-url.com"`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`
provider "openapi" {
  some_not_global_sec_def = ""
}
resource "openapi_cdns" "my_cdn" {
  label = "some label"
}`)

	expectedValidationError, _ := regexp.Compile(".*failed to configure the API request for POST http://127.0.0.1:[\\d]+/cdns: required security definition 'some_not_global_sec_def' is missing the value. Please make sure the property 'some_not_global_sec_def' is configured with a value in the provider's terraform configuration.*")
	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config:      tfFileContents,
				ExpectError: expectedValidationError,
			},
		},
	})
	assert.False(t, apiCalled)
}

func TestAcc_ResourceWithNoBodyInput(t *testing.T) {

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"someID", "creation_time": "today", "deploy_key":"someDeployKey"}`))
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: "%s"

schemes:
- "http"

paths:
  /v1/deployKey:
    post:
      x-terraform-resource-name: "deploykey"
      responses:
        201:
          schema:
            $ref: "#/definitions/DeployKeyV1"
  /v1/deployKey/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The deploy key id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/DeployKeyV1"
    delete:
      parameters: 
      - name: "id"
        in: "path"
        description: "The deploy key id to be deleted."
        required: true
        type: "string"
      responses: 
        204: 
          description: "successful operation, no content is returned"
definitions:
  DeployKeyV1: # All the properties are readOnly
    type: "object"
    properties:
      id:
        readOnly: true
        type: string
      creation_time:
        readOnly: true
        type: string
      deploy_key:
        readOnly: true
        type: string`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`resource "openapi_deploykey_v1" "my_deploykeyv1" {}`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "id", "someID"),
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "creation_time", "today"),
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "deploy_key", "someDeployKey"),
				),
			},
		},
	})
}

func TestAcc_PreferredResourceNameOnPathLevel(t *testing.T) {

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"id":"someID", "creation_time": "today", "deploy_key":"someDeployKey"}`))
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: "%s"

schemes:
- "http"

paths:
  /v1/deployKey:
    x-terraform-resource-name: "deploykey"
    post:
      responses:
        201:
          schema:
            $ref: "#/definitions/DeployKeyV1"
  /v1/deployKey/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        description: "The deploy key id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/DeployKeyV1"
    delete:
      parameters: 
      - name: "id"
        in: "path"
        description: "The deploy key id to be deleted."
        required: true
        type: "string"
      responses: 
        204: 
          description: "successful operation, no content is returned"
definitions:
  DeployKeyV1: # All the properties are readOnly
    type: "object"
    properties:
      id:
        readOnly: true
        type: string
      creation_time:
        readOnly: true
        type: string
      deploy_key:
        readOnly: true
        type: string`, apiHost)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`resource "openapi_deploykey_v1" "my_deploykeyv1" {}`)

	resource.Test(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: testAccProviders(provider),
		PreCheck:          func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "id", "someID"),
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "creation_time", "today"),
					resource.TestCheckResourceAttr(
						"openapi_deploykey_v1.my_deploykeyv1", "deploy_key", "someDeployKey"),
				),
			},
		},
	})
}
