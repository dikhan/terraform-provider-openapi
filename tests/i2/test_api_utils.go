package i2

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func assertExpectedRequestURI(t *testing.T, expectedRequestURI string, r *http.Request) {
	if r.RequestURI != expectedRequestURI {
		assert.Fail(t, fmt.Sprintf("%s request URI '%s' does not match the expected one '%s'", r.Method, r.RequestURI, expectedRequestURI))
	}
}

func testAccCheckWhetherResourceExist(resourceInstancesToCheck map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, res := range s.RootModule().Resources {
			if resourceInstanceURL, exists := resourceInstancesToCheck[res.Type]; exists {
				resourceID := res.Primary.ID

				resourceInstanceURL := fmt.Sprintf("http://%s/%s", resourceInstanceURL, resourceID)

				req, err := http.NewRequest(http.MethodGet, resourceInstanceURL, nil)
				if err != nil {
					return err
				}

				c := http.Client{}
				resp, err := c.Do(req)

				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("API returned a non expected status code %d when checking if resource %s exists (GET %s)", resp.StatusCode, res.Type, resourceInstanceURL)
				}
			}
		}
		return nil
	}
}
