package terraform_provider_api

import (
	"log"
	"os"
	"regexp"

	"github.com/hashicorp/terraform/helper/schema"
)

// ApiProvider returns a terraform.ResourceProvider.
func ApiProvider() *schema.Provider {
	apiDiscoveryUrl := getApiDiscoveryUrl()
	d := &ProviderFactory{
		Name:            getProviderName(),
		DiscoveryApiUrl: apiDiscoveryUrl,
	}
	return d.createProvider()
}

// This function is implemented with temporary code thus it can serve as an example
// on how the same code base can be used by binaries of this same provider named differently
// but internally each will end up calling a different service provider's api
func getApiDiscoveryUrl() string {
	var apiDiscoveryUrl string
	providerName := getProviderName()
	if providerName != "sp" {
		log.Fatalf("%s provider not supported...", providerName)
	}
	if providerName == "sp" {
		apiDiscoveryUrl = "http://localhost:8080/swagger.json"
	}
	return apiDiscoveryUrl
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
