package terraformutils

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/mitchellh/go-homedir"
	"log"
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
	log.Printf("[DEBUG] ConvertToTerraformCompliantName - originalName = %s; compliantName = %s)", name, compliantName)
	return compliantName
}
