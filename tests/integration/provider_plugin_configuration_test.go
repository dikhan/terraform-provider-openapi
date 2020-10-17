package integration

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v1/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
)

var otfVarPluginConfigEnvVariableName = fmt.Sprintf("OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE", providerName)

var testCDNCreateConfigWithoutProviderAuthProperty string

func init() {
	// Setting this up here as it is used by many different tests
	cdn = newContentDeliveryNetwork("someLabel", []string{"192.168.0.2"}, []string{"www.google.com"}, 10, 12.22, true, "someAccountValue", "some message news", "some more details", "http", 80, "https", 443)
	testCDNCreateConfigWithoutProviderAuthProperty = fmt.Sprintf(`provider "%s" {
  x_request_id = "some value..."
}
resource "%s" "my_cdn" {
  label = "%s"
  ips = ["%s"]
  hostnames = ["%s"]
}`, providerName, openAPIResourceNameCDN, cdn.Label, arrayToString(cdn.Ips), arrayToString(cdn.Hostnames))
}

func TestAccProviderConfiguration_EnvironmentVariable(t *testing.T) {
	os.Setenv("APIKEY_AUTH", "apiKeyValue")
	testCDNCreateConfigWithoutProviderAuthProperty := fmt.Sprintf(`provider "%s" {
  x_request_id = "some value..."
}
resource "%s" "my_cdn" {
  label = "%s"
  ips = ["%s"]
  hostnames = ["%s"]
}`, providerName, openAPIResourceNameCDN, cdn.Label, arrayToString(cdn.Ips), arrayToString(cdn.Hostnames))
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfigWithoutProviderAuthProperty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExistCDN(),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.#", fmt.Sprintf("%d", 1)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.0.name", "autogenerated name"),
				),
			},
		},
	})
	os.Unsetenv("APIKEY_AUTH")
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_DefaultValue(t *testing.T) {
	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    swagger-url: https://localhost:8443/swagger.yaml
    insecure_skip_verify: true
    schema_configuration:
    - schema_property_name: "apikey_auth"
      default_value: "apiKeyValue"`)
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: getTestAccProviderInitWithPluginConfigurationFile(file.Name()),
		CheckDestroy:      testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfigWithoutProviderAuthProperty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExistCDN(),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.#", fmt.Sprintf("%d", 1)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.0.name", "autogenerated name"),
				),
			},
		},
	})
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_ExternalConfiguration_Raw(t *testing.T) {
	apiKeyAuthFileContentRaw := "apiKeyValue"
	apiKeyAuthFile := createPluginConfigFile(apiKeyAuthFileContentRaw)
	defer os.Remove(apiKeyAuthFile.Name())

	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    swagger-url: https://localhost:8443/swagger.yaml
    insecure_skip_verify: true
    schema_configuration:
    - schema_property_name: "apikey_auth"
      schema_property_external_configuration:
        content_type: raw
        file: %s`, apiKeyAuthFile.Name())
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: getTestAccProviderInitWithPluginConfigurationFile(file.Name()),
		CheckDestroy:      testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfigWithoutProviderAuthProperty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExistCDN(),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.#", fmt.Sprintf("%d", 1)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.0.name", "autogenerated name"),
				),
			},
		},
	})
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_ExternalConfiguration_JSON(t *testing.T) {
	apiKeyAuthFileContentJSON := `{"apikey_auth":"apiKeyValue"}`
	apiKeyAuthFile := createPluginConfigFile(apiKeyAuthFileContentJSON)
	defer os.Remove(apiKeyAuthFile.Name())

	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    swagger-url: https://localhost:8443/swagger.yaml
    insecure_skip_verify: true
    schema_configuration:
    - schema_property_name: "apikey_auth"
      schema_property_external_configuration:
        content_type: json
        key_name: $.apikey_auth
        file: %s`, apiKeyAuthFile.Name())
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: getTestAccProviderInitWithPluginConfigurationFile(file.Name()),
		CheckDestroy:      testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfigWithoutProviderAuthProperty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExistCDN(),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.#", fmt.Sprintf("%d", 1)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.0.name", "autogenerated name"),
				),
			},
		},
	})
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_ExternalConfiguration_With_Successful_Command(t *testing.T) {
	apiKeyAuthFileContentJSON := `{"apikey_auth":"apiKeyValue"}`
	apiKeyAuthFile := createPluginConfigFile(apiKeyAuthFileContentJSON)
	defer os.Remove(apiKeyAuthFile.Name())

	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    swagger-url: https://localhost:8443/swagger.yaml
    insecure_skip_verify: true
    schema_configuration:
    - schema_property_name: "apikey_auth"
      cmd: ['date']
      schema_property_external_configuration:
        content_type: json
        key_name: $.apikey_auth
        file: %s`, apiKeyAuthFile.Name())
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: getTestAccProviderInitWithPluginConfigurationFile(file.Name()),
		CheckDestroy:      testCheckCDNsV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testCDNCreateConfigWithoutProviderAuthProperty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckResourceExistCDN(),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "label", cdn.Label),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.#", fmt.Sprintf("%d", len(cdn.Ips))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "ips.0", arrayToString(cdn.Ips)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.#", fmt.Sprintf("%d", len(cdn.Hostnames))),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "hostnames.0", arrayToString(cdn.Hostnames)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.#", fmt.Sprintf("%d", 1)),
					resource.TestCheckResourceAttr(
						openAPIResourceStateCDN, "object_nested_scheme_property.0.name", "autogenerated name"),
				),
			},
		},
	})
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_ExternalConfiguration_With_Non_Successful_Command(t *testing.T) {
	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    swagger-url: https://localhost:8443/swagger.yaml
    insecure_skip_verify: true
    schema_configuration:
    - schema_property_name: "apikey_auth"
      cmd: ["cat", "nonExistingFile"]`)
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	initPluginWithExternalConfigFile(file.Name())

	p := &openapi.ProviderOpenAPI{ProviderName: providerName}
	_, err := p.CreateSchemaProvider()
	assert.NoError(t, err)
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_ExternalConfiguration_With_Command_With_Timeout(t *testing.T) {
	testPluginConfig := fmt.Sprintf(`version: '1'
services:
 openapi:
   swagger-url: https://localhost:8443/swagger.yaml
   insecure_skip_verify: true
   schema_configuration:
   - schema_property_name: "apikey_auth"
     cmd: ["sleep", "2"]
     cmd_timeout: 1`)
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	initPluginWithExternalConfigFile(file.Name())
	p := &openapi.ProviderOpenAPI{ProviderName: providerName}
	_, err := p.CreateSchemaProvider()
	assert.NoError(t, err)
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_ExternalConfiguration_FileNotExists(t *testing.T) {
	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    swagger-url: https://localhost:8443/swagger.yaml
    insecure_skip_verify: true
    schema_configuration:
    - schema_property_name: "apikey_auth"
      cmd: ['date']
      schema_property_external_configuration:
        content_type: json
        key_name: $.apikey_auth
        file: /path_to_non_existing_file.json`)
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	initPluginWithExternalConfigFile(file.Name())
	p := &openapi.ProviderOpenAPI{ProviderName: providerName}
	_, err := p.CreateSchemaProvider()
	assert.NoError(t, err)
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_SchemaProperty_ExternalConfiguration_FileKeyNameWrong(t *testing.T) {
	apiKeyAuthFileContentJSON := `{"apikey_auth":"apiKeyValue"}`
	apiKeyAuthFile := createPluginConfigFile(apiKeyAuthFileContentJSON)
	defer os.Remove(apiKeyAuthFile.Name())

	testPluginConfig := fmt.Sprintf(`version: '1'
services:
  openapi:
    swagger-url: https://localhost:8443/swagger.yaml
    insecure_skip_verify: true
    schema_configuration:
    - schema_property_name: "apikey_auth"
      cmd: ['date']
      schema_property_external_configuration:
        content_type: json
        key_name: $.wrong_name
        file: %s`, apiKeyAuthFile.Name())
	file := createPluginConfigFile(testPluginConfig)
	defer os.Remove(file.Name())
	initPluginWithExternalConfigFile(file.Name())
	p := &openapi.ProviderOpenAPI{ProviderName: providerName}
	_, err := p.CreateSchemaProvider()
	assert.NoError(t, err)
	os.Unsetenv(otfVarPluginConfigEnvVariableName)
}

func TestAccProviderConfiguration_PluginExternalFile_NotFound(t *testing.T) {
	initPluginWithExternalConfigFile("/some/non/existing/path")
	p := &openapi.ProviderOpenAPI{ProviderName: providerName}
	_, err := p.CreateSchemaProvider()
	if !strings.Contains(err.Error(), "swagger url not provided, please export OTF_VAR_<provider_name>_SWAGGER_URL env variable with the URL where 'openapi' service provider is exposing the swagger file OR create a plugin configuration file at ~/.terraform.d/plugins following the Plugin configuration schema specifications") {
		log.Fatalf("test failed, non expected output: %s", err)
	}
}
