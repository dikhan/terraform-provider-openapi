package terraform_provider_api

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"os"
	"regexp"
)

// ApiProvider returns a terraform.ResourceProvider.
func ApiProvider() *schema.Provider {
	d := &DynamicProviderFactory{
		Name: getProviderName(),
	}
	return d.createProviderDynamically()
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	return nil, nil
}

func getProviderName() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	r, err := regexp.Compile("(\\w+)[^-]*$")
	if err != nil {
		panic(err)
	}
	match := r.FindStringSubmatch(ex)
	if len(match) != 2 {
		log.Fatalf("provider name does not match terraform naming convention 'terraform-provider-{name}', please rename the provider binary")
	}
	return match[0]
}
