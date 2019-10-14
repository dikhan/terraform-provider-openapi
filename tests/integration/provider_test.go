package integration

import (
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"os"
	"testing"

	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var exampleSwaggerFile string

const providerName = "openapi"

var otfVarSwaggerURLEnvVariable = fmt.Sprintf("OTF_VAR_%s_SWAGGER_URL", providerName)
var otfVarSwaggerURLEnvVariableValue = "https://localhost:8443/swagger.yaml"
var otfVarInsecureSkipVerifyEnvVariable = "OTF_INSECURE_SKIP_VERIFY"

var testAccProvider = getAPIProviderInitWithPluginEnvVariables()
var testAccProviders = map[string]terraform.ResourceProvider{providerName: testAccProvider}

func TestOpenAPIProvider(t *testing.T) {
	if err := testAccProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = testAccProvider
}

func getAPIProviderInitWithPluginEnvVariables() *schema.Provider {
	initPluginWithEnvironmentVariables()
	return getAPIProvider()
}

func getTestAccProviderInitWithPluginConfigurationFile(pluginConfigFile string) map[string]terraform.ResourceProvider {
	initPluginWithExternalConfigFile(pluginConfigFile)
	testAccProvider = getAPIProvider()
	return map[string]terraform.ResourceProvider{
		providerName: testAccProvider,
	}
}

func getAPIProvider() *schema.Provider {
	p := &openapi.ProviderOpenAPI{ProviderName: providerName}
	testAccProvider, err := p.CreateSchemaProvider()
	if err != nil {
		log.Fatalf("err: %s", err)
	}
	return testAccProvider
}

func initPluginWithEnvironmentVariables() {
	os.Setenv(otfVarSwaggerURLEnvVariable, otfVarSwaggerURLEnvVariableValue)
	os.Setenv(otfVarInsecureSkipVerifyEnvVariable, "true")
}

func initPluginWithExternalConfigFile(pluginConfigFile string) {
	os.Setenv(otfVarPluginConfigEnvVariableName, pluginConfigFile)
	// unset the other env variables to make sure the plugin will use the external config file
	os.Unsetenv(otfVarSwaggerURLEnvVariable)
	os.Unsetenv(otfVarInsecureSkipVerifyEnvVariable)
}
