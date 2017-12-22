package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const OTF_VAR_SWAGGER_URL = "OTF_VAR_%s_SWAGGER_URL"

func GetServiceProviderSwaggerUrl(providerName string) (string, error) {
	otfVarSwaggerUrlLC := fmt.Sprintf(OTF_VAR_SWAGGER_URL, providerName)
	apiDiscoveryUrl := os.Getenv(otfVarSwaggerUrlLC)
	if apiDiscoveryUrl != "" {
		log.Printf("[INFO] %s set with value %s", otfVarSwaggerUrlLC, apiDiscoveryUrl)
		return apiDiscoveryUrl, nil
	}

	// Falling back to upper case check
	otfVarSwaggerUrlUC := fmt.Sprintf(OTF_VAR_SWAGGER_URL, strings.ToUpper(providerName))
	apiDiscoveryUrl = os.Getenv(otfVarSwaggerUrlUC)
	if apiDiscoveryUrl == "" {
		return "", fmt.Errorf("swagger url not provided, please export %s or %s env variable with the URL where '%s' service provider is exposing the swagger file", otfVarSwaggerUrlUC, otfVarSwaggerUrlLC, providerName)
	}

	log.Printf("[INFO] %s set with value %s", otfVarSwaggerUrlUC, apiDiscoveryUrl)
	return apiDiscoveryUrl, nil
}
