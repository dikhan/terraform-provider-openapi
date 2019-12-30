package printers

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
)

type MarkdownPrinter struct{}

func (p MarkdownPrinter) PrintProviderConfigurationHeader() {
	fmt.Println("## Provider Configuration")
	fmt.Println()
}

type MultiRegionConfiguration struct {
	Regions       []string
	DefaultRegion string
}

func (p MarkdownPrinter) PrintProviderConfigurationExample(providerName string, multiRegionConfiguration *MultiRegionConfiguration, requiredSecuritySchemes, requiredHeaders []string) {
	fmt.Println("#### Example usage")
	fmt.Println("````")
	fmt.Printf("provider \"%s\" {\n", providerName)
	if multiRegionConfiguration != nil {
		fmt.Printf("    region = \"%s\"\n", multiRegionConfiguration.DefaultRegion)
	}
	if requiredSecuritySchemes != nil {
		for _, securityScheme := range requiredSecuritySchemes {
			fmt.Printf("    %s = \"value\"\n", securityScheme)
		}
	}
	if requiredHeaders != nil {
		for _, header := range requiredHeaders {
			fmt.Printf("    %s = \"value\"\n", header)
		}
	}
	fmt.Println(`}`)
	fmt.Println("````")
	fmt.Println()
}

func (p MarkdownPrinter) PrintProviderConfiguration(multiRegionConfiguration *MultiRegionConfiguration, requiredSecuritySchemes, requiredHeaders []string) {
	fmt.Println("#### Arguments Reference")
	fmt.Println("The following arguments are supported:")
	fmt.Println()
	if multiRegionConfiguration != nil {
		fmt.Printf("- region [string] (required): The core data center location to be used (%s). If region isn't specified, the default is '%s'.\n", multiRegionConfiguration.Regions, multiRegionConfiguration.DefaultRegion)
	}
	if requiredSecuritySchemes != nil {
		for _, securityScheme := range requiredSecuritySchemes {
			fmt.Printf("- %s [string] (required): \n", securityScheme)
		}
	}
	if requiredHeaders != nil {
		for _, header := range requiredHeaders {
			fmt.Printf("- %s [string] (required): \n", header)
		}
	}
	fmt.Println()
}

func (p MarkdownPrinter) PrintResourceHeader() {
	fmt.Println("## Provider Resources")
	fmt.Println()
}

func (p MarkdownPrinter) PrintResourceInfo(providerName, resourceName string) {
	fmt.Printf("### %s_%s\n", providerName, resourceName)
	// TODO: add support for extension x-terraform-docs-resource-description
	fmt.Println()
}

func (p MarkdownPrinter) PrintResourceExample(providerName, resourceName string, required openapi.SpecSchemaDefinitionProperties) {
	fmt.Println("#### Example usage")
	fmt.Println("````")
	fmt.Printf("resource \"%s_%s\" \"my_%s\" {\n", providerName, resourceName, resourceName)
	for _, property := range required {
		propertyName := property.GetTerraformCompliantPropertyName()
		switch property.Type {
		case openapi.TypeString:
			fmt.Printf("    %s = \"string value\"\n", propertyName)
		case openapi.TypeInt:
			fmt.Printf("    %s = 123\n", propertyName)
		case openapi.TypeBool:
			fmt.Printf("    %s = true\n", propertyName)
		case openapi.TypeFloat:
			fmt.Printf("    %s = 12.99\n", propertyName)
		}
	}
	fmt.Println(`}`)
	fmt.Println("````")
	fmt.Println()
}

func (p MarkdownPrinter) PrintArguments(required, optional openapi.SpecSchemaDefinitionProperties) {
	fmt.Println("#### Arguments Reference (input)")
	fmt.Println("The following arguments are supported:")
	fmt.Println()
	for _, property := range required {
		p.printProperty(property)
	}
	for _, property := range optional {
		p.printProperty(property)
	}
	fmt.Println()
}

func (p MarkdownPrinter) PrintAttributes(computed openapi.SpecSchemaDefinitionProperties) {
	fmt.Println("#### Attributes Reference (output)")
	fmt.Println("In addition to all arguments above, the following attributes are exported:")
	fmt.Println()
	for _, property := range computed {
		p.printProperty(property)
	}
	fmt.Println()
}

func (p MarkdownPrinter) printProperty(property *openapi.SpecSchemaDefinitionProperty) {
	propertyName := property.GetTerraformCompliantPropertyName()
	if property.IsRequired() {
		fmt.Printf("- %s [%s] (required): %s\n", propertyName, property.Type, property.Description)
	} else {
		if property.IsOptionalComputed() {
			fmt.Printf("- %s [%s] (optional): %s\n", propertyName, property.Type, property.Description)
		} else {
			fmt.Printf("- %s [%s]: %s\n", propertyName, property.Type, property.Description)
		}
	}
}
