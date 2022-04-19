package openapiterraformdocsgenerator

import (
	"errors"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v3/openapi"
	"github.com/mitchellh/hashstructure"
	"log"
	"sort"
)

// TerraformProviderDocGenerator defines the struct that holds the configuration needed to be able to generate the documentation
type TerraformProviderDocGenerator struct {
	// ProviderName defines the provider name
	ProviderName string
	// Hostname the Terraform registry that distributes the provider as documented in https://www.terraform.io/docs/language/providers/requirements.html#source-addresses
	// For in-house providers that you intend to distribute from a local filesystem directory, you can use an arbitrary hostname in a domain your organization controls. For example, if your corporate domain were example.com then you might choose
	// to use terraform.example.com as your placeholder hostname, even if that hostname doesn't actually resolve in DNS.
	Hostname string
	// Namespace An organizational namespace within the specified registry to be used for configuration purposes as documented in https://www.terraform.io/docs/language/providers/requirements.html#source-addresses
	Namespace string
	// PluginVersionConstraint should contain the OpenAPI plugin version constraint eg: "~> 2.1.0". If not populated the renderer
	// will default to ">= 2.1.0" OpenAPI provider version
	PluginVersionConstraint string
	// SpecAnalyser analyses the swagger doc and provides helper methods to retrieve all the end points that can
	// be used as terraform resources.
	SpecAnalyser openapi.SpecAnalyser
}

// NewTerraformProviderDocGenerator returns a TerraformProviderDocGenerator populated with the provider documentation which
// exposes methods to render the documentation in different formats (only html supported at the moment)
func NewTerraformProviderDocGenerator(providerName, hostname, namespace, openAPIDocURL string) (TerraformProviderDocGenerator, error) {
	analyser, err := openapi.CreateSpecAnalyser("v2", openAPIDocURL)
	if err != nil {
		return TerraformProviderDocGenerator{}, err
	}
	return TerraformProviderDocGenerator{
		ProviderName: providerName,
		Hostname:     hostname,
		Namespace:    namespace,
		SpecAnalyser: analyser,
	}, nil
}

// GenerateDocumentation creates a TerraformProviderDocumentation object populated based on the OpenAPIDocURL documentation
func (t TerraformProviderDocGenerator) GenerateDocumentation() (TerraformProviderDocumentation, error) {
	if t.ProviderName == "" {
		return TerraformProviderDocumentation{}, errors.New("provider name not provided")
	}
	if t.Hostname == "" {
		return TerraformProviderDocumentation{}, errors.New("hostname not provided, this is required to be able to render the provider installation section containing the required_providers block with the source address configuration in the form of [<HOSTNAME>/]<NAMESPACE>/<TYPE>")
	}
	if t.Namespace == "" {
		return TerraformProviderDocumentation{}, errors.New("namespace not provided, this is required to be able to render the provider installation section containing the required_providers block with the source address configuration in the form of [<HOSTNAME>/]<NAMESPACE>/<TYPE>")
	}
	if t.PluginVersionConstraint == "" {
		log.Println("PluginVersionConstraint not provided, default value in the plugin's terraform required_providers rendered documentation will be version = \">= 2.1.0\"")
	}
	regions, err := getRegions(t.SpecAnalyser)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}
	globalSecuritySchemes, securityDefinitions, err := getSecurity(t.SpecAnalyser)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}
	headers := t.SpecAnalyser.GetAllHeaderParameters()
	configRegions, configProperties := t.getRequiredProviderConfigurationProperties(regions, globalSecuritySchemes, securityDefinitions, headers)

	r, err := t.SpecAnalyser.GetTerraformCompliantResources()
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}
	resources, err := t.getProviderResources(r)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}

	// ignoring error from getDataSourceInstances bc resource errors will be caught when looping through resources in getProviderResources
	dataSourceInstances, _ := t.getDataSourceInstances(r)

	compliantDataSources := t.SpecAnalyser.GetTerraformCompliantDataSources()
	dataSourceFilters, err := t.getDataSourceFilters(compliantDataSources)
	if err != nil {
		return TerraformProviderDocumentation{}, err
	}

	sort.SliceStable(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})
	sort.SliceStable(dataSourceInstances, func(i, j int) bool {
		return dataSourceInstances[i].Name < dataSourceInstances[j].Name
	})
	sort.SliceStable(dataSourceFilters, func(i, j int) bool {
		return dataSourceFilters[i].Name < dataSourceFilters[j].Name
	})

	return TerraformProviderDocumentation{
		ProviderName: t.ProviderName,
		ProviderInstallation: ProviderInstallation{
			ProviderName:            t.ProviderName,
			Namespace:               t.Namespace,
			Hostname:                t.Hostname,
			PluginVersionConstraint: t.PluginVersionConstraint,
			Example: fmt.Sprintf("$ export PROVIDER_NAME=%s && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME<br>"+
				"[INFO] Downloading https://github.com/dikhan/terraform-provider-openapi/v3/releases/download/v3.0.0/terraform-provider-openapi_3.0.0_darwin_amd64.tar.gz in temporally folder /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh...<br>"+
				"[INFO] Extracting terraform-provider-openapi from terraform-provider-openapi_0.29.4_darwin_amd64.tar.gz...<br>"+
				"[INFO] Cleaning up tmp dir created for installation purposes: /var/folders/n_/1lrwb99s7f50xmn9jpmfnddh0000gp/T/tmp.Xv1AkIZh<br>"+
				"[INFO] Terraform provider 'terraform-provider-%s' successfully installed at: '~/.terraform.d/plugins'!", t.ProviderName, t.ProviderName),
			Other:        "You can then start running the Terraform provider:",
			OtherCommand: fmt.Sprintf(`$ export OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE="https://api.service.com/openapi.yaml"<br>`, t.ProviderName),
		},
		ProviderConfiguration: ProviderConfiguration{
			ProviderName:     t.ProviderName,
			Regions:          configRegions,
			ConfigProperties: configProperties,
		},
		ProviderResources: ProviderResources{
			ProviderName: t.ProviderName,
			Resources:    resources,
		},
		DataSources: DataSources{
			ProviderName:        t.ProviderName,
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
		IsOptionalComputed: specSchemaDefinitionProperty.IsOptionalComputed() || specSchemaDefinitionProperty.IsOptionalComputedWithDefault(),
		IsSensitive:        specSchemaDefinitionProperty.Sensitive,
		IsParent:           specSchemaDefinitionProperty.IsParentProperty,
		Description:        specSchemaDefinitionProperty.Description,
		Default:            specSchemaDefinitionProperty.Default,
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
