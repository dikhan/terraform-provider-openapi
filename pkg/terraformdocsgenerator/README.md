# OpenAPI Terraform Documentation Renderer

This library generates the Terraform documentation automatically given an already Terraform compatible OpenAPI document. 

## How to use this library

The library's [main.go](https://github.com/dikhan/terraform-provider-openapi/pkg/terraformdocsgenerator/main.go) show cases how to generate Terraform documentation given a swagger file. Currently, the generator supports rendering documentation in HTML.

Please note that this library uses Go's `text/template` package, which doesn't secure against HTML injection. It's the user's responsibility to ensure that data injected into the `TerraformProviderDocumentation` struct is safe against injection.

## How to run the example

The main.go file is configured with a [sample swagger file]("https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/examples/swaggercodegen/api/resources/swagger.yaml"). The example can be executed simply by running the following command:

````
$ go run main.go
````

The program will generate the Terraform documentation (in html format) for the sample swagger file and save the output locally. An example of the output: [example_provider_documentation_output.html](https://github.com/dikhan/terraform-provider-openapi/blob/master/pkg/terraformdocsgenerator/example_provider_documentation_output.html).

Note: The rendered resources and data sources are ordered alphabetically and should be deterministic, meaning that multiple 
executions of the terraform docs generator for a given OpenAPI document would result into the exact same documentation being rendered. 
However the order in which the endpoints are described in the OpenAPI document may not necessarily match the order of the 
corresponding rendered resources and data sources. Also, it is important to note that if the OpenAPI document is updated with new endpoints that are
terraform compatible the order of the resources and data sources rendered might also change. 

## Customizing the output documentation
You can customize sections of the documentation by overriding the default content used by `GenerateDocumentation()` before calling `RenderHTML()`.

For example, in [main.go](https://github.com/dikhan/terraform-provider-openapi/pkg/terraformdocsgenerator/main.go) we are adding a custom provider installation instruction for the user to login first with the following:
```
d.ProviderInstallation.Other = fmt.Sprintf("You will need to be logged in before running Terraform commands that use the '%s' Streamline Terraform provider:", d.ProviderName)
```
