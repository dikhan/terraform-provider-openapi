package terraformutils

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/go-homedir"
	"os"
	"runtime"
	"strings"
)

// TerraformPluginVendorDir defines the location where Terraform plugins are installed as per Terraform documentation:
// https://www.terraform.io/docs/extend/how-terraform-works.html#discovery
// https://www.terraform.io/docs/configuration/providers.html#third-party-plugins
const TerraformPluginVendorDir = ".terraform.d/plugins"
// TerraformPluginVendorDirWindows defines the path under which third party terraform plugins are to be installed
const TerraformPluginVendorDirWindows = "AppData\\terraform.d\\plugins"

// TerraformUtils defines a struct that exposes some handy terraform utils functions
type TerraformUtils struct {
	// Platform defines the OS (darwin, linux, windows) depending on which Terraform Vendor dir paths will be built differently
	Platform string
	// HomeDir defines the user's home directory
	HomeDir string
}

// NewTerraformUtils is a handy constructor to build a TerraformUtils object with default platform and homeDir values
// based on the user's computer settings
func NewTerraformUtils() (*TerraformUtils, error) {
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("failure occurred when getting the user's home directory: %s", err)
	}
	return &TerraformUtils{
		Platform: runtime.GOOS,
		HomeDir: homeDir,
	}, nil
}

// GetTerraformPluginsVendorDir returns Terraform's global plugin vendor directory where Terraform suggests installing
// custom plugins such as the OopenAPI Terraform provider. This function supports the most used platforms including
// darwin, linux and windows.
func (t *TerraformUtils) GetTerraformPluginsVendorDir() (string, error) {
	var terraformPluginsFolder string
	if t.Platform == "" {
		return "", fmt.Errorf("mandatory platform information is missing")
	}
	if t.HomeDir == "" {
		return "", fmt.Errorf("mandatory HomeDir value missing")
	}
	// On all other systems, in the sub-path .terraform.d/plugins in your user's home directory.
	terraformPluginsFolder = fmt.Sprintf("%s/%s", t.HomeDir, TerraformPluginVendorDir)
	// On Windows, in the sub-path (%APPDATA%\terraform.d\plugins) beneath your user's "Application Data" directory.
	if t.Platform == "windows" {
		terraformPluginsFolder = fmt.Sprintf("%s\\%s", t.HomeDir, TerraformPluginVendorDirWindows)
	}
	return terraformPluginsFolder, nil
}

// ConvertToTerraformCompliantName will convert the input string into a terraform compatible field name following
// Terraform's snake case field name convention (lower case and snake case).
func ConvertToTerraformCompliantName(name string) string {
	compliantName := strcase.ToSnake(name)
	return compliantName
}

// createSchema creates a terraform schema configured based upon the parameters passed in
func createSchema(propertyName string, schemaType schema.ValueType, required bool, defaultValue string) *schema.Schema {
	s := &schema.Schema{
		Type: schemaType,
	}
	if defaultValue != "" {
		s.DefaultFunc = envDefaultFunc(propertyName, defaultValue)
	} else {
		s.DefaultFunc = envDefaultFunc(propertyName, nil)
	}
	if required {
		s.Required = true
	} else {
		s.Optional = true
	}
	return s
}

// CreateStringSchemaProperty creates a terraform schema of type string configured based upon the parameters passed in
func CreateStringSchemaProperty(propertyName string, required bool, defaultValue string) *schema.Schema {
	return createSchema(propertyName, schema.TypeString, required, defaultValue)
}

// envDefaultFunc is a helper function that returns the value of the first
// environment variable in the given list 'ks' that returns a non-empty value. The ks are converted to upper case
// automatically for convenience. If none of the environment variables return a value, the default value is
// returned.
func envDefaultFunc(ks string, defaultValue interface{}) schema.SchemaDefaultFunc {
	key := strings.ToUpper(ks)
	return multiEnvDefaultFunc([]string{key}, defaultValue)

}

// multiEnvDefaultFunc is a helper function that returns a schema.SchemaDefaultFunc. The function returned
// returns the first environment variable in the given list 'ks' that returns a non-empty value. If none of the
// environment variables return a value, the default value is returned.
func multiEnvDefaultFunc(ks []string, defaultValue interface{}) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		for _, k := range ks {
			if v := os.Getenv(k); v != "" {
				return v, nil
			}
		}
		return defaultValue, nil
	}
}

// MultiEnvDefaultString is a helper function that returns the value as string of the first
// environment variable in the given list 'ks' that returns a non-empty value. If none of the environment variables
// return a value, the default value is
// returned.
func MultiEnvDefaultString(ks []string, defaultValue interface{}) (string, error) {
	dv, err := multiEnvDefaultFunc(ks, defaultValue)()
	if err != nil {
		return "", err
	}
	return dv.(string), nil
}
