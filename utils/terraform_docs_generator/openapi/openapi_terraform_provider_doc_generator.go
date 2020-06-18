package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/printers"
)

type TerraformProviderDocGenerator struct {
	ProviderName  string
	OpenAPIDocURL string
	Printer       printers.Printer
}

func (t TerraformProviderDocGenerator) GenerateDocumentation() (TerraformProviderDocumentation, error) {
	analyser, err := openapi.CreateSpecAnalyser("v2", t.OpenAPIDocURL)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}

	multiRegionConfiguration, requiredSecuritySchemes, requiredHeaders, err := t.getRequiredProviderConfigurationProperties(analyser)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}
	configProperties := []string{}
	configProperties = append(configProperties, requiredSecuritySchemes...)
	configProperties = append(configProperties, requiredHeaders...)
	providerConfiguration := ProviderConfiguration{
		Regions:          multiRegionConfiguration,
		ConfigProperties: configProperties,
	}

	return TerraformProviderDocumentation{
		ProviderName: t.ProviderName,
		ProviderInstallation: ProviderInstallation{
			Example: fmt.Sprintf("$ export PROVIDER_NAME=%s && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME<br>"+
				"[INFO] Downloading https://github.com/dikhan/terraform-provider-openapi/releases/download/v0.29.4/terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz in temporally folder /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh...<br>"+
				"[INFO] Extracting terraform-provider-openapi from terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz...<br>"+
				"[INFO] Cleaning up tmp dir created for installation purposes: /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh<br>"+
				"[INFO] Terraform provider 'terraform-provider-%s' successfully installed at: '~/.terraform.d/plugins'!", t.ProviderName, t.ProviderName),
			Other:        "You can then start running the Terraform provider:",
			OtherCommand: fmt.Sprintf("$ export OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE=\"https://api.service.com/openapi.yaml\"<br>", t.ProviderName),
		},
		ProviderConfiguration: providerConfiguration,
	}, err
}

func (t TerraformProviderDocGenerator) PrintDocumentation() error {
	analyser, err := openapi.CreateSpecAnalyser("v2", t.OpenAPIDocURL)
	if err != nil {
		return err
	}
	t.printProviderConfiguration(analyser)
	t.printProviderResources(analyser)
	return err
}

func (t TerraformProviderDocGenerator) printProviderConfiguration(analyser openapi.SpecAnalyser) error {
	t.Printer.PrintProviderConfigurationHeader(t.ProviderName)
	configProps, err := t.getRequiredProviderConfigurationProperties(analyser)
	if err != nil {
		return err
	}
	t.Printer.PrintProviderConfigurationExample(t.ProviderName, multiRegionConfiguration, requiredSecuritySchemes, requiredHeaders)
	t.Printer.PrintProviderConfiguration(multiRegionConfiguration, requiredSecuritySchemes, requiredHeaders)

	return nil
}

func (t TerraformProviderDocGenerator) getRequiredProviderConfigurationProperties(analyser openapi.SpecAnalyser) ([]Property, error) {
	var requiredSecuritySchemes []string
	var requiredHeaders []string
	var configProps []Property
	backendConfig, err := analyser.GetAPIBackendConfiguration()
	if err != nil {
		return nil, err
	}
	_, _, regions, err := backendConfig.IsMultiRegion()
	if err != nil {
		return nil, err
	}
	if regions != nil {
		configProps = append(configProps, Property{
			Name:           "region",
			Type:           "array",
			ArrayItemsType: "string",
			Required:       true,
			Description:    fmt.Sprintf("The region location to be used&nbsp;(%s). If region isn't specified, the default is '%s'", regions, regions[0]),
		})
	}
	globalSecuritySchemes, err := analyser.GetSecurity().GetGlobalSecuritySchemes()
	if err != nil {
		return nil, err
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
	return configProps, nil
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
