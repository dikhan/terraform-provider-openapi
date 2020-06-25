package main

import (
	"github.com/dikhan/terraform-provider-openapi/utils/terraform_docs_generator/openapi"
	"log"
	"os"
)

func main() {
	terraformProviderDocGenerator, err := openapi.NewTerraformProviderDocGenerator("openapi", "")
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("./utils/terraform_docs_generator/openapi/templates/zendesk_output.html")
	if err != nil {
		log.Fatal(err)
	}

	d, err := terraformProviderDocGenerator.GenerateDocumentation()
	if err != nil {
		log.Fatal(err)
	}
	err = d.RenderZendeskHTML(f)
	if err != nil {
		log.Fatal(err)
	}
}
