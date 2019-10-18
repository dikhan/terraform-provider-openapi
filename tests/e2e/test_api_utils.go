package e2e

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func assertExpectedRequestURI(t *testing.T, expectedRequestURI string, r *http.Request) {
	if r.RequestURI != expectedRequestURI {
		assert.Fail(t, fmt.Sprintf("%s request URI '%s' does not match the expected one '%s'", r.Method, r.RequestURI, expectedRequestURI))
	}
}

func testAccCheckWhetherResourceExist(resourceInstancesToCheck map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for openAPIResourceName, resourceInstancePath := range resourceInstancesToCheck {
			resourceExistsInState := false
			for _, res := range s.RootModule().Resources {
				if res.Type != openAPIResourceName {
					continue
				}
				resourceExistsInState = true
				err := checkResourcesExist(resourceInstancePath, res.Primary.ID)
				if err != nil {
					return fmt.Errorf("API returned a non expected status code when checking if resource %s exists (GET %s/%s)", openAPIResourceName, resourceInstancePath, res.Primary.ID)
				}
			}
			if !resourceExistsInState {
				return fmt.Errorf("expected resource '%s' does not exist in the state file", openAPIResourceName)
			}
		}
		return nil
	}
}

func testAccCheckDestroy(resourceInstancesToCheck map[string]string) func(state *terraform.State) error {
	return func(s *terraform.State) error {
		for openAPIResourceName, resourceInstancePath := range resourceInstancesToCheck {
			resourceExistsInState := false
			for _, res := range s.RootModule().Resources {
				if res.Type != openAPIResourceName {
					continue
				}
				resourceExistsInState = true
				err := checkResourceIsDestroyed(resourceInstancePath, res.Primary.ID)
				if err != nil {
					return fmt.Errorf("API returned a non expected status code when checking if resource %s was destroy properly (GET %s/%s): %s", openAPIResourceName, resourceInstancePath, res.Primary.ID, err)
				}
			}
			if !resourceExistsInState {
				return fmt.Errorf("expected resource '%s' does not exist in the state file", openAPIResourceName)
			}
		}
		return nil
	}
}

func checkResourceIsDestroyed(resourceInstancePath string, resourceID string) error {
	err := checkResourcesExist(resourceInstancePath, resourceID)
	if err == nil {
		return fmt.Errorf("resource %s/%s still exists", resourceInstancePath, resourceID)
	}
	if !strings.Contains(err.Error(), "404") {
		return fmt.Errorf("GET %s/%s returned a non expected error: %s", resourceInstancePath, resourceID, err)
	}
	return nil
}

func checkResourcesExist(resourceInstancePath string, resourceID string) error {
	resourceInstanceURL := fmt.Sprintf("http://%s/%s", resourceInstancePath, resourceID)
	req, err := http.NewRequest(http.MethodGet, resourceInstanceURL, nil)
	if err != nil {
		return err
	}
	c := http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resource %s returned %d HTTP response code", resourceInstanceURL, resp.StatusCode)
	}
	return nil
}

func testAccPreCheck(t *testing.T, swaggerURLEndpoint string) {
	res, err := http.Get(swaggerURLEndpoint)
	if err != nil {
		t.Fatalf("error occurred when verifying if the API is up and running: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned not expected response status code %d", swaggerURLEndpoint, res.StatusCode)
	}
}
