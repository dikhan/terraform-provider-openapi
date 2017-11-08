package main

import (
	"log"
	"os"
	"regexp"

	"net/http"

	"crypto/tls"

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
		// This is a temporary solution to be able to test out the example 'sp' binary produced
		// The reason why this is needed is because the http client library will reject the cert from the server
		// as it's not signed by an ofitial CA, it's just a self signed cert
		tr := http.DefaultTransport.(*http.Transport)
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		apiDiscoveryUrl = "https://localhost:8443/swagger.json"
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
