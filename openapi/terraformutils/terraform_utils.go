package terraformutils

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"strings"
)

const terraformPluginVendorDir = "terraform.d/plugins"

// TerraformUtils defines a struct that exposes some handy terraform utils functions
type TerraformUtils struct {
	Runtime string
}

// GetTerraformPluginsVendorDir returns Terraform's global plugin vendor directory where Terraform suggests installing
// custom plugins such as the OopenAPI Terraform provider. This function supports the most used platforms including
// windows, darwin and linux
func (t *TerraformUtils) GetTerraformPluginsVendorDir() (string, error) {
	var terraformPluginsFolder string
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Printf("[ERROR] A failure occurred when getting the user's home directory. Error = %s", err)
		return "", err
	}
	// On all other systems, in the sub-path .terraform.d/plugins in your user's home directory.
	terraformPluginsFolder = fmt.Sprintf("%s/.%s", homeDir, terraformPluginVendorDir)
	// On Windows, in the sub-path terraform.d/plugins beneath your user's "Application Data" directory.
	if t.Runtime == "windows" {
		terraformPluginsFolder = fmt.Sprintf("%s/%s", homeDir, terraformPluginVendorDir)
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
	return MultiEnvDefaultFunc([]string{key}, defaultValue)

}

// MultiEnvDefaultFunc is a helper function that returns the value of the first
// environment variable in the given list 'ks' that returns a non-empty value. If none of the environment variables
// return a value, the default value is
// returned.
func MultiEnvDefaultFunc(ks []string, defaultValue interface{}) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		for _, k := range ks {
			if v := os.Getenv(k); v != "" {
				return v, nil
			}
		}
		return defaultValue, nil
	}
}
