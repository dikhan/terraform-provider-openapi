package integration

import (
	"fmt"
	"testing"

	"github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen/api/api"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/go-openapi/loads"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
)

const providerName = "openapi"
const resourceName = "cdn_v1"

var openAPIResourceName = fmt.Sprintf("%s_%s", providerName, resourceName)
var openAPIResourceInstanceName = "my_cdn"
var openAPIResourceState = fmt.Sprintf("%s.%s", openAPIResourceName, openAPIResourceInstanceName)

var cdn api.ContentDeliveryNetwork
var testCDNCreateConfig string

func init() {
	// Setting this up here as it is used by many different tests
	cdn = newContentDeliveryNetwork("someLabel", []string{"192.168.0.2"}, []string{"www.google.com"}, 10, 12.22, true)
	testCDNCreateConfig = populateTemplateConfiguration(cdn.Label, cdn.Ips, cdn.Hostnames, cdn.ExampleInt, cdn.ExampleNumber, cdn.ExampleBoolean)
}

func TestAccCDN_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExist(),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_int", fmt.Sprintf("%d", cdn.ExampleInt)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "better_example_number_field_name", floatToString(cdn.ExampleNumber)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_boolean", fmt.Sprintf("%v", cdn.ExampleBoolean)),
				),
			},
		},
	})
}

func TestAccCDN_CreateFailsDueToMissingMandatoryApiKeyAuth(t *testing.T) {
	testCDNCreateMissingAPIKeyAuthConfig := fmt.Sprintf(`provider "%s" {
  # apikey_auth = "apiKeyValue" simulating configuration that is missing the mandatory apikey_auth (commented out for the reference)
  x_request_id = "some value..."
}
resource "%s" "my_cdn" {}`, providerName, openAPIResourceName)

	expectedValidationError, _ := regexp.Compile(".*\"apikey_auth\": required field is not set.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testCDNCreateMissingAPIKeyAuthConfig,
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccCDN_CreateFailsDueToWrongAuthKeyValue(t *testing.T) {
	testCDNCreateWrongAPIKeyAuthConfig := fmt.Sprintf(`provider "%s" {
  apikey_auth = "This is not the key expected by the API to authenticate the client, it should be 'apiKeyValue'' :)"
  x_request_id = "some value..."
}
resource "%s" "my_cdn" {
  label = "%s"
  ips = ["%s"]
  hostnames = ["%s"]
}`, providerName, openAPIResourceName, cdn.Label, arrayToString(cdn.Ips), arrayToString(cdn.Hostnames))

	expectedValidationError, _ := regexp.Compile(".*{\"code\":\"401\", \"message\": \"unauthorized user\"}.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testCDNCreateWrongAPIKeyAuthConfig,
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccCDN_CreateFailsDueToRequiredPropertyMissing(t *testing.T) {
	testCDNCreateConfigMissingRequiredProperty := fmt.Sprintf(`provider "%s" {
  apikey_auth = "apiKeyValue"
  x_request_id = "some value..."
}
resource "%s" "my_cdn" {
  #label = "%s" # ==> Simulating required field is missing (commented out for the reference)
  ips = ["%s"]
  hostnames = ["%s"]
}`, providerName, openAPIResourceName, cdn.Label, arrayToString(cdn.Ips), arrayToString(cdn.Hostnames))

	expectedValidationError, _ := regexp.Compile(".*\"label\": required field is not set.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testCDNCreateConfigMissingRequiredProperty,
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccCDN_Update(t *testing.T) {
	var cdnUpdated = newContentDeliveryNetwork(cdn.Label, cdn.Ips, cdn.Hostnames, 14, 14.14, false)
	testCDNUpdatedConfig := populateTemplateConfiguration(cdnUpdated.Label, cdnUpdated.Ips, cdnUpdated.Hostnames, cdnUpdated.ExampleInt, cdnUpdated.ExampleNumber, cdnUpdated.ExampleBoolean)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExist(),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_int", fmt.Sprintf("%d", cdn.ExampleInt)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "better_example_number_field_name", floatToString(cdn.ExampleNumber)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_boolean", fmt.Sprintf("%v", cdn.ExampleBoolean)),
				),
			},
			{
				Config: testCDNUpdatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExist(),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "label", cdnUpdated.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.#", fmt.Sprintf("%d", len(cdnUpdated.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.0", arrayToString(cdnUpdated.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.#", fmt.Sprintf("%d", len(cdnUpdated.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.0", arrayToString(cdnUpdated.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_int", fmt.Sprintf("%d", cdnUpdated.ExampleInt)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "better_example_number_field_name", floatToString(cdnUpdated.ExampleNumber)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_boolean", fmt.Sprintf("%v", cdnUpdated.ExampleBoolean)),
				),
			},
		},
	})
}

