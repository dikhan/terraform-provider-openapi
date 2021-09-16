package openapi

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ProviderOpenAPI defines the struct for the OpenAPI Terraform Provider
type ProviderOpenAPI struct {
	ProviderName string
	provider     *schema.Provider
	err          error
}

// CreateSchemaProvider returns a terraform.ResourceProvider.
func (p *ProviderOpenAPI) CreateSchemaProvider() (*schema.Provider, error) {
	serviceConfiguration, err := getServiceConfiguration(p.ProviderName)
	if err != nil {
		return nil, fmt.Errorf("plugin init error: %s", err)
	}
	return p.CreateSchemaProviderFromServiceConfiguration(serviceConfiguration)
}

// CreateSchemaProviderFromServiceConfiguration helper function to enable creation of schema.Provider with the given serviceConfiguration
func (p *ProviderOpenAPI) CreateSchemaProviderFromServiceConfiguration(serviceConfiguration ServiceConfiguration) (*schema.Provider, error) {
	if p.err != nil {
		return nil, p.err
	}
	if p.provider != nil {
		return p.provider, nil
	}

	log.Printf("[DEBUG] service configuration = %+v", serviceConfiguration)

	if serviceConfiguration.IsInsecureSkipVerifyEnabled() {
		log.Printf("[WARN] Provider '%s' is using insecure skip verify, therefore the HTTPs client will not verify the API server's certificate chain and host name. This should only be used for testing purposes and it's highly recommended avoiding the use of OTF_INSECURE_SKIP_VERIFY env variable or configuring the ServiceConfiguration with InsecureSkipVerifyEnabled when executing this provider", p.ProviderName)
		tr := http.DefaultTransport.(*http.Transport)
		// #nosec G402
		tr.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
		log.Printf("[WARN] TLSClientConfig has been configured with InsecureSkipVerify set to true, this means that TLS connections will accept any certificate presented by the server and any host name in that certificate")
	}

	openAPISpecAnalyser, err := CreateSpecAnalyser(specAnalyserV3, serviceConfiguration.GetSwaggerURL())
	if err != nil {
		return nil, fmt.Errorf("plugin OpenAPI spec analyser error: %s", err)
	}

	providerFactory, err := newProviderFactory(p.ProviderName, openAPISpecAnalyser, serviceConfiguration)
	if err != nil {
		return nil, fmt.Errorf("plugin provider factory init error: %s", err)
	}

	p.provider, err = providerFactory.createProvider()
	if err != nil {
		return nil, fmt.Errorf("plugin terraform-provider-%s init error while creating schema provider: %s", p.ProviderName, err)
	}
	return p.provider, nil
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

	log.Printf("[INFO] Provider %s is using the following swagger file: %s", providerName, serviceConfiguration.GetSwaggerURL())
	return serviceConfiguration, nil
}
