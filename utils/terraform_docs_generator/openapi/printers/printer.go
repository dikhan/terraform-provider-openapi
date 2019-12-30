package printers

import (
	"github.com/dikhan/terraform-provider-openapi/openapi"
)

type Printer interface {
	PrintProviderConfigurationHeader()
	PrintProviderConfigurationExample(providerName string)
	PrintProviderConfiguration()
	PrintResourceHeader()
	PrintResourceInfo(providerName, resourceName string)
	PrintResourceExample(providerName, resourceName string, required openapi.SpecSchemaDefinitionProperties)
	PrintArguments(required, optional openapi.SpecSchemaDefinitionProperties)
	PrintAttributes(computed openapi.SpecSchemaDefinitionProperties)
}
