package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"

	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v1/openapi"
	"github.com/dikhan/terraform-provider-openapi/v1/openapi/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"os"
	"regexp"
)

func main() {

	log.Printf("Running OpenAPI Terraform Provider v%s-%s; Released on: %s", version.Version, version.Commit, version.Date)

	ex, err := os.Executable()
	if err != nil {
		log.Fatalf("[ERROR] There was an error when getting the provider binary name: %s", err)
	}

	providerName, err := getProviderName(ex)
	if err != nil {
		log.Fatalf("[ERROR] There was an error when getting the provider's name from the binary '%s': %s", ex, err)
	}

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProvider()
	if err != nil {
		log.Fatalf("[ERROR] There was an error initialising the terraform provider: %s", err)
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return provider
		}})
}

func getProviderName(binaryName string) (string, error) {
	r, err := regexp.Compile("\\bterraform-provider-([a-zA-Z0-9]+)(?:_v[\\d]+\\.[\\d]+\\.[\\d]+)?\\b")
	if err != nil {
		return "", err
	}

	match := r.FindStringSubmatch(binaryName)
	if len(match) != 2 {
		return "", fmt.Errorf("provider binary name (%s) does not match terraform naming convention 'terraform-provider-{name}', please rename the provider binary", binaryName)
	}
	return match[1], nil
}
