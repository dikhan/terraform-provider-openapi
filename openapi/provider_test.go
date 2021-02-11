package openapi

import (
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestOpenAPIProvider(t *testing.T) {
	Convey("Given a provider name missing the service configuration", t, func() {
		providerName := "nonExistingProvider"
		Convey("When getServiceConfiguration method is called", func() {
			p := ProviderOpenAPI{ProviderName: providerName}
			tfProvider, err := p.CreateSchemaProvider()
			Convey("Then the schema provider returned should also be nil and the error returned should be the expected one", func() {
				So(err.Error(), ShouldContainSubstring, "plugin init error")
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
			Convey("Then the schema provider returned should also be nil and the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "plugin OpenAPI spec analyser error: failed to retrieve the OpenAPI document from '"+attemptedSwaggerURL+`' - error = could not access document at "`+attemptedSwaggerURL+`" [404 Not Found] `)
				So(tfProvider, ShouldBeNil)
			})
		})
	})

	Convey("Given a local server that exposes a swagger file containing a terraform compatible resource (cdn) and subresource (firewall)", t, func() {
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

  /v1/cdns/{id}/firewalls:
    x-terraform-resource-name: "firewall"
    post:
      summary: "Create firewall"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn where the firewall will be created"
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Created firewall"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
  /v1/cdns/{id}/firewalls/{firewall_id}:
    get:
      summary: "Get firewall by id"
      parameters:
      - name: "id"
        in: "path"
        description: "The cdn that the firewall belongs to"
        required: true
        type: "string"
      - name: "firewall_id"
        in: "path"
        description: "The firewall id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

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
        type: "string"
  ContentDeliveryNetworkFirewallV1:
    type: "object"
    required:
      - name
    properties:
      id:
        type: "string"
        readOnly: true
      name:
        type: "string"`

		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(swaggerContent))
		}))

		Convey("When CreateSchemaProviderWithConfiguration method is called", func() {
			providerName := "openapi"
			p := ProviderOpenAPI{ProviderName: providerName}
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerServer.URL})
			Convey("Then the provider returned should be configured as expected and the the error should be nil", func() {
				So(err, ShouldBeNil)
				So(tfProvider, ShouldNotBeNil)

				So(tfProvider.Schema, ShouldNotBeNil)
				So(tfProvider.Schema, ShouldContainKey, "apikey_auth")
				So(tfProvider.Schema["apikey_auth"].Required, ShouldBeTrue)
				So(tfProvider.Schema["apikey_auth"].Type, ShouldEqual, schema.TypeString)

				// the provider resource map should contain the cdn resource with the expected configuration
				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				resourceName := fmt.Sprintf("%s_cdn_v1", providerName)
				So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				resourceName = fmt.Sprintf("%s_cdn_v1", providerName)
				So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
				So(tfProvider.ResourcesMap[resourceName].Schema, ShouldContainKey, "label")
				So(tfProvider.ResourcesMap[resourceName].Schema["label"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.ResourcesMap[resourceName].Schema["label"].Required, ShouldBeTrue)
				So(tfProvider.ResourcesMap[resourceName].Schema["label"].Computed, ShouldBeFalse)
				So(tfProvider.ResourcesMap[resourceName].CreateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].UpdateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].DeleteContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].Importer, ShouldNotBeNil)

				// the provider data source map should contain the cdn data source instance with the expected configuration
				So(tfProvider.DataSourcesMap, ShouldNotBeNil)
				dataSourceInstanceName := fmt.Sprintf("%s_cdn_v1_instance", providerName)
				So(tfProvider.DataSourcesMap, ShouldContainKey, dataSourceInstanceName)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema, ShouldContainKey, "id")
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Required, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Computed, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema, ShouldContainKey, "label")
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["label"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["label"].Required, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["label"].Computed, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Create, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Read, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Update, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Delete, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Importer, ShouldBeNil)

				// the provider resource map should contain the firewall resource with the expected configuration
				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				resourceName = fmt.Sprintf("%s_cdn_v1_firewall", providerName)
				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
				So(tfProvider.ResourcesMap[resourceName].Schema, ShouldContainKey, "name")
				So(tfProvider.ResourcesMap[resourceName].Schema["name"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.ResourcesMap[resourceName].Schema["name"].Required, ShouldBeTrue)
				So(tfProvider.ResourcesMap[resourceName].Schema["name"].Computed, ShouldBeFalse)
				So(tfProvider.ResourcesMap[resourceName].CreateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].UpdateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].DeleteContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].Importer, ShouldNotBeNil)

				// the provider data source map should contain the firewall data source instance with the expected configuration
				So(tfProvider.DataSourcesMap, ShouldNotBeNil)
				dataSourceInstanceName = fmt.Sprintf("%s_cdn_v1_firewall_instance", providerName)
				So(tfProvider.DataSourcesMap, ShouldContainKey, dataSourceInstanceName)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema, ShouldContainKey, "id")
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Required, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Computed, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema, ShouldContainKey, "name")
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["name"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["name"].Required, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["name"].Computed, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Create, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Read, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Update, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Delete, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Importer, ShouldBeNil)

				// the provider configuration function should not be nil
				So(tfProvider.ConfigureFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a local server that exposes a swagger file containing a terraform compatible resource cdn where the POST (request) model only includes input properties (required/optional) and the POST (response) model includes the request properties as readOnly and any other computed property", t, func() {
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
          $ref: "#/definitions/ContentDeliveryNetworkInputV1"
      responses:
        201:
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkOutputV1"
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
            $ref: "#/definitions/ContentDeliveryNetworkOutputV1"

definitions:
  ContentDeliveryNetworkInputV1: # only used in POST request payload
    type: "object"
    required:
      - label
    properties:
      label:
        type: "string"
  ContentDeliveryNetworkOutputV1: # used in both POST response payload and GET response payload
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
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
			Convey("And the provider returned should be configured as expected and the the error should be nil", func() {
				So(tfProvider, ShouldNotBeNil)
				So(tfProvider.Schema, ShouldNotBeNil)
				// the provider resource map should contain the cdn resource with the expected configuration
				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				resourceName := fmt.Sprintf("%s_cdn_v1", providerName)
				So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				resourceName = fmt.Sprintf("%s_cdn_v1", providerName)
				So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
				So(tfProvider.ResourcesMap[resourceName].Schema, ShouldContainKey, "label")
				So(tfProvider.ResourcesMap[resourceName].Schema["label"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.ResourcesMap[resourceName].Schema["label"].Required, ShouldBeTrue)
				So(tfProvider.ResourcesMap[resourceName].Schema["label"].Computed, ShouldBeFalse)
				So(tfProvider.ResourcesMap[resourceName].CreateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].UpdateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].DeleteContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].Importer, ShouldNotBeNil)

				// the provider data source map should contain the cdn data source instance with the expected configuration
				So(tfProvider.DataSourcesMap, ShouldNotBeNil)
				dataSourceInstanceName := fmt.Sprintf("%s_cdn_v1_instance", providerName)
				So(tfProvider.DataSourcesMap, ShouldContainKey, dataSourceInstanceName)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema, ShouldContainKey, "id")
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Required, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["id"].Computed, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema, ShouldContainKey, "label")
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["label"].Type, ShouldEqual, schema.TypeString)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["label"].Required, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Schema["label"].Computed, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].CreateContext, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].UpdateContext, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].DeleteContext, ShouldBeNil)
				So(tfProvider.DataSourcesMap[dataSourceInstanceName].Importer, ShouldBeNil)

				// the provider configuration function should not be nil
				So(tfProvider.ConfigureFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a local server that exposes a swagger file containing a terraform compatible data source (cdn_datasource) using a preferred name defined in the root path level", t, func() {
		swaggerContent := `swagger: "2.0"
host: "localhost:8443"
basePath: "/api"

schemes:
- "https"

security:
 - apikey_auth: []

paths:
 /v1/cdndatasource:
   x-terraform-resource-name: "cdn_datasource"
   get:
     responses:
       200:
         schema:
           $ref: "#/definitions/ContentDeliveryNetworkV1Collection"

securityDefinitions:
 apikey_auth:
   type: "apiKey"
   name: "Authorization"
   in: "header"

definitions:
 ContentDeliveryNetworkV1Collection:
   type: array
   items:
     $ref: "#/definitions/ContentDeliveryNetworkV1"
 ContentDeliveryNetworkV1:
   type: object
   properties:
     id:
       type: string
       readOnly: true
     label:
       type: string
     owners:
       type: array
       items:
         type: string
     int_property:
       type: integer
     bool_property:
       type: boolean
     float_property:
       type: number
       format: float
     obj_property:
       type: object
       properties:
         name:
           type: string`

		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(swaggerContent))
		}))

		Convey("When CreateSchemaProviderWithConfiguration method is called", func() {
			providerName := "openapi"
			p := ProviderOpenAPI{ProviderName: providerName}
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerServer.URL})

			Convey("Then the provider returned should be configured as expected and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(tfProvider, ShouldNotBeNil)
				So(tfProvider.Schema, ShouldNotBeNil)
				So(tfProvider.Schema, ShouldContainKey, "apikey_auth")
				So(tfProvider.Schema["apikey_auth"].Required, ShouldBeTrue)
				So(tfProvider.Schema["apikey_auth"].Type, ShouldEqual, schema.TypeString)

				// the provider dataSource map should contain the cdn resource with the expected configuration
				So(tfProvider.DataSourcesMap, ShouldNotBeNil)
				So(len(tfProvider.DataSourcesMap), ShouldEqual, 1)
				resourceName := fmt.Sprintf("%s_cdn_datasource_v1", providerName)
				So(tfProvider.DataSourcesMap, ShouldContainKey, resourceName)
				resourceName = fmt.Sprintf("%s_cdn_datasource_v1", providerName)
				So(tfProvider.DataSourcesMap, ShouldContainKey, resourceName)

				assertDataSourceSchemaProperty(t, tfProvider.DataSourcesMap[resourceName].Schema["label"], schema.TypeString)
				assertDataSourceSchemaProperty(t, tfProvider.DataSourcesMap[resourceName].Schema["owners"], schema.TypeList)
				assertDataSourceSchemaProperty(t, tfProvider.DataSourcesMap[resourceName].Schema["int_property"], schema.TypeInt)
				assertDataSourceSchemaProperty(t, tfProvider.DataSourcesMap[resourceName].Schema["bool_property"], schema.TypeBool)
				assertDataSourceSchemaProperty(t, tfProvider.DataSourcesMap[resourceName].Schema["float_property"], schema.TypeFloat)
				assertDataSourceSchemaProperty(t, tfProvider.DataSourcesMap[resourceName].Schema["obj_property"], schema.TypeList)

				So(tfProvider.DataSourcesMap[resourceName].Schema, ShouldContainKey, "filter")
				So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Type, ShouldEqual, schema.TypeSet)
				So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Required, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Optional, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Computed, ShouldBeFalse)

				elements := tfProvider.DataSourcesMap[resourceName].Schema["filter"].Elem.(*schema.Resource).Schema
				So(elements["name"].Type, ShouldEqual, schema.TypeString)
				So(elements["values"].Type, ShouldEqual, schema.TypeList)

				// the provider cdn-datasource data source should have only the READ operation configured
				So(tfProvider.DataSourcesMap[resourceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.DataSourcesMap[resourceName].Read, ShouldBeNil)
				So(tfProvider.DataSourcesMap[resourceName].Create, ShouldBeNil)
				So(tfProvider.DataSourcesMap[resourceName].Delete, ShouldBeNil)

				// the provider resource map must be nil as no resources are configured in the swagger"
				So(tfProvider.ResourcesMap, ShouldBeEmpty)
				So(tfProvider.ConfigureFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a local server that exposes a swagger file containing a terraform compatible data source that has a subresource path", t, func() {
		swaggerContent := `swagger: "2.0"
host: "localhost:8443"
basePath: "/api"

paths:

 /v1/cdns/{id}/firewalls:
   get:
     responses:
       200:
         schema:
           $ref: "#/definitions/ContentDeliveryNetworkV1Collection"

definitions:
 ContentDeliveryNetworkV1Collection:
   type: "array"
   items:
     $ref: "#/definitions/ContentDeliveryNetworkV1"
 ContentDeliveryNetworkV1:
   type: "object"
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

			Convey("And the provider returned should be configured as expected and the error should be nil", func() {
				So(tfProvider, ShouldNotBeNil)
				So(err, ShouldBeNil)
				So(tfProvider.Schema, ShouldNotBeNil)
				So(tfProvider.DataSourcesMap, ShouldNotBeNil)
				So(len(tfProvider.DataSourcesMap), ShouldEqual, 1)

				dataSourceName := fmt.Sprintf("%s_cdns_v1_firewalls", providerName)
				So(tfProvider.DataSourcesMap, ShouldContainKey, dataSourceName)

				So(tfProvider.DataSourcesMap, ShouldContainKey, dataSourceName)
				// check parent id is part of the schema
				assertTerraformSchemaProperty(t, tfProvider.DataSourcesMap[dataSourceName].Schema["cdns_v1_id"], schema.TypeString, true, false)
				// check actual model properties are part of the schema
				assertDataSourceSchemaProperty(t, tfProvider.DataSourcesMap[dataSourceName].Schema["label"], schema.TypeString)
				// check filter property is in the schema
				So(tfProvider.DataSourcesMap[dataSourceName].Schema, ShouldContainKey, "filter")
				So(tfProvider.DataSourcesMap[dataSourceName].Schema["filter"].Type, ShouldEqual, schema.TypeSet)
				So(tfProvider.DataSourcesMap[dataSourceName].Schema["filter"].Required, ShouldBeFalse)
				So(tfProvider.DataSourcesMap[dataSourceName].Schema["filter"].Optional, ShouldBeTrue)
				So(tfProvider.DataSourcesMap[dataSourceName].Schema["filter"].Computed, ShouldBeFalse)

				elements := tfProvider.DataSourcesMap[dataSourceName].Schema["filter"].Elem.(*schema.Resource).Schema
				So(elements["name"].Type, ShouldEqual, schema.TypeString)
				So(elements["values"].Type, ShouldEqual, schema.TypeList)
				So(tfProvider.DataSourcesMap[dataSourceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.DataSourcesMap[dataSourceName].Read, ShouldBeNil)
				So(tfProvider.ResourcesMap, ShouldBeEmpty)
				So(tfProvider.ConfigureFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a local server that exposes a swagger file containing a terraform compatible resource that has a model containing nested objects", t, func() {
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
			Convey("Then the provider returned should be configured as expected and the error should be nil", func() {
				So(tfProvider, ShouldNotBeNil)
				So(err, ShouldBeNil)

				// provider schema should only include the endpoints property that enables users to override the resource host from the configuration
				So(tfProvider.Schema, ShouldNotBeNil)
				assertTerraformSchemaProperty(t, tfProvider.Schema["endpoints"], schema.TypeSet, false, false)
				So(tfProvider.Schema["endpoints"].Elem.(*schema.Resource).Schema, ShouldContainKey, "cdn_v1")

				// provider resource map should contain the cdn resource with the expected configuration
				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				resourceName := fmt.Sprintf("%s_cdn_v1", providerName)
				So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)

				So(tfProvider.ResourcesMap, ShouldNotBeNil)
				resourceName = fmt.Sprintf("%s_cdn_v1", providerName)
				So(tfProvider.ResourcesMap, ShouldContainKey, resourceName)
				assertTerraformSchemaProperty(t, tfProvider.ResourcesMap[resourceName].Schema["label"], schema.TypeString, true, false)
				assertTerraformSchemaNestedObjectProperty(t, tfProvider.ResourcesMap[resourceName].Schema["object_nested_scheme_property"], false, false)
				nestedObject := tfProvider.ResourcesMap[resourceName].Schema["object_nested_scheme_property"]
				assertTerraformSchemaProperty(t, nestedObject.Elem.(*schema.Resource).Schema["name"], schema.TypeString, false, true)
				assertTerraformSchemaProperty(t, nestedObject.Elem.(*schema.Resource).Schema["object_property"], schema.TypeList, false, false)
				object := nestedObject.Elem.(*schema.Resource).Schema["object_property"]
				assertTerraformSchemaProperty(t, object.Elem.(*schema.Resource).Schema["account"], schema.TypeString, false, false)
				assertTerraformSchemaProperty(t, object.Elem.(*schema.Resource).Schema["read_only"], schema.TypeString, false, true)

				So(tfProvider.ResourcesMap[resourceName].CreateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].ReadContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].UpdateContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].DeleteContext, ShouldNotBeNil)
				So(tfProvider.ResourcesMap[resourceName].Importer, ShouldNotBeNil)

				So(tfProvider.ConfigureFunc, ShouldNotBeNil)
			})
		})
	})

	Convey("Given a ProviderOpenAPI with an err field populated (this only happens once the provider has been initialized and the resulted provider/error have been cached in the ProviderOpenAPI", t, func() {
		p := ProviderOpenAPI{ProviderName: providerName, err: errors.New("some error")}
		Convey("When CreateSchemaProviderWithConfiguration method is called", func() {
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{})
			Convey("Then the error returned should be the expecting one", func() {
				So(err.Error(), ShouldEqual, "some error")
			})
			Convey("And the tfProvider returned should be nil ", func() {
				So(tfProvider, ShouldBeNil)
			})
		})
	})

	Convey("Given a ProviderOpenAPI with an provider field populated (this only happens once the provider has been initialized and the resulted provider/error have been cached in the ProviderOpenAPI", t, func() {
		cachedProvider := &schema.Provider{}
		p := ProviderOpenAPI{ProviderName: providerName, provider: cachedProvider}
		Convey("When CreateSchemaProviderWithConfiguration method is called", func() {
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{})
			Convey("Then the error returned should be the nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the tfProvider returned should be the cached one ", func() {
				So(tfProvider, ShouldEqual, cachedProvider)
			})
		})
	})

	Convey("Given a ProviderOpenAPI", t, func() {
		swaggerContent := `swagger: "2.0"
host: "localhost:8443"
basePath: "/api"
schemes:
- "https"
paths:
  /v1/cdns:
    x-terraform-resource-name: "cdn"
    post:
      summary: "Create cdn"
      
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
        type: "string"`

		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(swaggerContent))
		}))
		p := ProviderOpenAPI{ProviderName: providerName}
		Convey("When CreateSchemaProviderWithConfiguration method is called with a serviceConfiguration that has InsecureSkipVerifyEnabled set to true", func() {
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerServer.URL, InsecureSkipVerify: true})
			Convey("Then the error returned should be the nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the tfProvider returned should not be nil (skipping schema verification since that's covered in other tests)", func() {
				So(tfProvider, ShouldNotBeNil)
			})
			Convey("And default TLS transport configuration should be configured to skip verify", func() {
				tr := http.DefaultTransport.(*http.Transport)
				So(tr.TLSClientConfig.InsecureSkipVerify, ShouldBeTrue)
			})

		})
	})

	Convey("Given a ProviderOpenAPI missing the providerName", t, func() {
		swaggerContent := `swagger: "2.0"`
		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(swaggerContent))
		}))
		p := ProviderOpenAPI{}
		Convey("When CreateSchemaProviderWithConfiguration method is called", func() {
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerServer.URL})
			Convey("Then the error returned should be the nil", func() {
				So(err.Error(), ShouldEqual, "plugin provider factory init error: provider name not specified")
			})
			Convey("And the tfProvider returned should be nil", func() {
				So(tfProvider, ShouldBeNil)
			})
		})
	})

}

func Test_colliding_resource_names(t *testing.T) {
	makeSwaggerDoc := func(path1, preferredName1, path2, preferredName2 string, markIgnorePath1 bool) string {
		if path1 == "" {
			path1 = "/v1/abc"
		}
		if path2 == "" {
			path2 = "/v1/xyz"
		}
		xTerraformResourceName1 := ""
		if preferredName1 != "" {
			xTerraformResourceName1 = `x-terraform-resource-name: "` + preferredName1 + `"`
		}
		xTerraformResourceName2 := ""
		if preferredName2 != "" {
			xTerraformResourceName2 = `x-terraform-resource-name: "` + preferredName2 + `"`
		}
		ignorePath1 := ""
		if markIgnorePath1 {
			ignorePath1 = `x-terraform-exclude-resource: true`
		}
		swagger := `swagger: "2.0"
paths:
  ` + path1 + `:
    post:
      ` + xTerraformResourceName1 + `
      ` + ignorePath1 + `
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/whatever"
      responses:
        201:
          schema:
            $ref: "#/definitions/whatever"
  ` + path1 + `/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/whatever"

  ` + path2 + `:
    post:
      ` + xTerraformResourceName2 + `
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/whatever"
      responses:
        201:
          schema:
            $ref: "#/definitions/whatever"
  ` + path2 + `/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/whatever"

definitions:
  whatever:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true`
		return swagger
	}

	swaggerDocServerURL := func(swaggerDoc string) (serverURL string) {
		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(swaggerDoc))
		}))
		return swaggerServer.URL
	}

	Convey("Given a ProviderOpenAPI and different swagger docs that declare resources with colliding names", t, func() {
		p := ProviderOpenAPI{ProviderName: "something"}

		testcases := []struct {
			label             string
			path1             string
			preferredName1    string
			path2             string
			preferredName2    string
			expectedResources int
			expectedWarning   string
		}{
			{
				label:             "resources with colliding x-terraform-resource-names",
				preferredName1:    "collision",
				preferredName2:    "collision",
				expectedResources: 0,
				expectedWarning:   "'collision_v1' is a duplicate resource name and is being removed from the provider"},
			{
				label:             "resources with colliding x-terraform-resource-name calculated name and calculated versioned name",
				path1:             "/v1/collision",
				path2:             "/xyz",
				preferredName2:    "collision_v1",
				expectedResources: 0,
				expectedWarning:   "'collision_v1' is a duplicate resource name and is being removed from the provider"},
			{
				label:             "resources with colliding x-terraform-resource-name calculated preferred name and calculated versioned name",
				path1:             "/v1/collision",
				path2:             "/v1/xyz",
				preferredName2:    "collision",
				expectedResources: 0,
				expectedWarning:   "'collision_v1' is a duplicate resource name and is being removed from the provider"},
			{
				label:             "resources with colliding calculated names",
				path1:             "/v1/collision",
				path2:             "/collision_v1",
				expectedResources: 0,
				expectedWarning:   "'collision_v1' is a duplicate resource name and is being removed from the provider"},
			{
				label:             "resources with identical paths and a colliding preferred name",
				path1:             "/v1/collision",
				path2:             "/v1/collision",
				preferredName1:    "collision_v1",
				expectedResources: 1,
				expectedWarning:   "",
			},
			{
				label:             "resources with identical paths and identical preferred names",
				path1:             "/v1/collision",
				path2:             "/v1/collision",
				preferredName1:    "collision",
				preferredName2:    "collision",
				expectedResources: 1,
				expectedWarning:   "",
			},
		}

		for _, tc := range testcases {
			out := newTestWriter()
			log.SetOutput(out)
			swaggerDoc := makeSwaggerDoc(tc.path1, tc.preferredName1, tc.path2, tc.preferredName2, false)
			fmt.Println(">>>", swaggerDoc)
			Convey(fmt.Sprintf("When CreateSchemaProviderFromServiceConfiguration method is called: %s", tc.label), func() {
				tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerDocServerURL(swaggerDoc)})
				Convey("Then the result returned should be the expected one", func() {
					So(err, ShouldBeNil)
					So(len(tfProvider.ResourcesMap), ShouldEqual, tc.expectedResources)
					if tc.expectedWarning == "" {
						So(out.written, ShouldNotContainSubstring, "duplicate resource name")
					} else {
						So(out.written, ShouldContainSubstring, tc.expectedWarning)
					}
				})
			})
		}
	})

	Convey("Given a ProviderOpenAPI and a swagger doc that declares resources with colliding names, and all but one is ignored", t, func() {
		out := newTestWriter()
		log.SetOutput(out)
		p := ProviderOpenAPI{ProviderName: "something"}
		swaggerDoc := makeSwaggerDoc("/v1/abc", "collision", "/v1/xyz", "collision", true)
		Convey("When CreateSchemaProviderWithConfiguration is called", func() {
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerDocServerURL(swaggerDoc)})
			Convey("Then there should be no error and the provider should have the un-ignored resource and no warning should be logged", func() {
				So(err, ShouldBeNil)
				So(len(tfProvider.ResourcesMap), ShouldEqual, 1)
				So(out.written, ShouldNotContainSubstring, "duplicate resource name")
				So(out.written, ShouldContainSubstring, "is marked to be ignored")
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
			Convey("Then the service configuration swagger URL should be the expected one, the service configuration should be false and error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(serviceConfiguration.GetSwaggerURL(), ShouldEqual, expectedSwaggerURL)
				So(serviceConfiguration.IsInsecureSkipVerifyEnabled(), ShouldBeFalse)
			})
		})
	})
}

type logWriter struct {
	written string
}

func newTestWriter() *logWriter {
	return &logWriter{""}
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	w.written = w.written + string(p)
	return 0, nil
}
