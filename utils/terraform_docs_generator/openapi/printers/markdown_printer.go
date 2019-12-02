package printers

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type MarkdownPrinter struct{}

func (p MarkdownPrinter) PrintResourceInfo(providerName, resourceName string) {
	fmt.Printf("# Resource %s_%s\n", providerName, resourceName)
	fmt.Println()
}

func (p MarkdownPrinter) PrintResourceExample(providerName, resourceName string, required map[string]*schema.Schema) {
	fmt.Println("## Example")
	fmt.Printf("resource \"%s_%s\" \"my_%s\" {\n", providerName, resourceName, resourceName)
	for k, v := range required {
		switch v.Type {
		case schema.TypeString:
			fmt.Printf("    %s = \"string value\"\n", k)
		case schema.TypeInt:
			fmt.Printf("    %s = 123\n", k)
		case schema.TypeBool:
			fmt.Printf("    %s = true\n", k)
		case schema.TypeFloat:
			fmt.Printf("    %s = 12.99\n", k)
		}
	}
	fmt.Println(`}`)
	fmt.Println()
}

func (p MarkdownPrinter) PrintArguments(required, optional map[string]*schema.Schema) {
	fmt.Println("## Argument Reference (input)")
	for k, v := range required {
		p.printProperty(k, v)
	}
	for k, v := range optional {
		p.printProperty(k, v)
	}
	fmt.Println()
}

func (p MarkdownPrinter) PrintAttributes(computed map[string]*schema.Schema) {
	fmt.Println("## Attributes Reference (output)")
	for k, v := range computed {
		p.printProperty(k, v)
	}
	fmt.Println()
}

func (p MarkdownPrinter) printProperty(propertyName string, s *schema.Schema) {
	if s.Required {
		fmt.Printf("**%s** [%s] (required): %s\n", propertyName, s.Type, s.Description)
	} else {
		if s.Computed {
			// TODO: handle complex objects
			fmt.Printf("**%s** [%s]: %s\n", propertyName, s.Type, s.Description)
		} else {
			fmt.Printf("**%s** [%s] (optional): %s\n", propertyName, s.Type, s.Description)
		}
	}
}
