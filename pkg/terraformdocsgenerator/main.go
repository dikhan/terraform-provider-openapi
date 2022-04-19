package main

import (
	"github.com/dikhan/terraform-provider-openapi/v3/pkg/terraformdocsgenerator/openapiterraformdocsgenerator"
	"log"
	"os"
)

func main() {
	providerName := "openapi"
	openAPIDocURL := "https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/examples/swaggercodegen/api/resources/swagger.yaml"

	// Note: The service provider is responsible for installing the in-house provider in the corresponding location that Terraform expects
	// plugins to be located at. Terraform 0.12 and Terraform >=0.13 have different requirements in terms of where the custom plugins should be
	// installed. The following script may be used to automate the plugin provisioning task which depending on what version of Terraform the
	// user is using the plugin will be installed accordingly in the expected location for Terraform:
	// https://github.com/dikhan/terraform-provider-openapi/blob/master/scripts/install.sh
	// The NewTerraformProviderDocGenerator requires the provider name to be passed in, as well as the hostname and namespace which are used
	// to render the provider installation section containing the required_providers block with the source address configuration in the form of [<HOSTNAME>/]<NAMESPACE>/<TYPE>
	terraformProviderDocGenerator, err := openapiterraformdocsgenerator.NewTerraformProviderDocGenerator(providerName, "terraform.example.com", "examplecorp", openAPIDocURL)
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
