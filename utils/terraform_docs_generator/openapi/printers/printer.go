package printers

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

type Printer interface {
	PrintResourceInfo(providerName, resourceName string)
	PrintResourceExample(providerName, resourceName string, required map[string]*schema.Schema)
	PrintArguments(required, optional map[string]*schema.Schema)
	PrintAttributes(computed map[string]*schema.Schema)
}
