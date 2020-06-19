package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/printers"
	"github.com/mitchellh/hashstructure"
	"sort"
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
	regions, configProperties, err := t.getRequiredProviderConfigurationProperties(analyser)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}

	resources, err := t.getProviderResources(analyser)
	if err != nil {
		return TerraformProviderDocumentation{}, err
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
		ProviderConfiguration: ProviderConfiguration{
			Regions:          regions,
			ConfigProperties: configProperties,
		},
		ProviderResources: ProviderResources{
			Resources: resources,
		},
	}, err
}

func (t TerraformProviderDocGenerator) getProviderResources(analyser openapi.SpecAnalyser) ([]Resource, error) {
	r := []Resource{}
	resources, err := analyser.GetTerraformCompliantResources()
	if err != nil {
		return nil, err
	}
	for _, resource := range resources {
		if resource.ShouldIgnoreResource() {
			continue
		}
		resourceSchema, err := resource.GetResourceSchema()
		if err != nil {
			return nil, err
		}
		props := []Property{}
		requiredProps := []Property{}
		optionalProps := []Property{}
		for _, p := range resourceSchema.Properties {
			prop := t.resourceSchemaToProperty(*p)
			if prop.Required {
				requiredProps = append(requiredProps, prop)
			}
			if !prop.Required {
				optionalProps = append(optionalProps, prop)
			}
		}
		props = append(props, orderProps(requiredProps)...)
		props = append(props, orderProps(optionalProps)...)

		r = append(r, Resource{
			Name:        resource.GetResourceName(),
			Description: "",
			Properties:  props,
			ArgumentsReference: ArgumentsReference{
				Notes: []string{},
			},
		})
	}
	return r, nil
}

func (t TerraformProviderDocGenerator) resourceSchemaToProperty(specSchemaDefinitionProperty openapi.SpecSchemaDefinitionProperty) Property {
	var schema []Property
	if specSchemaDefinitionProperty.Type == openapi.TypeObject || specSchemaDefinitionProperty.ArrayItemsType == openapi.TypeObject {
		if specSchemaDefinitionProperty.SpecSchemaDefinition != nil {
			for _, p := range specSchemaDefinitionProperty.SpecSchemaDefinition.Properties {
				schema = append(schema, t.resourceSchemaToProperty(*p))
			}
		}
	}
	return Property{
		Name:           specSchemaDefinitionProperty.GetTerraformCompliantPropertyName(),
		Type:           string(specSchemaDefinitionProperty.Type),
		ArrayItemsType: string(specSchemaDefinitionProperty.ArrayItemsType),
		Required:       specSchemaDefinitionProperty.IsRequired(),
		Computed:       specSchemaDefinitionProperty.Computed,
		Description:    specSchemaDefinitionProperty.Description,
		Schema:         schema,
	}
}

func (t TerraformProviderDocGenerator) getRequiredProviderConfigurationProperties(analyser openapi.SpecAnalyser) ([]string, []Property, error) {
	var configProps []Property
	backendConfig, err := analyser.GetAPIBackendConfiguration()
	if err != nil {
		return nil, nil, err
	}
	_, _, regions, err := backendConfig.IsMultiRegion()
	if err != nil {
		return nil, nil, err
	}
	globalSecuritySchemes, err := analyser.GetSecurity().GetGlobalSecuritySchemes()
	if err != nil {
		return nil, nil, err
	}
	securityDefinitions, err := analyser.GetSecurity().GetAPIKeySecurityDefinitions()
	if err != nil {
		return nil, nil, err
	}
	for _, securityDefinition := range *securityDefinitions {
		secDefName := securityDefinition.GetTerraformConfigurationName()
		configProps = append(configProps, Property{
			Name:        secDefName,
			Type:        "string",
			Required:    false,
			Description: "",
		})
	}
	// Mark as required the properties that are set in the security schemes (they are mandatory)
	for _, securityScheme := range globalSecuritySchemes {
		for _, configProp := range configProps {
			c := &configProp
			if c.Name == securityScheme.GetTerraformConfigurationName() {
				c.Required = true
				break
			}
		}
	}

	headers, err := analyser.GetAllHeaderParameters()
	for _, header := range headers {
		configProps = append(configProps, Property{
			Name:        header.GetHeaderTerraformConfigurationName(),
			Type:        "string",
			Required:    header.IsRequired,
			Description: "",
		})
	}
	return regions, configProps, nil
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

func orderProps(props []Property) []Property {
	sort.Slice(props, func(i, j int) bool {
		hash1, _ := hashstructure.Hash(props[i], nil)
		hash2, _ := hashstructure.Hash(props[j], nil)
		return hash1 > hash2
	})
	return props
}
