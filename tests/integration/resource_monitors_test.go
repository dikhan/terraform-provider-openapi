package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const resourceNameMonitor = "monitors_v1"

var regionRst1 = "rst1"
var regionDub1 = "dub1"

var openAPIResourceNameMonitorRst = fmt.Sprintf("%s_%s_%s", providerName, resourceNameMonitor, regionRst1)
var openAPIResourceNameMonitorDub = fmt.Sprintf("%s_%s_%s", providerName, resourceNameMonitor, regionDub1)
var openAPIResourceInstanceNameMonitor = "my_monitor"

var testCreateConfigMonitor string
var testCreateConfigMonitorMultiRegionProvider string

func init() {
	testCreateConfigMonitor = populateTemplateConfigurationMonitor()
}

func TestAccMonitor_CreateRst1(t *testing.T) {
	expectedValidationError, _ := regexp.Compile(".*unable to unmarshal response body \\['invalid character '<' looking for beginning of value'\\] for request = 'POST https://some\\.api\\.rst1\\.domain\\.com/v1/monitors HTTP/1\\.1'\\. Response = '404 Not Found'.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testCreateConfigMonitor,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccMonitor_CreateDub1(t *testing.T) {
	expectedValidationError, _ := regexp.Compile(".*unable to unmarshal response body \\['invalid character '<' looking for beginning of value'\\] for request = 'POST https://some\\.api\\.dub1\\.domain\\.com/v1/monitors HTTP/1\\.1'\\. Response = '404 Not Found'.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testCreateConfigMonitor,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: expectedValidationError,
			},
		},
	})
}

func TestAccMonitor_MultiRegion_CreateRst1(t *testing.T) {
	testCreateConfigMonitor = populateTemplateConfigurationMonitorServiceProvider("rst1")
	expectedValidationError, _ := regexp.Compile(".*unable to unmarshal response body \\['invalid character '<' looking for beginning of value'\\] for request = 'POST https://some\\.api\\.rst1\\.domain\\.com/v1/multiregionmonitors HTTP/1\\.1'\\. Response = '404 Not Found'.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
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
	expectedValidationError, _ := regexp.Compile(".*unable to unmarshal response body \\['invalid character '<' looking for beginning of value'\\] for request = 'POST https://some\\.api\\.dub1\\.domain\\.com/v1/multiregionmonitors HTTP/1\\.1'\\. Response = '404 Not Found'.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
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

	expectedValidationError, _ := regexp.Compile(".*unable to unmarshal response body \\['invalid character '<' looking for beginning of value'\\] for request = 'POST https://some\\.api\\.rst1\\.domain\\.com/v1/multiregionmonitors HTTP/1\\.1'\\. Response = '404 Not Found'.*")
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testCreateConfigMonitor,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: expectedValidationError,
			},
		},
	})
}

func populateTemplateConfigurationMonitor() string {
	return fmt.Sprintf(`provider "%s" {
  apikey_auth = "apiKeyValue"
  x_request_id = "some value..."
}

resource "%s" "%s" {
  name = "someName"
}

resource "%s" "%s" {
  name = "someName"
}`, providerName, openAPIResourceNameMonitorRst, openAPIResourceInstanceNameMonitor, openAPIResourceNameMonitorDub, openAPIResourceInstanceNameMonitor)
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
