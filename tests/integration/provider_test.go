package integration

import (
	"github.com/dikhan/terraform-provider-openapi/v2/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"os"
	"testing"

	"fmt"
)

var exampleSwaggerFile string

const providerName = "openapi"

var otfVarSwaggerURLEnvVariable = fmt.Sprintf("OTF_VAR_%s_SWAGGER_URL", providerName)
var otfVarSwaggerURLEnvVariableValue = "https://localhost:8443/swagger.yaml"
var otfVarInsecureSkipVerifyEnvVariable = "OTF_INSECURE_SKIP_VERIFY"

var testAccProvider = getAPIProviderInitWithPluginEnvVariables()
var testAccProviders = testAccProvidersFactory(testAccProvider)

func testAccProvidersFactory(provider *schema.Provider) map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		providerName: func() (*schema.Provider, error) {
			return provider, nil
		},
	}
}

func TestOpenAPIProvider(t *testing.T) {
	if err := testAccProvider.InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func getAPIProviderInitWithPluginEnvVariables() *schema.Provider {
	initPluginWithEnvironmentVariables()
	return getAPIProvider()
}

func getTestAccProviderInitWithPluginConfigurationFile(pluginConfigFile string) map[string]func() (*schema.Provider, error) {
	initPluginWithExternalConfigFile(pluginConfigFile)
	testAccProvider = getAPIProvider()
	return testAccProvidersFactory(testAccProvider)
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
