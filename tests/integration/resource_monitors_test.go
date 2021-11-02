package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var openAPIResourceInstanceNameMonitor = "my_monitor"

var testCreateConfigMonitor string

func TestAccMonitor_MultiRegion_CreateRst1(t *testing.T) {
	testCreateConfigMonitor = populateTemplateConfigurationMonitorServiceProvider("rst1")
	expectedValidationError, _ := regexp.Compile(".*request POST https://some.api.rst1.nonexistingrandomdomain.io/v1/multiregionmonitors HTTP/1.1 failed. Response Error: 'Post \"https://some.api.rst1.nonexistingrandomdomain.io/v1/multiregionmonitors\": dial tcp: lookup some.api.rst1.nonexistingrandomdomain.io.*")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testCreateConfigMonitor,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccMonitor_MultiRegion_CreateDub1(t *testing.T) {
	testCreateConfigMonitor = populateTemplateConfigurationMonitorServiceProvider("dub1")
	expectedValidationError, _ := regexp.Compile(".*request POST https://some.api.dub1.nonexistingrandomdomain.io/v1/multiregionmonitors HTTP/1.1 failed. Response Error: 'Post \"https://some.api.dub1.nonexistingrandomdomain.io/v1/multiregionmonitors\": dial tcp: lookup some.api.dub1.nonexistingrandomdomain.io.*")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testCreateConfigMonitor,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: expectedValidationError,
			},
		},
	})
}

// TestAccMonitor_MultiRegion_Create_Default_Region tests the case where the user did not provide a value for region
// and the provider uses the default value set in the swagger file instead: x-terraform-provider-regions: "rst1,dub1"
func TestAccMonitor_MultiRegion_Create_Default_Region(t *testing.T) {
	testCreateConfigMonitor = fmt.Sprintf(`provider "%s" {
  apikey_auth = "apiKeyValue"
  x_request_id = "some value..."
}
resource "openapi_multiregionmonitors_v1" "%s" {
  name = "someName"
}`, providerName, openAPIResourceInstanceNameMonitor)

	expectedValidationError, _ := regexp.Compile(".*request POST https://some.api.rst1.nonexistingrandomdomain.io/v1/multiregionmonitors HTTP/1.1 failed. Response Error: 'Post \"https://some.api.rst1.nonexistingrandomdomain.io/v1/multiregionmonitors\": dial tcp: lookup some.api.rst1.nonexistingrandomdomain.io.*")
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config:      testCreateConfigMonitor,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: expectedValidationError,
			},
		},
	})
}

func populateTemplateConfigurationMonitorServiceProvider(region string) string {
	return fmt.Sprintf(`provider "%s" {
  apikey_auth = "apiKeyValue"
  x_request_id = "some value..."
  region = "%s"
}

resource "openapi_multiregionmonitors_v1" "%s" {
  name = "someName"
}`, providerName, region, openAPIResourceInstanceNameMonitor)
}
