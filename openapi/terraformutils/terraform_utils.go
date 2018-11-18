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
func createSchema(propertyName string, schemaType schema.ValueType, required bool) *schema.Schema {
	s := &schema.Schema{
		Type:        schemaType,
		DefaultFunc: envDefaultFunc(propertyName, nil),
	}
	if required {
		s.Required = true
	} else {
		s.Optional = true
	}
	return s
}

// CreateStringSchema creates a terraform schema of type string configured based upon the parameters passed in
func CreateStringSchema(propertyName string, required bool) *schema.Schema {
	return createSchema(propertyName, schema.TypeString, required)
}

func envDefaultFunc(k string, defaultValue interface{}) schema.SchemaDefaultFunc {
	return func() (interface{}, error) {
		key := strings.ToUpper(k)
		if v := os.Getenv(key); v != "" {
			return v, nil
		}

		return defaultValue, nil
	}
}
