package i2

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

//const resourcePathCDN = "/v1/cdns"
//const resouceSchemaDefinitionNameCDN = "ContentDeliveryNetworkV1"

//const resourceCDNName = "cdn_v1"
//const resourceCDNName = "cdn_v1"

//var openAPIResourceNameCDN = fmt.Sprintf("%s_%s", providerName, resourceCDNName)
//var openAPIResourceInstanceNameCDN = "my_cdn"
//var openAPIResourceStateCDN = fmt.Sprintf("%s.%s", openAPIResourceNameCDN, openAPIResourceInstanceNameCDN)

//var cdn api.ContentDeliveryNetworkV1
//var testCreateConfigCDN string

const cdnSwaggerYAMLTemplate = `swagger: "2.0"

host: %s 
schemes:
- "http"

paths:
  ######################
  #### CDN Resource ####
  ######################

  /v1/cdns:
    post:
      x-terraform-resource-name: "cdn"
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV1"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkV1"

  /v1/cdns/{id}:
    get:
      summary: "Get cdn by id"
      description: ""
      operationId: "ContentDeliveryNetworkGetV1"
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

    put:
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
    delete:
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

  ## CDN sub-resource

  /v1/cdns/{parent_id}/v1/firewalls:
    post:
      summary: "Create cdn firewall"
      operationId: "ContentDeliveryNetworkFirewallCreateV1"
      parameters:
      - name: "parent_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - in: "body"
        name: "body"
        description: "Created CDN firewall"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"
      responses:
        201:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

  /v1/cdns/{parent_id}/v1/firewalls/{id}:
    get:
      summary: "Get cdn firewall by id"
      description: ""
      operationId: "ContentDeliveryNetworkFirewallGetV1"
      parameters:
      - name: "parent_id"
        in: "path"
        description: "The cdn id that contains the firewall to be fetched."
        required: true
        type: "string"
      - name: "id"
        in: "path"
        description: "The cdn firewall id that needs to be fetched."
        required: true
        type: "string"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"

definitions:
  ContentDeliveryNetworkFirewallV1:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
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
func (fakeServiceConfiguration) GetSchemaPropertyConfiguration(schemaPropertyName string) openapi.ServiceSchemaPropertyConfiguration {
	return fakeServiceSchemaPropertyConfiguration{}
}
func (fakeServiceConfiguration) Validate(runningPluginVersion string) error {
	return nil
}

func TestAccCDN_Subresource(t *testing.T) {
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

	tfFileContents := `# URI /v1/cdns/
resource "openapi_cdn_v1" "my_cdn" {
   label = "cdn label"
}

# URI /v1/cdns/{parent_id}/v1/firewalls/
resource "openapi_cdns_v1_firewalls_v1" "my_cdn_firewall_v1" {
   cdn_v1_id = "{openapi_cdn_v1.my_cdn.id}"
}`

	provider, e := openapi.CreateSchemaProviderFromServiceConfiguration(&openapi.ProviderOpenAPI{ProviderName: "openapi"}, fakeServiceConfiguration{
		getSwaggerURL: func() string {
			return swaggerServer.URL
		},
	})

	assert.NoError(t, e)

	assert.Nil(t, provider.ResourcesMap["openapi_cdn_v1"].Schema["id"]) //TODO: this needs to be not nil
	assert.NotNil(t, provider.ResourcesMap["openapi_cdn_v1"].Schema["label"])
	assert.Nil(t, provider.ResourcesMap["openapi_cdns_v1_firewalls_v1"].Schema["id"]) //TODO: this needs to be not nil
	assert.NotNil(t, provider.ResourcesMap["openapi_cdns_v1_firewalls_v1"].Schema["label"])
	assert.Nil(t, provider.ResourcesMap["openapi_cdns_v1_firewalls_v1"].Schema["cdn_v1_id"]) //TODO: this needs to be not nil

	var testAccProviders = map[string]terraform.ResourceProvider{"openapi": provider}

	resource.Test(t, resource.TestCase{
		IsUnitTest:   true,
		PreCheck:     nil,
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					//testAccCheckResourceExistCDN(),
					resource.TestCheckResourceAttr(
						"test_cdn_v1.my_cdn", "label", "cdn label"),
					resource.TestCheckResourceAttr(
						"test_cdns_v1_firewalls_v1.my_cdn_firewall_v1", "cdns_v1_id", "???"),
				),
			},
		},
	})

}
