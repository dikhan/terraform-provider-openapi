package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const otfVarSwaggerURL = "OTF_VAR_%s_SWAGGER_URL"

func getServiceProviderSwaggerURL(providerName string) (string, error) {
	otfVarSwaggerURLLC := fmt.Sprintf(otfVarSwaggerURL, providerName)
	apiDiscoveryURL := os.Getenv(otfVarSwaggerURLLC)
	if apiDiscoveryURL != "" {
		log.Printf("[INFO] %s set with value %s", otfVarSwaggerURLLC, apiDiscoveryURL)
		return apiDiscoveryURL, nil
	}

	// Falling back to upper case check
	otfVarSwaggerURLUC := fmt.Sprintf(otfVarSwaggerURL, strings.ToUpper(providerName))
	apiDiscoveryURL = os.Getenv(otfVarSwaggerURLUC)
	if apiDiscoveryURL == "" {
		return "", fmt.Errorf("swagger url not provided, please export %s or %s env variable with the URL where '%s' service provider is exposing the swagger file", otfVarSwaggerURLUC, otfVarSwaggerURLLC, providerName)
	}

	log.Printf("[INFO] %s set with value %s", otfVarSwaggerURLUC, apiDiscoveryURL)
	return apiDiscoveryURL, nil
}
