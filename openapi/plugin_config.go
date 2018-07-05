package openapi

import (
	"bufio"
	"fmt"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const tfPluginsFolder = ".terraform.d/plugins"
const otfPluginConfigurationFileName = "terraform-provider-openapi.yaml"
const otfVarSwaggerURL = "OTF_VAR_%s_SWAGGER_URL"

// PluginConfiguration defines the OpenAPI plugin's configuration
type PluginConfiguration struct {
	// ProviderName defines the <provider_name> (should match the provider name of the terraform provider binary; terraform-provider-<provider_name>)
	ProviderName string
	// Configuration defines the reader that contains the plugin's external configuration (located at ~/.terraform.d/plugins)
	// If the plugin configuration file is not present the OTF_VAR_<provider_name>_SWAGGER_URL environment variable will
	// be required when invoking the openapi provider.
	// If at runtime both the OTF_VAR_<provider_name>_SWAGGER_URL as well as the plugin configuration file are present
	// the former takes preference. This allows the user to override the url specified in the configuration file with
	// the value provided in the OTF_VAR_<provider_name>_SWAGGER_URL
	Configuration io.Reader
}

// NewPluginConfiguration creates a new PluginConfiguration
func NewPluginConfiguration(providerName string) (*PluginConfiguration, error) {
	var configurationFile io.Reader
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, err
	}
	configurationFilePath := fmt.Sprintf("%s/%s/%s", homeDir, tfPluginsFolder, otfPluginConfigurationFileName)
	if _, err := os.Stat(configurationFilePath); os.IsNotExist(err) {
		log.Printf("[INFO] open api plugin configuration not present at %s", configurationFilePath)
	} else {
		log.Printf("[INFO] found open api plugin configuration at %s", configurationFilePath)
		file, err := os.Open(configurationFilePath)
		if err != nil {
			return nil, err
		}
		configurationFile = bufio.NewReader(file)
	}
	return &PluginConfiguration{
		ProviderName:  providerName,
		Configuration: configurationFile,
	}, nil
}

func (p *PluginConfiguration) getServiceProviderSwaggerURL() (string, error) {
	otfVarSwaggerURLLC := fmt.Sprintf(otfVarSwaggerURL, p.ProviderName)
	apiDiscoveryURL := os.Getenv(otfVarSwaggerURLLC)
	if apiDiscoveryURL != "" {
		log.Printf("[INFO] %s set with value %s", otfVarSwaggerURLLC, apiDiscoveryURL)
		return apiDiscoveryURL, nil
	}

	// Falling back to upper case check
	otfVarSwaggerURLUC := fmt.Sprintf(otfVarSwaggerURL, strings.ToUpper(p.ProviderName))
	apiDiscoveryURL = os.Getenv(otfVarSwaggerURLUC)
	if apiDiscoveryURL == "" {
		// Falling back to read from plugin configuration reader
		if p.Configuration == nil {
			return "", fmt.Errorf("swagger url not provided, please export %s or %s env variable with the URL where '%s' service provider is exposing the swagger file; or provide a %s configuration file with the provider swagger details", otfVarSwaggerURLUC, otfVarSwaggerURLLC, p.ProviderName, otfPluginConfigurationFileName)
		}
		source, err := ioutil.ReadAll(p.Configuration)
		if err != nil {
			return "", fmt.Errorf("failed to read %s configuration file", otfPluginConfigurationFileName)
		}
		var pluginConfigV1 = PluginConfigSchemaV1{}
		err = yaml.Unmarshal(source, &pluginConfigV1)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshall %s configuration file - error = %s", otfPluginConfigurationFileName, err)
		}
		pluginConfig := PluginConfigSchema(pluginConfigV1)
		if err = pluginConfig.Validate(); err != nil {
			return "", fmt.Errorf("error occurred while validating %s - error = %s", otfPluginConfigurationFileName, err)
		}
		apiDiscoveryURL, err = pluginConfig.GetProviderURL(p.ProviderName)
		return apiDiscoveryURL, err
	}

	if apiDiscoveryURL == "" {
		return "", fmt.Errorf("swagger url not provided, please export %s or %s env variable with the URL where '%s' service provider is exposing the swagger file", otfVarSwaggerURLUC, otfVarSwaggerURLLC, p.ProviderName)
	}
	log.Printf("[INFO] %s set with value %s", otfVarSwaggerURLUC, apiDiscoveryURL)
	return apiDiscoveryURL, nil
}
