package terraformutils

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"

	"github.com/iancoleman/strcase"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mitchellh/go-homedir"
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
		HomeDir:  homeDir,
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

var numberInName = regexp.MustCompile("([0-9]+)")

// ConvertToTerraformCompliantName will convert the input string into a terraform compatible field name following
// Terraform's snake case field name convention (lower case and snake case).
func ConvertToTerraformCompliantName(name string) string {
	//convert the name is Snake Case, this is the ONLY operation is needed in most of the case...
	compliantName := strcase.ToSnake(name)

	// ... but if colons are present in the `name`, replace them
	compliantName = strings.ReplaceAll(compliantName, ":", "_")
	// replaced all colon characters with _ character

	if name == compliantName {
		return compliantName
	}

	// ... however if numbers are present in the `name` toSnake separated number with _X_ ...
	matches := numberInName.FindAllString(compliantName, -1)
	// ... in this case why we need to remove the ALL the surrounding underscores for each number found in `name`
	for _, match := range matches {
		positionInString := strings.Index(compliantName, match)
		// remove the prepended `_`
		tmpName := compliantName[:positionInString-1] + match
		// for the postpended `_` there we need to be careful that we don't go out of bound or we don't remove any accidental '_' present in the name
		if len(compliantName) < positionInString+2 {
			tmpName += compliantName[positionInString : len(compliantName)-1]
		} else if len(compliantName) > positionInString+2 && string(compliantName[positionInString+2]) == "_" {
			tmpName += compliantName[positionInString+2:]
		} else {
			tmpName += compliantName[positionInString+1:]
		}
		// removed the surrounding underscores for the first number match, now tmpName is compliantName
		// unless other matches are found (for loop continue)
		compliantName = tmpName
	}

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
// return a value, the default value is returned.
func MultiEnvDefaultString(ks []string, defaultValue interface{}) (string, error) {
	dv, err := multiEnvDefaultFunc(ks, defaultValue)()
	if err != nil {
		return "", err
	}
	return dv.(string), nil
}

// CastTerraformSliceToMap tries to cast the provided input to a map and returns an unaltered shallow
// copy of the input if it fails to cast. It exists to deal with the fact the Terraform defines nested
// objects as single element lists, and this method normalises the structure so that it can be
// compared against non-Terraform produced models.
func CastTerraformSliceToMap(item interface{}) (interface{}, bool) {
	if item == nil {
		return make(map[string]interface{}), true
	}
	sliceItem, successfulCast := item.([]interface{})
	if successfulCast {
		if len(sliceItem) == 0 {
			return make(map[string]interface{}), true
		}
		if len(sliceItem) == 1 {
			mapItem, successfulCast := sliceItem[0].(map[string]interface{})
			if successfulCast {
				return mapItem, true
			}
		}
	}
	return item, false
}

// CastToIntegerIfFloat tries to cast the input to an integer if it can be represented as an integer.
// Failure to cast to an integer will return an unaltered shallow copy of the input.
func CastToIntegerIfFloat(item interface{}) interface{} {
	if item == nil {
		return 0
	}
	var floatValue, successfulCast = item.(float64)
	if successfulCast {
		var integerValue = int(floatValue)
		if item == float64(integerValue) { // check if float is really an integer
			return integerValue
		}
	}
	return item
}
