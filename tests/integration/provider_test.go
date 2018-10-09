package integration

import (
	"testing"

	"os"

	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"net/http"
)

var exampleSwaggerFile string

const providerName = "openapi"

var otfVarSwaggerURLEnvVariable = fmt.Sprintf("OTF_VAR_%s_SWAGGER_URL", providerName)
var otfVarInsecureSkipVerifyEnvVariable = "OTF_INSECURE_SKIP_VERIFY"
var otfVarSwaggerURLEnvVariableValue string
var otfVarInsecureSkipVerifyEnvVariableValue string

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	otfVarSwaggerURLEnvVariableValue = "https://localhost:8443/swagger.yaml"
	otfVarInsecureSkipVerifyEnvVariableValue = "true"
	os.Setenv(otfVarSwaggerURLEnvVariable, otfVarSwaggerURLEnvVariableValue)
	os.Setenv(otfVarInsecureSkipVerifyEnvVariable, otfVarInsecureSkipVerifyEnvVariableValue)

	testAccProvider = getAPIProvider()
	testAccProviders = map[string]terraform.ResourceProvider{
		providerName: testAccProvider,
	}
}

func TestOpenAPIProvider(t *testing.T) {
	if err := testAccProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = testAccProvider
}

func testAccPreCheck(t *testing.T) {
	if otfVarSwaggerURLEnvVariableValue == "" {
		t.Fatalf("env variable '%s' must be set for acceptance tests", otfVarSwaggerURLEnvVariable)
	}
	if otfVarInsecureSkipVerifyEnvVariableValue != "true" {
		t.Fatalf("env variable '%s' must be set to true for acceptance tests", otfVarInsecureSkipVerifyEnvVariable)
	}
	versionEndpoint := "https://localhost:8443/version"
	res, err := http.Get(versionEndpoint)
	if err != nil {
		t.Fatalf("error occured when verifying if the API is up and running: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned not expected response status code %d", versionEndpoint, res.StatusCode)
	}
}

func getAPIProvider() *schema.Provider {
	testAccProvider, err := openapi.APIProvider(providerName)
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	return testAccProvider
}
