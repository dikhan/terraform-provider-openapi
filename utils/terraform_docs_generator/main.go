package main

import (
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi/printers"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func main() {

	providerName := "openapi"
	resourceName := "resource"
	s := map[string]*schema.Schema{
		"requiredInputProp": {
			Required:    true,
			Type:        schema.TypeString,
			Description: "this is the description for the input string required property",
		},
		"optionalInputProp": {
			Optional:    true,
			Type:        schema.TypeInt,
			Description: "this is the description for the input integer optional property",
		},
		"computedProperty": {
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeBool,
			Description: "this is the description for the computed bool property",
		},
	}

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

	var p printers.Printer

	// Use Markdown printer
	p = printers.MarkdownPrinter{}

	p.PrintResourceInfo(providerName, resourceName)
	p.PrintResourceExample(providerName, resourceName, required)
	p.PrintArguments(required, optional)
	p.PrintAttributes(computed)
}
