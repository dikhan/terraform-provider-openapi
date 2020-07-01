package main

import (
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi_terraform_docs_generator"
	"log"
	"os"
)

func main() {
	terraformProviderDocGenerator, err := openapi_terraform_docs_generator.NewTerraformProviderDocGenerator("openapi", "https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/examples/swaggercodegen/api/resources/swagger.yaml")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("./utils/terraform_docs_generator/provider_documentation.html")
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
