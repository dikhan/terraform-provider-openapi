package main

import (
	"os"
	"regexp"

	"net/http"

	"crypto/tls"

	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

// APIProvider returns a terraform.ResourceProvider.
func APIProvider() (*schema.Provider, error) {
	providerName, apiDiscoveryURL, err := getProviderNameAndAPIDiscoveryURL()
	if err != nil {
		return nil, fmt.Errorf("plugin init error: %s", err)
	}
	d := &providerFactory{
		name:            providerName,
		discoveryAPIURL: apiDiscoveryURL,
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
func getProviderNameAndAPIDiscoveryURL() (string, string, error) {
	var apiDiscoveryURL string
	providerName, err := getProviderName()
	if err != nil {
		return "", "", err
	}
	apiDiscoveryURL, err = getServiceProviderSwaggerURL(providerName)
	if err != nil {
		return "", "", err
	}
	if skip, ok := os.LookupEnv("OTF_INSECURE_SKIP_VERIFY"); ok {
		if skip, _ := strconv.ParseBool(skip); skip {
			tr := http.DefaultTransport.(*http.Transport)
			tr.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
			log.Printf("[WARN] Provider %s is using insecure skip verify for '%s'. Please make sure you trust the aforementioned server hosting the swagger file. Otherwise, it's highly recommended avoiding the use of OTF_INSECURE_SKIP_VERIFY env variable when executing this provider", providerName, apiDiscoveryURL)
		}
	}
	log.Printf("[INFO] Provider %s is using the following remote swagger URL: %s", providerName, apiDiscoveryURL)
	return providerName, apiDiscoveryURL, nil
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
