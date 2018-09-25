package openapiutils

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/terraformutils"
	"github.com/go-openapi/spec"
	"regexp"
	"strings"
)

const swaggerResourcePayloadDefinitionRegex = "(\\w+)[^//]*$"
const fqdnInURLRegex = `\b(?:(?:[^.-/]{0,1})[\w-]{1,63}[-]{0,1}[.]{1})+(?:[a-zA-Z]{2,63})?|localhost(?:[:]\d+)?\b`
const extTfHeader = "x-terraform-header"

// HeaderConfigurations defines the header configurations that on runtime will be used by the resource factory to
// lookup which headers should be passed along with the request
type HeaderConfigurations map[string]spec.Parameter
type parameterGroups [][]spec.Parameter

// GetHeaderConfigurations gets all the header configurations for a specific
func GetHeaderConfigurations(parameters []spec.Parameter) HeaderConfigurations {
	return GetHeaderConfigurationsForParameterGroups(parameterGroups{parameters})
}

// GetHeaderConfigurationsForParameterGroups loops through the provided parametersGroup (collection of parameters per operation) and
// returns a map containing all the header configurations; the key will either be the value specified in the extTfHeader
// or if not present the default value will be the name of the header. In any case, the key name will be translated to
// a terraform compliant field name if needed (more details in convertToTerraformCompliantFieldName method)
func GetHeaderConfigurationsForParameterGroups(parametersGroup parameterGroups) HeaderConfigurations {
	headerConfigProps := HeaderConfigurations{}
	for _, parameters := range parametersGroup {
		for _, parameter := range parameters {
			switch parameter.In {
			case "header":
				if headerConfigProp, exists := parameter.Extensions.GetString(extTfHeader); exists {
					headerConfigProps[terraformutils.ConvertToTerraformCompliantName(headerConfigProp)] = parameter
				} else {
					headerConfigProps[terraformutils.ConvertToTerraformCompliantName(parameter.Name)] = parameter
				}
			}
		}
	}
	return headerConfigProps
}

// GetAllHeaderParameters gets all the parameters of type headers present in the swagger file and returns the header
// configurations. Currently only the following parameters are supported:
// - root level parameters (not supported)
// - path level parameters (not supported)
// - operation level parameters (supported)
func GetAllHeaderParameters(spec *spec.Swagger) HeaderConfigurations {
	headerConfigProps := HeaderConfigurations{}
	// add header configuration names/values defined per path operation
	for _, path := range spec.Paths.Paths {
		for k, v := range getPathHeaderParams(path) {
			headerConfigProps[k] = v
		}
	}
	return headerConfigProps
}

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

// getPathHeaderParams aggregates all header type parameters found in the given path and returns the corresponding
// header configurations
func getPathHeaderParams(path spec.PathItem) HeaderConfigurations {
	parametersGroup := parameterGroups{}
	parametersGroup = appendOperationParametersIfPresent(parametersGroup, path.Post)
	parametersGroup = appendOperationParametersIfPresent(parametersGroup, path.Get)
	parametersGroup = appendOperationParametersIfPresent(parametersGroup, path.Put)
	parametersGroup = appendOperationParametersIfPresent(parametersGroup, path.Delete)
	return GetHeaderConfigurationsForParameterGroups(parametersGroup)
}

// appendOperationParametersIfPresent is a helper function that checks whether the given operation is not nil and if so
// appends its parameters to the parametersGroups
func appendOperationParametersIfPresent(parametersGroups parameterGroups, operation *spec.Operation) parameterGroups {
	if operation != nil {
		parametersGroups = append(parametersGroups, operation.Parameters)
	}
	return parametersGroups
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
