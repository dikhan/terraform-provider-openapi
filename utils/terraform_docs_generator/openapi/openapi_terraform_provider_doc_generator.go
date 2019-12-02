package openapi

import (
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/printers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type TerraformProviderDocGenerator struct {
	ProviderName  string
	OpenAPIDocURL string
	Printer       printers.Printer
}

func (o TerraformProviderDocGenerator) generateDocumentation() error {
	p := openapi.ProviderOpenAPI{ProviderName: o.ProviderName}
	serviceConfig := &openapi.ServiceConfigV1{SwaggerURL: o.OpenAPIDocURL}
	tfProvider, err := p.CreateSchemaProviderFromServiceConfiguration(serviceConfig)
	if err != nil {
		return err
	}
	for resourceName, resource := range tfProvider.ResourcesMap {
		o.printResourceDoc(resourceName, *resource)
	}
	return nil
}

func (o TerraformProviderDocGenerator) printResourceDoc(resourceName string, resource schema.Resource) {
	required, optional, computed := o.createRequiredOptionalComputedMaps(resource.Schema)
	o.Printer.PrintResourceInfo(o.ProviderName, resourceName)
	o.Printer.PrintResourceExample(o.ProviderName, resourceName, required)
	o.Printer.PrintArguments(required, optional)
	o.Printer.PrintAttributes(computed)
}

func (o TerraformProviderDocGenerator) createRequiredOptionalComputedMaps(s map[string]*schema.Schema) (map[string]*schema.Schema, map[string]*schema.Schema, map[string]*schema.Schema) {
	required := map[string]*schema.Schema{}
	optional := map[string]*schema.Schema{}
	computed := map[string]*schema.Schema{}
	for k, v := range s {
		if v.Required {
			required[k] = v
		} else {
			if v.Computed {
				computed[k] = v
			} else {
				optional[k] = v
			}
		}
	}
	return required, optional, computed
}
