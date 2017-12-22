package main

import (
	"os"
	"regexp"

	"net/http"

	"crypto/tls"

	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

// ApiProvider returns a terraform.ResourceProvider.
func ApiProvider() (*schema.Provider, error) {
	providerName, apiDiscoveryUrl, err := getProviderNameAndApiDiscoveryUrl()
	if err != nil {
		return nil, fmt.Errorf("plugin init error: %s", err)
	}
	d := &ProviderFactory{
		Name:            providerName,
		DiscoveryApiUrl: apiDiscoveryUrl,
	}
	provider, err := d.createProvider()
	if err != nil {
		return nil, fmt.Errorf("plugin terraform-provider-%s init error: %s", providerName, err)
	}
	return provider, nil
}

// This function is implemented with temporary code thus it can serve as an example
// on how the same code base can be used by binaries of this same provider named differently
// but internally each will end up calling a different service provider's api
func getProviderNameAndApiDiscoveryUrl() (string, string, error) {
	var apiDiscoveryUrl string
	providerName, err := getProviderName()
	if err != nil {
		return "", "", err
	}
	apiDiscoveryUrl, err = GetServiceProviderSwaggerUrl(providerName)
	if err != nil {
		return "", "", err
	}
	if providerName == "sp" {
		// This is a temporary solution to be able to test out the example 'sp' binary produced
		// The reason why this is needed is because the http client library will reject the cert from the server
		// as it's not signed by an official CA, it's just a self signed cert.
		tr := http.DefaultTransport.(*http.Transport)
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	log.Printf("[INFO] Provider %s is using the following remote swagger URL: %s", providerName, apiDiscoveryUrl)
	return providerName, apiDiscoveryUrl, nil
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
