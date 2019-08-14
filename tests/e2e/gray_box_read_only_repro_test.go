package e2e

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/stretchr/testify/assert"
)

const swaggerREADONLYpropertyContent = `swagger: "2.0"
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

  /v1/cdns/{cdn_id}:
    get:
      summary: "Get cdn by id"
      description: ""
      operationId: "ContentDeliveryNetworkGetV1"
      parameters:
      - name: "cdn_id"
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
    # delete:
    #   summary: "Delete cdn"
    #   operationId: "ContentDeliveryNetworkDeleteV1"
    #   parameters:
    #   - name: "id"
    #     in: "path"
    #     description: "The cdn that needs to be deleted"
    #     required: true
    #     type: "string"
    #   responses:
    #     204:
    #       description: "successful operation, no content is returned"

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
          name_that_is_readonly:
            type: "string"
            readOnly: true
          object_property:
            type: "object"
            properties:
              account:
                type: string
              a_read_read_only_property:
                type: string
                readOnly: true
`

// TODO: Given the above swagger we see a CDN resource with one nested property, which translate in the following TF config
//  # URI /v1/cdns/
//		resource "openapi_cdn_v1" "my_cdn" {
//		  label = "cdn"
//          object_nested_scheme_property {
//             name_that_is_readonly = "hello"
//             object_property = {
//                a_read_read_only_property = "whatever"
//             }
//          }
//		}   (see createTerraformFileREADONLY function)

// todo:  The test below consist of 3 steps, assuming that each step correspond to a Terraform apply command:
//      - Step 0 : we build the base terraform STATE file, taking care of building the nested object with the
//                 `a_read_read_only_property` = whatever

// todo:   State Step 0:
//      State: openapi_cdn_v1.my_cdn:
//          ID = 42
//          provider = provider.openapi
//          label =  cdn
//          object_nested_scheme_property.# = 1
//          object_nested_scheme_property.0.name_that_is_readonly = hello
//          object_nested_scheme_property.0.object_property.% = 1
//          object_nested_scheme_property.0.object_property.a_read_read_only_property = whatever

// todo:  - Step 1: given the TF state generated at Step 0, we apply a change to the above resource by changing the label name
//                from `cdn` to `updatedCDNLabel`. This is a change which occurs on the TOP level resource and
//                should not change the nested object ( see createTerraformFileREADONLY_UPDATE function)

// todo:   after this Apply i expect to see a state similar to:
//     State: openapi_cdn_v1.my_cdn:
//          ID = 42
//          provider = provider.openapi
//          label = updatedCDNLabel
//          object_nested_scheme_property.# = 1
//          object_nested_scheme_property.0.name_that_is_readonly = hello
//          object_nested_scheme_property.0.object_property.% = 2
//          object_nested_scheme_property.0.object_property.a_read_read_only_property = whatever
//          object_nested_scheme_property.0.object_property.account = im new here, but you should still see read_only : whatever

// todo:   INSTEAD what i get is :
//     State: openapi_cdn_v1.my_cdn:
//          ID = 42
//          provider = provider.openapi
//          label = updatedCDNLabel
//          object_nested_scheme_property.# = 1
//          object_nested_scheme_property.0.name_that_is_readonly = hello
//          object_nested_scheme_property.0.object_property.% = 1
//          object_nested_scheme_property.0.object_property.account = im new here, but you should still see read_only : whatever

// todo:     HENCE here is the BUG (IHMO) : line `object_nested_scheme_property.0.object_property.a_read_read_only_property = whatever` disappears
//        while it shioudl behave like line `object_nested_scheme_property.0.name = hello` whcih is still READ-ONLY but NOT definied in a nested object

// todo:    when applying Step 2, (using createTerraformFileREADONLY_UPDATE2 func) the  i get the following state:
//     State: openapi_cdn_v1.my_cdn:
//          ID = 42
//          provider = provider.openapi
//          label = updatedCDNLabel
//          object_nested_scheme_property.# = 1
//          object_nested_scheme_property.0.name = hello
//          object_nested_scheme_property.0.object_property.% = 0

