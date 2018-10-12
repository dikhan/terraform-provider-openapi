package integration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

const resourceNameMonitor = "monitors_v1"

var regionRst1 = "rst1"
var regionDub1 = "dub1"

var openAPIResourceNameMonitorRst = fmt.Sprintf("%s_%s_%s", providerName, resourceNameMonitor, regionRst1)
var openAPIResourceNameMonitorDub = fmt.Sprintf("%s_%s_%s", providerName, resourceNameMonitor, regionDub1)
var openAPIResourceInstanceNameMonitor = "my_monitor"

var testCreateConfigMonitor string

func init() {
	testCreateConfigMonitor = populateTemplateConfigurationMonitor()
}

func TestAccMonitor_CreateRst1(t *testing.T) {
	expectedValidationError, _ := regexp.Compile(".*openapi_monitors_v1_rst1.my_monitor: unable to unmarshal response body \\['invalid character '<' looking for beginning of value'\\] for request = 'POST https://some\\.api\\.rst1\\.domain\\.com/v1/monitors HTTP/1\\.1'\\. Response = '404 Not Found'.*")
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
	expectedValidationError, _ := regexp.Compile(".*openapi_monitors_v1_dub1.my_monitor: unable to unmarshal response body \\['invalid character '<' looking for beginning of value'\\] for request = 'POST https://some\\.api\\.dub1\\.domain\\.com/v1/monitors HTTP/1\\.1'\\. Response = '404 Not Found'.*")
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
