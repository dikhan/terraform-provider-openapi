package openapi

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"

	. "github.com/smartystreets/goconvey/convey"
)

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

	Convey("Given a local server that exposes a swagger file containing a terraform compatible reource (cdn)", t, func() {
		swaggerContent := `swagger: "2.0"

host: "localhost:8443"
basePath: "/api"

schemes:
- "https"

security:
  - apikey_auth: []

paths:

  /v1/cdns:
    post:
      summary: "Create cdn"
      x-terraform-resource-name: "cdn"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
  /v1/cdns/{id}:
    get:
      summary: "Get cdn by id"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
    put:
      summary: "Updated cdn"
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
    delete:
      summary: "Delete cdn"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn that needs to be deleted"
        required: true
        type: "string"
      responses:
        204:
          description: "successful operation, no content is returned"

securityDefinitions:
  apikey_auth:
    type: "apiKey"
    name: "Authorization"
    in: "header"

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
        type: "string"`

		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(swaggerContent))
		}))

		Convey("When CreateSchemaProviderWithConfiguration method is called", func() {
			providerName := "openapi"
			p := ProviderOpenAPI{ProviderName: providerName}
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerServer.URL})

			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the provider returned should be configured as expected", func() {
				So(tfProvider, ShouldNotBeNil)
				Convey("the provider schema should be the expected one", func() {
					So(tfProvider.Schema, ShouldNotBeNil)
					So(tfProvider.Schema, ShouldContainKey, "apikey_auth")
					So(tfProvider.Schema["apikey_auth"].Required, ShouldBeTrue)
					So(tfProvider.Schema["apikey_auth"].Type, ShouldEqual, schema.TypeString)
				})
				Convey("the provider resource map should contain the cdn resource with the expected configuration", func() {
					So(tfProvider.ResourcesMap, ShouldNotBeNil)
					resourceName := fmt.Sprintf("%s_cdn_v1", providerName)
					So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
					Convey("the provider cdn resource should have the expected schema", func() {
						So(tfProvider.ResourcesMap, ShouldNotBeNil)
						resourceName := fmt.Sprintf("%s_cdn_v1", providerName)
						So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
						So(tfProvider.ResourcesMap[resourceName].Schema, ShouldContainKey, "label")
						So(tfProvider.ResourcesMap[resourceName].Schema["label"].Type, ShouldEqual, schema.TypeString)
						So(tfProvider.ResourcesMap[resourceName].Schema["label"].Required, ShouldBeTrue)
						So(tfProvider.ResourcesMap[resourceName].Schema["label"].Computed, ShouldBeFalse)
					})
					Convey("the provider cdn resource should have the expected operations configured", func() {
						So(tfProvider.ResourcesMap[resourceName].Create, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Read, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Update, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Delete, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Importer, ShouldNotBeNil)
					})
				})
				Convey("the provider configuration function should not be nil", func() {
					So(tfProvider.ConfigureFunc, ShouldNotBeNil)
				})
			})
		})
	})

	Convey("Given a local server that exposes a swagger file containing a terraform compatible resource taht has a model containing nested objects", t, func() {
		swaggerContent := `swagger: "2.0"
host: "localhost:8443"
basePath: "/api"
schemes:
- "https"
paths:
  /v1/cdns:
    post:
      summary: "Create cdn"
      x-terraform-resource-name: "cdn"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"
  /v1/cdns/{id}:
    get:
      summary: "Get cdn by id"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
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
      object_nested_scheme_property:
        type: "object"
        properties:
          name:
            type: "string"
            readOnly: true
          object_property:
            type: "object"
            properties:
              account:
                type: string
              read_only:
                type: string
                readOnly: true`

		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(swaggerContent))
		}))

		Convey("When CreateSchemaProviderWithConfiguration method is called", func() {
			providerName := "openapi"
			p := ProviderOpenAPI{ProviderName: providerName}
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerServer.URL})
			Convey("Then the error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the provider returned should be configured as expected", func() {
				So(tfProvider, ShouldNotBeNil)
				Convey("the provider schema should only include the endpoints property that enables users to override the resource host from the configuration", func() {
					So(tfProvider.Schema, ShouldNotBeNil)
					assertTerraformSchemaProperty(tfProvider.Schema, "endpoints", schema.TypeSet, false, false)
					So(tfProvider.Schema["endpoints"].Elem.(*schema.Resource).Schema, ShouldContainKey, "cdn_v1")
				})
				Convey("the provider resource map should contain the cdn resource with the expected configuration", func() {
					So(tfProvider.ResourcesMap, ShouldNotBeNil)
					resourceName := fmt.Sprintf("%s_cdn_v1", providerName)
					So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
					Convey("the provider cdn resource should have the expected schema", func() {
						So(tfProvider.ResourcesMap, ShouldNotBeNil)
						resourceName := fmt.Sprintf("%s_cdn_v1", providerName)
						So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
						assertTerraformSchemaProperty(tfProvider.ResourcesMap[resourceName].Schema, "label", schema.TypeString, true, false)
						assertTerraformSchemaNestedObjectProperty(tfProvider.ResourcesMap[resourceName].Schema, "object_nested_scheme_property", false, false)
						nestedObject := tfProvider.ResourcesMap[resourceName].Schema["object_nested_scheme_property"]
						assertTerraformSchemaProperty(nestedObject.Elem.(*schema.Resource).Schema, "name", schema.TypeString, false, true)
						assertTerraformSchemaProperty(nestedObject.Elem.(*schema.Resource).Schema, "object_property", schema.TypeMap, false, false)
						object := nestedObject.Elem.(*schema.Resource).Schema["object_property"]
						assertTerraformSchemaProperty(object.Elem.(*schema.Resource).Schema, "account", schema.TypeString, false, false)
						assertTerraformSchemaProperty(object.Elem.(*schema.Resource).Schema, "read_only", schema.TypeString, false, true)
					})
					Convey("the provider cdn resource should have the expected operations configured", func() {
						So(tfProvider.ResourcesMap[resourceName].Create, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Read, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Update, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Delete, ShouldNotBeNil)
						So(tfProvider.ResourcesMap[resourceName].Importer, ShouldNotBeNil)
					})
				})
				Convey("the provider configuration function should not be nil", func() {
					So(tfProvider.ConfigureFunc, ShouldNotBeNil)
				})
			})
		})
	})
}

func assertTerraformSchemaProperty(actualSchema map[string]*schema.Schema, expectedPropertyName string, expectedType schema.ValueType, expectedRequired, expectedComputed bool) {
	fmt.Printf(">>> Validating '%s' property settings\n", expectedPropertyName)
	So(actualSchema, ShouldContainKey, expectedPropertyName)
	property := actualSchema[expectedPropertyName]
	So(property.Type, ShouldEqual, expectedType)
	if expectedRequired {
		So(property.Required, ShouldBeTrue)
		So(property.Optional, ShouldBeFalse)
	} else {
		So(property.Optional, ShouldBeTrue)
		So(property.Required, ShouldBeFalse)
	}
	So(property.Computed, ShouldEqual, expectedComputed)
}

func assertTerraformSchemaNestedObjectProperty(actualSchema map[string]*schema.Schema, expectedPropertyName string, expectedRequired, expectedComputed bool) {
	assertTerraformSchemaProperty(actualSchema, expectedPropertyName, schema.TypeList, expectedRequired, expectedComputed)
	So(actualSchema[expectedPropertyName].MaxItems, ShouldEqual, 1)
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