//  todo: which * i think * correctly removes `object_nested_scheme_property.0.object_property.account ` as is it not defined anymore in the TF config file, but as well don't show
//    `object_nested_scheme_property.0.object_property.a_read_read_only_property = whatever` which is read only and must be there as
//    `object_nested_scheme_property.0.name = hello` does

func TestAccCDN_Create_and_UpdateSubResource_WITH_READONLY(t *testing.T) {
	api := initAPI(t, swaggerREADONLYpropertyContent)
	tfFileContents := createTerraformFileREADONLY(expectedCDNLabel)

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProviderFromServiceConfiguration(&openapi.ServiceConfigStub{SwaggerURL: api.swaggerURL})
	assert.NoError(t, err)
	assertProviderSchemaForReadONLY(t, provider)

	resourceInstancesToCheck := map[string]string{
		openAPIResourceNameCDN: fmt.Sprintf("%s/v1/cdns", api.apiHost),
	}

	var testAccProviders = map[string]terraform.ResourceProvider{providerName: provider}
	resource.Test(t, resource.TestCase{
		IsUnitTest:   true,
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t, api.swaggerURL) },
		CheckDestroy: testAccCheckDestroy(resourceInstancesToCheck),
		Steps: []resource.TestStep{
			{ // STEP 0 : create a Terraform State which has a Value for the readOnly property called `read_only`
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWhetherResourceExist(resourceInstancesToCheck),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", expectedCDNLabel),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.name", "hello"),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.object_property.read_only", "bub"),
				),
			},
			{ //STEP 1: Using a new configuration widtout readOnly property  keeps `read_only` = "bub"
				Config: createTerraformFileREADONLY_UPDATE("updatedCDNLabel"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWhetherResourceExist(resourceInstancesToCheck),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", "updatedCDNLabel"),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.name", "hello"),
					//resource.TestCheckResourceAttr(
					//	"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.object_property.read_only", "bub"),
					//resource.TestCheckResourceAttr(
					//	"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.object_property.account", "im new here, but you should still see read_only : bub"),
				),
			},
			{ //STEP 2: Using a new configuration widtout readOnly property  keeps `read_only` = "bub"
				Config: createTerraformFileREADONLY_UPDATE2("updatedCDNLabel"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWhetherResourceExist(resourceInstancesToCheck),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", "updatedCDNLabel"),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.name", "hello"),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.object_property.read_only", "bub"),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn", "object_nested_scheme_property.0.object_property.account", "im new here, but you should still see read_only : bub"),
				),
			},
		},
	})
}

func assertProviderSchemaForReadONLY(t *testing.T, provider *schema.Provider) {
	assert.Nil(t, provider.ResourcesMap[openAPIResourceNameCDN].Schema["id"])
	assert.NotNil(t, provider.ResourcesMap[openAPIResourceNameCDN].Schema["label"])
}

func createTerraformFileREADONLY(expectedCDNLabel string) string {
	return fmt.Sprintf(`
		# URI /v1/cdns/
		resource "%s" "%s" {
		  label = "%s"
          object_nested_scheme_property {
             name_that_is_readonly = "hello"
             object_property = {
               a_read_read_only_property = "whatever"
             }
          }
		}`, openAPIResourceNameCDN, openAPIResourceInstanceNameCDN, expectedCDNLabel)
}

func createTerraformFileREADONLY_UPDATE(expectedCDNLabel string) string {
	return fmt.Sprintf(`
		# URI /v1/cdns/
		resource "%s" "%s" {
          label = "%s"
          object_nested_scheme_property {
             object_property = {
                account = "im new here, but you should still see read_only : whatever"
             }
          }
		}`, openAPIResourceNameCDN, openAPIResourceInstanceNameCDN, expectedCDNLabel)
}

func createTerraformFileREADONLY_UPDATE2(expectedCDNLabel string) string {
	return fmt.Sprintf(`
		# URI /v1/cdns/
		resource "%s" "%s" {
          label = "%s"
          object_nested_scheme_property {
				object_property = {
             }
          }
		}`, openAPIResourceNameCDN, openAPIResourceInstanceNameCDN, expectedCDNLabel)
}
