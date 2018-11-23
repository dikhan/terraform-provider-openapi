package openapi

import (
	"net/http"

	"crypto/tls"

	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
)

// ProviderOpenAPI returns a terraform.ResourceProvider.
func ProviderOpenAPI(providerName string) (*schema.Provider, error) {
	serviceConfiguration, err := getServiceConfiguration(providerName)
	if err != nil {
		return nil, fmt.Errorf("plugin init error: %s", err)
	}

	log.Printf("[DEBUG] service configuration = %+v", serviceConfiguration)

	openAPISpecAnalyser, err := CreateSpecAnalyser(specAnalyserV2, serviceConfiguration.GetSwaggerURL())
	if err != nil {
		return nil, fmt.Errorf("plugin OpenAPI spec analyser error: %s", err)
	}

	providerFactory, err := newProviderFactory(providerName, openAPISpecAnalyser, serviceConfiguration)
	if err != nil {
		return nil, fmt.Errorf("plugin provider factory init error: %s", err)
	}

	provider, err := providerFactory.createProvider()
	if err != nil {
		return nil, fmt.Errorf("plugin terraform-provider-%s init error while creating schema provider: %s", providerName, err)
	}
	return provider, nil
}

// This function is implemented with temporary code thus it can serve as an example
// on how the same code base can be used by binaries of this same provider named differently
// but internally each will end up calling a different service provider's api
func getServiceConfiguration(providerName string) (ServiceConfiguration, error) {
	var serviceConfiguration ServiceConfiguration
	pluginConfiguration, err := NewPluginConfiguration(providerName)
	if err != nil {
		return nil, err
	}
	serviceConfiguration, err = pluginConfiguration.getServiceConfiguration()
	if err != nil {
		return nil, err
	}

	if serviceConfiguration.IsInsecureSkipVerifyEnabled() {
		tr := http.DefaultTransport.(*http.Transport)
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		log.Printf("[WARN] Provider %s is using insecure skip verify for '%s'. Please make sure you trust the aforementioned server hosting the swagger file. Otherwise, it's highly recommended avoiding the use of OTF_INSECURE_SKIP_VERIFY env variable when executing this provider", providerName, serviceConfiguration)
	}

	log.Printf("[INFO] Provider %s is using the following swagger file: %s", providerName, serviceConfiguration.GetSwaggerURL())
	return serviceConfiguration, nil
}
