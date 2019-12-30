package printers

import (
	"github.com/dikhan/terraform-provider-openapi/openapi"
)

type Printer interface {
	PrintProviderConfigurationHeader()
	// Example will display the provider configuration properties that are required. For instance:
	//  - region (if multi-region is configured for the provider)
	//  - global security definitions (if provider has global definitions they are considered required)
	//  - headers (some resource operations require header)
	PrintProviderConfigurationExample(providerName string, multiRegionConfiguration *MultiRegionConfiguration, requiredSecuritySchemes, requiredHeaders []string)
	// printing for now only the required properties in the configuration
	PrintProviderConfiguration(multiRegionConfiguration *MultiRegionConfiguration, requiredSecuritySchemes, requiredHeaders []string)
	PrintResourceHeader()
	PrintResourceInfo(providerName, resourceName string)
	PrintResourceExample(providerName, resourceName string, required openapi.SpecSchemaDefinitionProperties)
	PrintArguments(required, optional openapi.SpecSchemaDefinitionProperties)
	PrintAttributes(computed openapi.SpecSchemaDefinitionProperties)
}
