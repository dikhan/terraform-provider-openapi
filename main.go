package main

import (
	"log"

	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"regexp"
)

var (
	Build string
)

func main() {

	log.Printf("OpenAPI Terraform Provider v%s", Build)

	providerName, err := getProviderName()
	if err != nil {
		log.Fatalf("[ERROR] There was an error when getting the provider's name: %s", err)
	}
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: func() terraform.ResourceProvider {
				provider, err := openapi.ProviderOpenAPI(providerName)
				if err != nil {
					log.Fatalf("[ERROR] There was an error initialising the terraform provider: %s", err)
				}
				return provider
			},
		})
}

func getProviderName() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	r, err := regexp.Compile("(\\w+)[^-]*$")
	if err != nil {
		return "", err
	}
	match := r.FindStringSubmatch(ex)
	if len(match) != 2 {
		return "", fmt.Errorf("provider name (%s) does not match terraform naming convention 'terraform-provider-{name}', please rename the provider binary", ex)
	}
	return match[0], nil
}
