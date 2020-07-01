package main

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/pkg/terraformdocsgenerator/openapiterraformdocsgenerator"
	"log"
	"os"
)

func main() {
	terraformProviderDocGenerator, err := openapiterraformdocsgenerator.NewTerraformProviderDocGenerator("openapi", "https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/examples/swaggercodegen/api/resources/swagger.yaml")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("./provider_documentation.html")
	if err != nil {
		log.Fatal(err)
	}

	d, err := terraformProviderDocGenerator.GenerateDocumentation()
	if err != nil {
		log.Fatal(err)
	}

	d.ProviderInstallation.Other = fmt.Sprintf("You will need to be logged in before running Terraform commands that use the '%s' Streamline Terraform provider:", d.ProviderName)

	err = d.RenderHTML(f)
	if err != nil {
		log.Fatal(err)
	}
}
