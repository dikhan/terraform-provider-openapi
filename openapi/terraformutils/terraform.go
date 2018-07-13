package terraformutils

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"log"
	"runtime"
)

const terraformPluginVendorDir = "terraform.d/plugins"

// GetTerraformPluginsVendorDir returns Terraform's global plugin vendor directory where Terraform suggests installing
// custom plugins such as the OopenAPI Terraform provider. This function supports the most used platforms including
// windows, darwin and linux
func GetTerraformPluginsVendorDir() (string, error) {
	var terraformPluginsFolder string
	homeDir, err := homedir.Dir()
	if err != nil {
		log.Printf("[ERROR] A failure occurred when getting the user's home directory. Error = %s", err)
		return "", err
	}
	// On all other systems, in the sub-path .terraform.d/plugins in your user's home directory.
	terraformPluginsFolder = fmt.Sprintf("%s/.%s", homeDir, terraformPluginVendorDir)
	// On Windows, in the sub-path terraform.d/plugins beneath your user's "Application Data" directory.
	if runtime.GOOS == "windows" {
		terraformPluginsFolder = fmt.Sprintf("%s/%s", homeDir, terraformPluginVendorDir)
	}
	return terraformPluginsFolder, nil
}
