package openapi

import (
	"log"

	"github.com/getkin/kin-openapi/openapi3"
)

type parameterGroupsV3 [][]*openapi3.Parameter

// getHeaderConfigurations gets all the header configurations for a specific
func getHeaderConfigurationsV3(parameters []*openapi3.Parameter) SpecHeaderParameters {
	return getHeaderConfigurationsForParameterGroupsV3(parameterGroupsV3{parameters})
}

// getHeaderConfigurationsForParameterGroupsV3 loops through the provided parametersGroup (collection of parameters per operation) and
// returns a map containing all the header configurations; the key will either be the value specified in the extTfHeader
// or if not present the default value will be the name of the header. In any case, the key name will be translated to
// a terraform compliant field name if needed (more details in convertToTerraformCompliantFieldName method)
func getHeaderConfigurationsForParameterGroupsV3(parametersGroup parameterGroupsV3) SpecHeaderParameters {
	headerParameters := SpecHeaderParameters{}
	headers := map[string]string{}
	for _, parameters := range parametersGroup {
		for _, parameter := range parameters {
			// The below statement avoids dup headers in the list. Note subsequent encounters with a header type that has
			// already been registered will be ignored
			if _, exists := headers[parameter.Name]; !exists {
				headers[parameter.Name] = parameter.Name
				switch parameter.In {
				case "header":
					if preferredName, exists := getExtensionAsJsonString(parameter.Extensions, extTfHeader); exists {
						headerParameters = append(headerParameters, SpecHeaderParam{Name: parameter.Name, TerraformName: preferredName, IsRequired: parameter.Required})
					} else {
						headerParameters = append(headerParameters, SpecHeaderParam{Name: parameter.Name, IsRequired: parameter.Required})
					}
				}
			} else {
				log.Printf("[DEBUG] found duplicate header '%s' for an operation, ignoring it as it has been registered already", parameter.Name)
			}
		}
	}
	return headerParameters
}

// getPathHeaderParamsV3 aggregates all header type parameters found in the given path and returns the corresponding
// header configurations
func getPathHeaderParamsV3(path *openapi3.PathItem) SpecHeaderParameters {
	parametersGroup := parameterGroupsV3{}
	parametersGroup = appendOperationParametersIfPresentV3(parametersGroup, path.Post)
	parametersGroup = appendOperationParametersIfPresentV3(parametersGroup, path.Get)
	parametersGroup = appendOperationParametersIfPresentV3(parametersGroup, path.Put)
	parametersGroup = appendOperationParametersIfPresentV3(parametersGroup, path.Delete)
	return getHeaderConfigurationsForParameterGroupsV3(parametersGroup)
}

func getAllHeaderParametersV3(paths map[string]*openapi3.PathItem) SpecHeaderParameters {
	specHeaderParameters := SpecHeaderParameters{}
	for _, path := range paths {
		for _, headerParam := range getPathHeaderParamsV3(path) {
			// The below statement avoids dup headers in the list. Note subsequent encounters with a header type that has
			// already been registered will be ignored
			if !specHeaderParameters.specHeaderExists(headerParam) {
				specHeaderParameters = append(specHeaderParameters, headerParam)
			}
		}
	}
	return specHeaderParameters
}

// appendOperationParametersIfPresentV3 is a helper function that checks whether the given operation is not nil and if so
// appends its parameters to the parametersGroups
func appendOperationParametersIfPresentV3(parametersGroups parameterGroupsV3, operation *openapi3.Operation) parameterGroupsV3 {
	if operation != nil {
		var params []*openapi3.Parameter
		for _, param := range operation.Parameters {
			// TODO: support .Ref
			params = append(params, param.Value)
		}
		parametersGroups = append(parametersGroups, params)
	}
	return parametersGroups
}
