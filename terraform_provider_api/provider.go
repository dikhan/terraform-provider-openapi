package terraform_provider_api

import (
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"os"
	"regexp"
)

// ApiProvider returns a terraform.ResourceProvider.
func ApiProvider() *schema.Provider {
	apiDiscoveryUrl := os.Getenv("API_DISCOVERY_URL")
	d := &ProviderFactory{
		Name:            getProviderName(),
		DiscoveryApiUrl: apiDiscoveryUrl,
	}
	return d.createProvider()
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
