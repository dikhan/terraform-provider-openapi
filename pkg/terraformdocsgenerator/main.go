package main

import (
	"github.com/dikhan/terraform-provider-openapi/v2/pkg/terraformdocsgenerator/openapiterraformdocsgenerator"
	"log"
	"os"
)

func main() {
	providerName := "openapi"
	openAPIDocURL := "https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/examples/swaggercodegen/api/resources/swagger.yaml"

	terraformProviderDocGenerator, err := openapiterraformdocsgenerator.NewTerraformProviderDocGenerator(providerName, openAPIDocURL)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("./example_provider_documentation_output.html")
	if err != nil {
		log.Fatal(err)
	}

	d, err := terraformProviderDocGenerator.GenerateDocumentation()
	if err != nil {
		log.Fatal(err)
	}

	err = d.RenderHTML(f)
	if err != nil {
		log.Fatal(err)
	}
}
