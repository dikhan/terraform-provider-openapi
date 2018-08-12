package openapiutils

import (
	"github.com/go-openapi/spec"
	"regexp"
	"strings"
)

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
					headerConfigProps[convertToTerraformCompliantFieldName(headerConfigProp)] = parameter
				} else {
					headerConfigProps[convertToTerraformCompliantFieldName(parameter.Name)] = parameter
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

func convertToTerraformCompliantFieldName(name string) string {
	// lowering the case of the name for name consistency reasons
	lowerCaseName := strings.ToLower(name)
	// replace non terraform field compliant characters
	return strings.Replace(lowerCaseName, "-", "_", -1)
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