func TestAccCDN_UpdateImmutableProperty(t *testing.T) {
	testCDNUpdatedImmutableConfig := populateTemplateConfiguration("label updated", cdn.Ips, cdn.Hostnames, cdn.ExampleInt, cdn.ExampleNumber, cdn.ExampleBoolean)
	expectedValidationError, _ := regexp.Compile(".*property label is immutable and therefore can not be updated. Update operation was aborted; no updates were performed.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExist(),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_int", fmt.Sprintf("%d", cdn.ExampleInt)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "better_example_number_field_name", floatToString(cdn.ExampleNumber)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_boolean", fmt.Sprintf("%v", cdn.ExampleBoolean)),
				),
			},
			{
				Config:      testCDNUpdatedImmutableConfig,
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccCDN_UpdateForceNewProperty(t *testing.T) {
	var cdnUpdatedForceNew = newContentDeliveryNetwork(cdn.Label, []string{"192.168.1.5"}, cdn.Hostnames, cdn.ExampleInt, cdn.ExampleNumber, cdn.ExampleBoolean)
	testCDNUpdatedForceNewConfig := populateTemplateConfiguration(cdnUpdatedForceNew.Label, cdnUpdatedForceNew.Ips, cdnUpdatedForceNew.Hostnames, cdnUpdatedForceNew.ExampleInt, cdnUpdatedForceNew.ExampleNumber, cdnUpdatedForceNew.ExampleBoolean)
	var originalID string
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfig,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						for _, res := range s.RootModule().Resources {
							if res.Type != openAPIResourceName {
								continue
							}
							originalID = res.Primary.ID
						}
						return nil
					},
					resource.TestCheckResourceAttr(
						openAPIResourceState, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_int", fmt.Sprintf("%d", cdn.ExampleInt)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "better_example_number_field_name", floatToString(cdn.ExampleNumber)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_boolean", fmt.Sprintf("%v", cdn.ExampleBoolean)),
				),
			},
			{
				Config: testCDNUpdatedForceNewConfig,
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						for _, res := range s.RootModule().Resources {
							if res.Type != openAPIResourceName {
								continue
							}
							// check that the ID generated in the first config apply has changed to a different one as the force new resource was required by the change applied
							forceNewID := res.Primary.ID
							if originalID == forceNewID {
								return fmt.Errorf("force new operation did not work, resource still has the same ID %s", originalID)
							}
						}
						resourceExistsFunc := testAccCheckResourceExist()
						return resourceExistsFunc(s)
					},
					resource.TestCheckResourceAttr(
						openAPIResourceState, "label", cdnUpdatedForceNew.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.#", fmt.Sprintf("%d", len(cdnUpdatedForceNew.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "ips.0", arrayToString(cdnUpdatedForceNew.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.#", fmt.Sprintf("%d", len(cdnUpdatedForceNew.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "hostnames.0", arrayToString(cdnUpdatedForceNew.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_int", fmt.Sprintf("%d", cdnUpdatedForceNew.ExampleInt)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "better_example_number_field_name", floatToString(cdnUpdatedForceNew.ExampleNumber)),
					resource.TestCheckResourceAttr(
						openAPIResourceState, "example_boolean", fmt.Sprintf("%v", cdnUpdatedForceNew.ExampleBoolean)),
				),
			},
		},
	})
}

func newContentDeliveryNetwork(label string, ips, hostnames []string, exampleInt int32, exampleNumber float32, exampleBool bool) api.ContentDeliveryNetwork {
	return api.ContentDeliveryNetwork{
		Label:          label,
		Ips:            ips,
		Hostnames:      hostnames,
		ExampleInt:     exampleInt,
		ExampleNumber:  exampleNumber,
		ExampleBoolean: exampleBool,
	}
}

func populateTemplateConfiguration(label string, ips, hostnames []string, exampleInt int32, exampleNumber float32, exampleBool bool) string {
	return fmt.Sprintf(`provider "%s" {
  apikey_auth = "apiKeyValue"
  x_request_id = "some value..."
}

resource "%s" "my_cdn" {
  label = "%s"
  ips = ["%s"]
  hostnames = ["%s"]

  example_int = %d
  better_example_number_field_name = %s
  example_boolean = %v
}`, providerName, openAPIResourceName, label, arrayToString(ips), arrayToString(hostnames), exampleInt, floatToString(exampleNumber), exampleBool)
}

func floatToString(number float32) string {
	return fmt.Sprintf("%.2f", number)
}

func arrayToString(value []string) string {
	var result = "["
	for _, v := range value {
		result += fmt.Sprintf("%s,", v)
	}
	result = strings.TrimRight(result, ",")
	result += "]"
	return result
}

// Check if resource exists remotely
func testAccCheckResourceExist() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := testCheckCDNsV1Destroy(s)
		if strings.Contains(err.Error(), "still exists") {
			return nil
		}
		return err
	}
}

// Acceptance test resource-destruction for openapi_cdns_v1:
//
// Check all CDNs specified in the configuration have been destroyed.
func testCheckCDNsV1Destroy(state *terraform.State) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != openAPIResourceName {
			continue
		}
		cdnID := res.Primary.ID
		openAPIClient := testAccProvider.Meta().(openapi.ClientOpenAPI)
		abs, err := filepath.Abs(exampleSwaggerFile)
		if err != nil {
			return err
		}
		apiSpec, err := loads.JSONSpec(abs)
		if err != nil {
			return err
		}

		specResource := &openapi.SpecV2Resource{
			Name:             resourceName,
			Path:             "/v1/cdns",
			SchemaDefinition: apiSpec.Spec().Definitions["ContentDeliveryNetworkV1"],
			InstancePathItem: apiSpec.Spec().Paths.Paths["/v1/cdns/{id}"],
			RootPathItem:     apiSpec.Spec().Paths.Paths["/v1/cdns"],
		}

		resp, err := openAPIClient.Get(specResource, cdnID, nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("cdn '%s' still exists", cdnID)
		}
	}
	return nil
}
