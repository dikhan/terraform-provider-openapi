package e2e

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func testAccProviders(provider *schema.Provider) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		providerName: func() (*schema.Provider, error) {
			return provider, nil
		},
	}
}

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
		if len(s.RootModule().Resources) == 0 {
			return nil
		}
		for openAPIResourceName, resourceInstancePath := range resourceInstancesToCheck {
			for _, res := range s.RootModule().Resources {
				if res.Type != openAPIResourceName {
					continue
				}
				err := checkResourceIsDestroyed(resourceInstancePath, res.Primary.ID)
				if err != nil {
					return fmt.Errorf("API returned a non expected status code when checking if resource %s was destroyed properly (GET %s/%s): %s", openAPIResourceName, resourceInstancePath, res.Primary.ID, err)
				}
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
	res, err := http.Get(swaggerURLEndpoint) // #nosec G107
	if err != nil {
		t.Fatalf("error occurred when verifying if the API is up and running: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned not expected response status code %d", swaggerURLEndpoint, res.StatusCode)
	}
}

func getFileContents(t *testing.T, filePath string) string {
	return string(getFileContentsBytes(t, filePath))
}

func getFileContentsBytes(t *testing.T, filePath string) []byte {
	fileContents, err := ioutil.ReadFile(filePath)
	assert.Nil(t, err)
	return fileContents
}
