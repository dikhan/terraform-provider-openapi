package integration

import (
	"fmt"
	"testing"

	"github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen/api/api"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

const resourceNameLB = "lbs_v1"
const resourcePathLB = "/v1/lbs"
const resouceSchemaDefinitionNameLB = "LB"

var openAPIResourceNameLB = fmt.Sprintf("%s_%s", providerName, resourceNameLB)
var openAPIResourceInstanceNameLB = "my_lb"
var openAPIResourceStateLB = fmt.Sprintf("%s.%s", openAPIResourceNameLB, openAPIResourceInstanceNameLB)

var lb api.Lbv1
var testCreateConfigLB string

func init() {
	// Setting this up here as it is used by many different tests
	lb = newLB("some_name", []string{"backend.com"}, 1, false)
	testCreateConfigLB = populateTemplateConfigurationLB(lb.Name, lb.Backends, lb.TimeToProcess, lb.SimulateFailure)
}

func TestAccLB_Create(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckLBsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCreateConfigLB,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExistLBs(),
					resource.TestCheckResourceAttr(
						openAPIResourceStateLB, "name", lb.Name),
					resource.TestCheckResourceAttr(
						openAPIResourceStateLB, "backends.#", fmt.Sprintf("%d", len(lb.Backends))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateLB, "backends.0", arrayToString(lb.Backends)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateLB, "time_to_process", fmt.Sprintf("%d", lb.TimeToProcess)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateLB, "simulate_failure", fmt.Sprintf("%v", lb.SimulateFailure)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateLB, "status", "deploy_pending"),
				),
			},
		},
	})
}

func newLB(name string, backend []string, timeToProcess int32, simulateFailure bool) api.Lbv1 {
	return api.Lbv1{
		Name:            name,
		Backends:        backend,
		TimeToProcess:   timeToProcess,
		SimulateFailure: simulateFailure,
	}
}

func populateTemplateConfigurationLB(name string, backend []string, timeToProcess int32, simulateFailure bool) string {
	return fmt.Sprintf(`provider "%s" {
  apikey_auth = "apiKeyValue"
  x_request_id = "some value..."
}

resource "%s" "%s" {
  name = "%s"
  backends = ["%s"]
  time_to_process = %d # the operation (post,update,delete) will take 15s in the API to complete
  simulate_failure = %v # no failures wished now ;) (post,update,delete)
}`, providerName, openAPIResourceNameLB, openAPIResourceInstanceNameLB, name, arrayToString(backend), timeToProcess, simulateFailure)
}

// Acceptance test resource-destruction for openapi_lbs_v1:
//
// Check all CDNs specified in the configuration have been destroyed.
func testCheckLBsV1Destroy(state *terraform.State) error {
	return testCheckDestroy(state, openAPIResourceNameLB, resourceNameLB, resourcePathLB, resouceSchemaDefinitionNameLB)
}

func testAccCheckResourceExistLBs() resource.TestCheckFunc {
	return testAccCheckResourceExist(openAPIResourceNameLB, resourceNameLB, resourcePathLB, resouceSchemaDefinitionNameLB)
}
