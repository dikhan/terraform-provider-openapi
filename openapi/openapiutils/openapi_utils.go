package openapiutils

import (
	"fmt"
	"github.com/go-openapi/spec"
	"regexp"
	"strings"
)

const swaggerResourcePayloadDefinitionRegex = "(\\w+)[^//]*$"
const fqdnInURLRegex = `\b(?:(?:[^.-/]{0,1})[\w-]{1,63}[-]{0,1}[.]{1})+(?:[a-zA-Z]{2,63})?(?:[:]\d+)?|localhost(?:[:]\d+)?\b`

// GetHostFromURL returns the fqdn of a given string (localhost including port number is also handled).
// Example domains that would match:
// - http://domain.com/
// - domain.com/parameter
// - domain.com?anything
// - example.domain.com
// - example.domain-hyphen.com
// - www.domain.com
// - localhost
// - localhost:8080
// Example domains that would not match:
// - http://domain.com:8080 (this use case is not support at the moment, it is assumed that actual domains will use standard ports)
func GetHostFromURL(url string) string {
	re := regexp.MustCompile(fqdnInURLRegex)
	return re.FindString(url)
}

// StringExtensionExists tries to find a match using the built-in extensions GetString method and if there is no match
// it will try to find a match without converting the key lower case (as done behind the scenes by GetString method).
// Context: The Extensions look up methods tweaks the given key making it lower case and then trying to match against
// the keys in the map. However this may not always work as the Extensions might have been added without going through
// the Add method which lower cases the key, though in situations where the struct was un-marshaled directly instead this
// translation would not have happened and therefore the look up queiry will not find matches
func StringExtensionExists(extensions spec.Extensions, key string) (string, bool) {
	var value string
	value, exists := extensions.GetString(key)
	if !exists {
		// Fall back to look up with actual given key name (without converting to lower case as the GetString method from extensions does behind the scenes)
		for k, v := range extensions {
			if strings.ToLower(k) == strings.ToLower(key) {
				return v.(string), true
			}
		}
	}
	return value, exists
}

// getPayloadDefName only supports references to the same document. External references like URLs is not supported at the moment
func getPayloadDefName(ref string) (string, error) {
	reg, err := regexp.Compile(swaggerResourcePayloadDefinitionRegex)
	if err != nil {
		return "", fmt.Errorf("an error occurred while compiling the swaggerResourcePayloadDefinitionRegex regex '%s': %s", swaggerResourcePayloadDefinitionRegex, err)
	}
	payloadDefName := reg.FindStringSubmatch(ref)[0]
	if payloadDefName == "" {
		return "", fmt.Errorf("could not find a valid definition name for '%s'", ref)
	}
	return payloadDefName, nil
}

// GetSchemaDefinition queries the definitions and tries to find the schema definition for the given ref. If the schema
// definition the ref value is pointing at does not exist and error is returned. Otherwise, the corresponding schema definition is returned.
func GetSchemaDefinition(definitions map[string]spec.Schema, ref string) (*spec.Schema, error) {
	payloadDefName, err := getPayloadDefName(ref)
	if err != nil {
		return nil, err
	}
	payloadDefinition, exists := definitions[payloadDefName]
	if !exists {
		return nil, fmt.Errorf("missing schema definition in the swagger file with the supplied ref '%s'", ref)
	}
	return &payloadDefinition, nil
}
