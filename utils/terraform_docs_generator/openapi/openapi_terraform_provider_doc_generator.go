package openapi

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/mitchellh/hashstructure"
	"sort"
)

// TerraformProviderDocGenerator defines the struct that holds the configuration needed to be able to generate the documentation
type TerraformProviderDocGenerator struct {
	// ProviderName defines the provider name
	ProviderName string
	// OpenAPIDocURL defines the URL where the service provider OpenAPI documentation lives and it will be used to fetch from
	// to create the corresponding TerraformProviderDocumentation
	OpenAPIDocURL string
}

// GenerateDocumentation creates a TerraformProviderDocumentation object populated based on the OpenAPIDocURL documentation
func (t TerraformProviderDocGenerator) GenerateDocumentation() (TerraformProviderDocumentation, error) {
	analyser, err := openapi.CreateSpecAnalyser("v2", t.OpenAPIDocURL)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}

	regions, err := getRegions(analyser)
	if err != nil {
		return TerraformProviderDocumentation{}, nil
	}
	globalSecuritySchemes, securityDefinitions, err := getSecurity(analyser)
	if err != nil {
		return TerraformProviderDocumentation{}, nil
	}
	headers, err := analyser.GetAllHeaderParameters()
	if err != nil {
		return TerraformProviderDocumentation{}, nil
	}
	configRegions, configProperties := t.getRequiredProviderConfigurationProperties(regions, globalSecuritySchemes, securityDefinitions, headers)

	r, err := analyser.GetTerraformCompliantResources()
	if err != nil {
		return TerraformProviderDocumentation{}, nil
	}
	resources, err := t.getProviderResources(r)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}

	dataSourceInstances, err := t.getDataSourceInstances(r)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}

	compliantDataSources := analyser.GetTerraformCompliantDataSources()
	dataSourceFilters, err := t.getDataSourceFilters(compliantDataSources)
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
			Regions:          configRegions,
			ConfigProperties: configProperties,
		},
		ProviderResources: ProviderResources{
			Resources: resources,
		},
		DataSources: DataSources{
			DataSources:         dataSourceFilters,
			DataSourceInstances: dataSourceInstances,
		},
	}, err
}

func getRegions(s openapi.SpecAnalyser) ([]string, error) {
	backendConfig, err := s.GetAPIBackendConfiguration()
	if err != nil {
		return nil, err
	}
	if backendConfig != nil {
		_, _, regions, err := backendConfig.IsMultiRegion()
		if err != nil {
			return nil, err
		}
		return regions, nil
	}
	return nil, nil
}

func getSecurity(s openapi.SpecAnalyser) (openapi.SpecSecuritySchemes, *openapi.SpecSecurityDefinitions, error) {
	security := s.GetSecurity()
	if security != nil {
		globalSecuritySchemes, err := security.GetGlobalSecuritySchemes()
		if err != nil {
			return nil, nil, err
		}
		securityDefinitions, err := security.GetAPIKeySecurityDefinitions()
		if err != nil {
			return nil, nil, err
		}
		return globalSecuritySchemes, securityDefinitions, nil
	}
	return nil, nil, nil
}

func (t TerraformProviderDocGenerator) getDataSourceFilters(dataSourcesFilter []openapi.SpecResource) ([]DataSource, error) {
	dataSources := []DataSource{}
	for _, dataSource := range dataSourcesFilter {
		s, err := dataSource.GetResourceSchema()
		if err != nil {
			return nil, err
		}
		dataSourceSchemaDefinition := s.ConvertToDataSourceSpecSchemaDefinition()
		if err != nil {
			return nil, err
		}
		props := []Property{}
		for _, p := range dataSourceSchemaDefinition.Properties {
			prop := t.resourceSchemaToProperty(*p)
			props = append(props, prop)
		}
		dataSources = append(dataSources, DataSource{
			Name:       dataSource.GetResourceName(),
			Properties: orderProps(props),
		})
	}
	return dataSources, nil
}

func (t TerraformProviderDocGenerator) getDataSourceInstances(dataSourceInstances []openapi.SpecResource) ([]DataSource, error) {
	dataSourcesInstance := []DataSource{}
	for _, dataSource := range dataSourceInstances {
		s, err := dataSource.GetResourceSchema()
		if err != nil {
			return nil, err
		}
		dataSourceSchemaDefinition := s.ConvertToDataSourceSpecSchemaDefinition()
		if err != nil {
			return nil, err
		}
		props := []Property{}
		for _, p := range dataSourceSchemaDefinition.Properties {
			prop := t.resourceSchemaToProperty(*p)
			props = append(props, prop)
		}
		dataSourcesInstance = append(dataSourcesInstance, DataSource{
			Name:       fmt.Sprintf("%s_instance", dataSource.GetResourceName()),
			Properties: orderProps(props),
		})
	}
	return dataSourcesInstance, nil
}

func (t TerraformProviderDocGenerator) getProviderResources(resources []openapi.SpecResource) ([]Resource, error) {
	r := []Resource{}
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

		parentInfo := resource.GetParentResourceInfo()
		var parentProperties []string
		if parentInfo != nil {
			parentProperties = parentInfo.GetParentPropertiesNames()
		}

		r = append(r, Resource{
			Name:             resource.GetResourceName(),
			Description:      "",
			Properties:       props,
			ParentProperties: parentProperties,
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
		Name:               specSchemaDefinitionProperty.GetTerraformCompliantPropertyName(),
		Type:               string(specSchemaDefinitionProperty.Type),
		ArrayItemsType:     string(specSchemaDefinitionProperty.ArrayItemsType),
		Required:           specSchemaDefinitionProperty.IsRequired(),
		Computed:           specSchemaDefinitionProperty.Computed,
		IsOptionalComputed: specSchemaDefinitionProperty.IsOptionalComputed(),
		Description:        specSchemaDefinitionProperty.Description,
		Schema:             orderProps(schema),
	}
}

func (t TerraformProviderDocGenerator) getRequiredProviderConfigurationProperties(regions []string, globalSecuritySchemes openapi.SpecSecuritySchemes, securityDefinitions *openapi.SpecSecurityDefinitions, headers openapi.SpecHeaderParameters) ([]string, []Property) {
	var configProps []Property
	if securityDefinitions != nil {
		for _, securityDefinition := range *securityDefinitions {
			secDefName := securityDefinition.GetTerraformConfigurationName()
			configProps = append(configProps, Property{
				Name:        secDefName,
				Type:        "string",
				Required:    false,
				Description: "",
			})
		}
	}
	// Mark as required the properties that are set in the security schemes (they are mandatory)
	if globalSecuritySchemes != nil {
		for _, securityScheme := range globalSecuritySchemes {
			for idx, configProp := range configProps {
				if configProp.Name == securityScheme.GetTerraformConfigurationName() {
					configProps[idx].Required = true
					break
				}
			}
		}
	}

	if headers != nil {
		for _, header := range headers {
			configProps = append(configProps, Property{
				Name:        header.GetHeaderTerraformConfigurationName(),
				Type:        "string",
				Required:    header.IsRequired,
				Description: "",
			})
		}
	}
	return regions, configProps
}

func orderProps(props []Property) []Property {
	sort.Slice(props, func(i, j int) bool {
		hash1, _ := hashstructure.Hash(props[i], nil)
		hash2, _ := hashstructure.Hash(props[j], nil)
		return hash1 > hash2
	})
	return props
}
