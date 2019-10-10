package e2e

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/stretchr/testify/assert"
)

func TestAcc_RefreshTokenSecurityDefinition(t *testing.T) {

	refreshTokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		refreshToken := r.Header.Get("Authorization")
		assert.NotEmpty(t, refreshToken)
		assert.Contains(t, refreshToken, "Bearer")
		assert.Contains(t, refreshToken, "refreshTokenValue")
		w.Header().Set("Authorization", "Bearer some_access_token")
	}))

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken := r.Header.Get("Authorization")
		assert.NotEmpty(t, accessToken)
		assert.Equal(t, accessToken, "Bearer some_access_token")
		w.Write([]byte(`{"id":"someID", "label": "my_label"}`))
	}))
	apiHost := apiServer.URL[7:]

	swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		swaggerYAMLTemplate := fmt.Sprintf(`swagger: "2.0"
host: "%s"
basePath: "/api"

schemes:
- "http"

security:
  - apikey_auth: []

securityDefinitions:
  apikey_auth:
    x-terraform-refresh-token-url: "%s"
    type: "apiKey"
    name: "Authorization"
    in: "header"

paths:
  /v1/cdns:
    post:
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
    delete: 
      operationId: ContentDeliveryNetworkFirewallDeleteV1
      parameters: 
        - description: "The cdn id that contains the firewall to be fetched."
          in: path
          name: parent_id
          required: true
          type: string
        - description: "The cdn firewall id that needs to be fetched."
          in: path
          name: id
          required: true
          type: string
      responses: 
        204: 
          description: "successful operation, no content is returned"
      summary: "Delete firewall"
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
        type: "string"`, apiHost, refreshTokenServer.URL)
		w.Write([]byte(swaggerYAMLTemplate))
	}))

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{
		SwaggerURL: swaggerServer.URL,
		SchemaConfiguration: []*openapi.ServiceSchemaPropertyConfigurationStub{
			{
				SchemaPropertyName: "apikey_auth",
				DefaultValue:       "refreshTokenValue",
			},
		},
	})
	assert.NoError(t, err)

	tfFileContents := fmt.Sprintf(`
resource "openapi_cdn_v1" "my_cdn_v1" {
	label = "my_label"
}`)

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
	resource.Test(t, resource.TestCase{
		IsUnitTest: true,
		Providers:  testAccProviders,
		PreCheck:   func() { testAccPreCheck(t, swaggerServer.URL) },
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn_v1", "id", "someID"),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn_v1", "label", "my_label"),
				),
			},
		},
	})
}
