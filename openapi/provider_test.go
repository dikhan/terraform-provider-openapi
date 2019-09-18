package openapi

import (
	"fmt"
	"log"
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

	Convey("Given a local server that exposes a swagger file containing a terraform compatible data source (cdn_datasource)", t, func() {
		swaggerContent := `swagger: "2.0"
host: "localhost:8443"
basePath: "/api"

schemes:
- "https"

security:
  - apikey_auth: []

paths:
  /v1/cdn_datasource:
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
				Convey("the provider dataSource map should contain the cdn resource with the expected configuration", func() {
					So(tfProvider.DataSourcesMap, ShouldNotBeNil)
					So(len(tfProvider.DataSourcesMap), ShouldEqual, 1)

					resourceName := fmt.Sprintf("%s_cdn_datasource_v1", providerName)
					So(tfProvider.DataSourcesMap, ShouldContainKey, resourceName)
					Convey("the provider cdn resource should have the expected schema", func() {
						resourceName := fmt.Sprintf("%s_cdn_datasource_v1", providerName)
						So(tfProvider.DataSourcesMap, ShouldContainKey, resourceName)

						So(tfProvider.DataSourcesMap[resourceName].Schema, ShouldContainKey, "label")
						So(tfProvider.DataSourcesMap[resourceName].Schema["label"].Type, ShouldEqual, schema.TypeString)
						So(tfProvider.DataSourcesMap[resourceName].Schema["label"].Required, ShouldBeFalse)
						So(tfProvider.DataSourcesMap[resourceName].Schema["label"].Optional, ShouldBeTrue)
						So(tfProvider.DataSourcesMap[resourceName].Schema["label"].Computed, ShouldBeTrue)

						So(tfProvider.DataSourcesMap[resourceName].Schema, ShouldContainKey, "filter")
						So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Type, ShouldEqual, schema.TypeSet)
						So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Required, ShouldBeFalse)
						So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Optional, ShouldBeTrue)
						So(tfProvider.DataSourcesMap[resourceName].Schema["filter"].Computed, ShouldBeFalse)

						elements := tfProvider.DataSourcesMap[resourceName].Schema["filter"].Elem.(*schema.Resource).Schema
						So(elements["name"].Type, ShouldEqual, schema.TypeString)
						So(elements["values"].Type, ShouldEqual, schema.TypeList)
					})
					Convey("the provider cdn-datasource data source should have only the READ operation configured", func() {
						So(tfProvider.DataSourcesMap[resourceName].Read, ShouldNotBeNil)
					})

					Convey("and the provider resource map must be nil as no resources are configured in the swagger", func() {
						So(tfProvider.ResourcesMap, ShouldBeEmpty)
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

	Convey("Given a swagger doc that declares resources with colliding names, "+
		"When CreateSchemaProviderWithConfiguration is called, "+
		"Then there should be no error but the provider should not have those resources and a warning should be logged", t, func() {

		testcases := []struct {
			label           string
			path1           string
			preferredName1  string
			path2           string
			preferredName2  string
			expectedWarning string
		}{
			{label: "resources with colliding x-terraform-resource-names",
				preferredName1:  "collision",
				preferredName2:  "collision",
				expectedWarning: "'collision_v1' is a duplicate resource name and is being removed from the provider"},
			{label: "resources with colliding x-terraform-resource-name calculated name and calculated versioned name",
				path1:           "/v1/collision",
				path2:           "/xyz",
				preferredName2:  "collision_v1",
				expectedWarning: "'collision_v1' is a duplicate resource name and is being removed from the provider"},
			{label: "resources with colliding x-terraform-resource-name calculated name and calculated versioned name",
				path1:           "/v1/collision",
				path2:           "/v1/xyz",
				preferredName2:  "collision",
				expectedWarning: "'collision_v1' is a duplicate resource name and is being removed from the provider"},
			{label: "resources with colliding calculated names",
				path1:           "/v1/collision",
				path2:           "/collision_v1",
				expectedWarning: "'collision_v1' is a duplicate resource name and is being removed from the provider"},
		}

		for _, tc := range testcases {
			out := newTestWriter()
			log.SetOutput(out)

			p := ProviderOpenAPI{ProviderName: "something"}
			swaggerDoc := makeSwaggerDoc(tc.path1, tc.preferredName1, tc.path2, tc.preferredName2, false)
			fmt.Println(">>>", swaggerDoc)
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerDocServerURL(swaggerDoc)})
			fmt.Println(">>>>", out.written)
			So(err, ShouldBeNil)
			So(len(tfProvider.ResourcesMap), ShouldEqual, 0)
			So(out.written, ShouldContainSubstring, tc.expectedWarning)
		}
	})

	Convey("Given a swagger doc that declares resources identical paths and colliding names preferred names, "+
		"When CreateSchemaProviderWithConfiguration is called, "+
		"Then there will be no error and the provider will have one of those resources (indeterminately selected) and no warning will be logged", t, func() {

		testcases := []struct {
			label          string
			path1          string
			preferredName1 string
			path2          string
			preferredName2 string
		}{
			{label: "resources with identical paths and a colliding preferred name",
				path1:          "/v1/collision",
				path2:          "/v1/collision",
				preferredName1: "collision_v1"},
			{label: "resources with identical paths and identical preferred names",
				path1:          "/v1/collision",
				path2:          "/v1/collision",
				preferredName1: "collision",
				preferredName2: "collision"},
		}

		for _, tc := range testcases {
			out := newTestWriter()
			log.SetOutput(out)

			p := ProviderOpenAPI{ProviderName: "something"}
			swaggerDoc := makeSwaggerDoc("/v1/collision", tc.preferredName1, "/v1/collision", tc.preferredName2, false)
			tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerDocServerURL(swaggerDoc)})

			So(err, ShouldBeNil)
			So(len(tfProvider.ResourcesMap), ShouldEqual, 1)
			So(out.written, ShouldNotContainSubstring, "duplicate resource name")
		}
	})

	Convey("Given a swagger doc that declares resources with colliding names, and all but one is ignored, "+
		"When CreateSchemaProviderWithConfiguration is called, "+
		"Then there should be no error and the provider should have the un-ignored resource and no warning should be logged", t, func() {

		out := newTestWriter()
		log.SetOutput(out)

		p := ProviderOpenAPI{ProviderName: "something"}
		swaggerDoc := makeSwaggerDoc("/v1/abc", "collision", "/v1/xyz", "collision", true)
		tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(&ServiceConfigStub{SwaggerURL: swaggerDocServerURL(swaggerDoc)})

		So(err, ShouldBeNil)
		So(len(tfProvider.ResourcesMap), ShouldEqual, 1)
		So(out.written, ShouldNotContainSubstring, "duplicate resource name")
		So(out.written, ShouldContainSubstring, "is marked to be ignored")

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
