package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/printers"
)

type TerraformProviderDocGenerator struct {
	ProviderName  string
	OpenAPIDocURL string
	Printer       printers.Printer
}

func (t TerraformProviderDocGenerator) GenerateDocumentation() error {
	analyser, err := openapi.CreateSpecAnalyser("v2", t.OpenAPIDocURL)
	if err != nil {
		return err
	}
	t.printProviderResources(analyser)
	return nil
}

func (t TerraformProviderDocGenerator) printProviderResources(analyser openapi.SpecAnalyser) error {
	resources, err := analyser.GetTerraformCompliantResources()
	if err != nil {
		return err
	}
	for _, resource := range resources {
		if resource.ShouldIgnoreResource() {
			continue
		}
		resourceSchema, err := resource.GetResourceSchema()
		if err != nil {
			return err
		}
		t.printResourceDoc(resource.GetResourceName(), resourceSchema)
	}
	return nil
}

func (t TerraformProviderDocGenerator) printResourceDoc(resourceName string, resourceSchema *openapi.SpecSchemaDefinition) {
	required, optional, computed := t.createRequiredOptionalComputedMaps(resourceSchema)
	t.Printer.PrintResourceHeader()
	t.Printer.PrintResourceInfo(t.ProviderName, resourceName)
	t.Printer.PrintResourceExample(t.ProviderName, resourceName, required)
	t.Printer.PrintArguments(required, optional)
	t.Printer.PrintAttributes(computed)
}

func (t TerraformProviderDocGenerator) createRequiredOptionalComputedMaps(resourceSchema *openapi.SpecSchemaDefinition) (openapi.SpecSchemaDefinitionProperties, openapi.SpecSchemaDefinitionProperties, openapi.SpecSchemaDefinitionProperties) {
	required := openapi.SpecSchemaDefinitionProperties{}
	optional := openapi.SpecSchemaDefinitionProperties{}
	computed := openapi.SpecSchemaDefinitionProperties{}
	for _, property := range resourceSchema.Properties {
		if property.Required {
			required = append(required, property)
		} else {
			if property.Computed {
				computed = append(computed, property)
			} else {
				optional = append(optional, property)
			}
		}
	}
	return required, optional, computed
}
