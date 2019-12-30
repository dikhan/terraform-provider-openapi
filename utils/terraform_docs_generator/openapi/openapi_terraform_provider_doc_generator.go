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
	t.printProviderConfiguration(analyser)
	t.printProviderResources(analyser)
	return nil
}

func (t TerraformProviderDocGenerator) printProviderConfiguration(analyser openapi.SpecAnalyser) error {
	t.Printer.PrintProviderConfigurationHeader()
	multiRegionConfiguration, requiredSecuritySchemes, requiredHeaders, err := t.getRequiredProviderConfigurationProperties(analyser)
	if err != nil {
		return err
	}
	t.Printer.PrintProviderConfigurationExample(t.ProviderName, multiRegionConfiguration, requiredSecuritySchemes, requiredHeaders)
	t.Printer.PrintProviderConfiguration(multiRegionConfiguration, requiredSecuritySchemes, requiredHeaders)

	return nil
}

func (t TerraformProviderDocGenerator) getRequiredProviderConfigurationProperties(analyser openapi.SpecAnalyser) (*printers.MultiRegionConfiguration, []string, []string, error) {
	var multiRegionConfiguration *printers.MultiRegionConfiguration
	var requiredSecuritySchemes []string
	var requiredHeaders []string
	backendConfig, err := analyser.GetAPIBackendConfiguration()
	if err != nil {
		return nil, nil, nil, err
	}
	isMultiRegion, _, regions, err := backendConfig.IsMultiRegion()
	if err != nil {
		return nil, nil, nil, err
	}
	if isMultiRegion {
		defaultRegion, err := backendConfig.GetDefaultRegion(regions)
		if err != nil {
			return nil, nil, nil, err
		}
		multiRegionConfiguration = &printers.MultiRegionConfiguration{
			Regions:       regions,
			DefaultRegion: defaultRegion,
		}
	}
	globalSecuritySchemes, err := analyser.GetSecurity().GetGlobalSecuritySchemes()
	if err != nil {
		return nil, nil, nil, err
	}
	for _, securityScheme := range globalSecuritySchemes {
		requiredSecuritySchemes = append(requiredSecuritySchemes, securityScheme.GetTerraformConfigurationName())
	}
	headers, err := analyser.GetAllHeaderParameters()
	for _, header := range headers {
		if header.IsRequired {
			requiredHeaders = append(requiredHeaders, header.GetHeaderTerraformConfigurationName())
		}
	}
	return multiRegionConfiguration, requiredSecuritySchemes, requiredHeaders, nil
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
