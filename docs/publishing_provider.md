# Publishing OpenAPI Terraform providers in the Terraform Registry

This document contains information on how API providers can make use of the [OpenAPI Terraform provider library](https://github.com/dikhan/terraform-provider-openapi/blob/master/go.mod#L1) 
to generate their own Terraform provider and register it in the [Terraform Registry](https://registry.terraform.io/).

- [How can I publish my own Terraform provider in the Terraform Registry using the OpenAPI Terraform provider library?](#how-can-i-publish-my-own-terraform-provider-in-the-terraform-registry-using-the-openapi-terraform-provider-library)
  - [Preparing the provider](#preparing-the-provider)
  - [Documenting your Provider](#documenting-your-provider)
  - [Creating a GitHub Release](#creating-a-github-release)
  - [Publishing to the Registry](#publishing-to-the-registry)

## How can I publish my own Terraform provider in the Terraform Registry using the OpenAPI Terraform provider library?

The OpenAPI Terraform provider library makes it very easy for service providers to generate their own Terraform provider leveraging
their API OpenAPI docs. The following instructions will guide you to get a Terraform provider for your API and successfully 
register it into the Terraform Registry. Your users then will be able to make use of the Terraform provider and interact with your APIs through Terraform.  

This instructions expand on the [Release and Publish a Provider to the Terraform Registry](https://learn.hashicorp.com/tutorials/terraform/provider-release-publish) docs
and are more tailored to leveraging the OpenAPI Terraform provider library.

### Preparing the provider

- Create a new repo in your GitHub org and name it following Terraform provider's naming convention `terraform-provider-{NAME}`
- Create a file named main.go which will be the entry point for the Terraform plugin, and the default executable when 
the binary is built. Populate `main.go` with the following code setting the `providerName` and `providerOpenAPIURL` variables
to the corresponding values. The `providerName` must match the `{NAME}` specified in the repository `terraform-provider-{NAME}`
and the `providerOpenAPIURL` should be the URL where the service API OpenAPI docs are exposed. 

````
package main

import (
	"github.com/dikhan/terraform-provider-openapi/v3/openapi"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"log"
)

	// Version specifies the version of the provider (will be set statically at compile time)
	Version = "dev"
	// Commit specifies the commit hash of the provider at the time of building the binary (will be set statically at compile time)
	Commit = "none"
	// Date specifies the data which the binary was build (will be set statically at compile time)
	Date = "unknown"

func main() {

	log.Printf("[INFO] Running Terraform Provider %s v%s-%s; Released on: %s", ProviderName, Version, Commit, Date)

	log.Printf("[INFO] Initializing OpenAPI Terraform provider '%s' with service provider's OpenAPI document: %s", ProviderName, ProviderOpenAPIURL)

	providerName = "openapiexample"
	providerOpenAPIURL = "https://localhost:8443/api/openapi"

	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	serviceProviderConfig := &openapi.ServiceConfigV1{
		SwaggerURL: providerOpenAPIURL,
	}

	provider, err := p.CreateSchemaProviderFromServiceConfiguration(serviceProviderConfig)
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize the terraform provider: %s", err)
	}

	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: func() *schema.Provider {
				return provider
			},
		},
	)
}
````
 
- It's recommended to use go mod to keep track of the go dependencies as well as be able to lock the dependencies versions. This can
be done running `go mod init [module]` where module is the module name you pick. For instance, the following command will
initialize go mod with a module name being `github.com/dikhan/terraform-provider-openapiexample`.

````
$ go mod init github.com/dikhan/terraform-provider-openapiexample
````

The above will generate a `go.mod` file containing all the dependencies the project relies on. To confirm the dependencies are pulled
successfully also run `go tidy`. The resulting `go.mod` file should look like:

````
module github.com/dikhan/terraform-provider-openapiexample

go 1.16

require (
	github.com/dikhan/terraform-provider-openapi/v3 v2.0.6
	github.com/hashicorp/terraform-plugin-sdk/v2 v2.6.1
	github.com/mattn/go-colorable v0.1.8 // indirect
)
````

Note: If you want to use a different version of the `github.com/dikhan/terraform-provider-openapi/v3` library you can update
the version with the commit hash you want to use (eg: a branch that has not been merged to master yet) or even alpha versions
released, etc.

- Build the plugin binary:

````
$ go build -o terraform-provider-openapiexample
````

- Validate that the plugin starts successfully by executing the binary. Confirm that the output shows the resources and data sources
 that are OpenAPI Terraform compliant. For instance, the following output would be produced if the OpenAPI endpoint configured
 previously `https://localhost:8443/api/openapi` returned  and OpenAPI document that has two OpenAPI Terraform compatible resources:
 
````
./terraform-provider-openapiexample
2021/06/13 17:14:31 [INFO] Running Terraform Provider openapiexample dev-none; Released on: 2021-06-13T17:10:57Z-0700
2021/06/13 17:14:31 [INFO] Initializing OpenAPI Terraform provider 'openapiexample' with service provider's OpenAPI document: https://localhost:8443/api/openapi
...
2021/06/13 17:14:31 [INFO] found terraform compliant resource [name='cdn_v1', rootPath='/v1/cdns', instancePath='/v1/cdns/{cdn_id}']
2021/06/13 17:14:31 [INFO] found terraform compliant resource [name='cdn_v1_firewall', rootPath='/v1/cdns/{cdn_id}/firewalls', instancePath='/v1/cdns/{cdn_id}/firewalls/{fw_id}']
...
2021/06/13 17:14:31 [INFO] resource 'openapiexample_cdn_v1' successfully registered in the provider (time:555.42µs)
2021/06/13 17:14:31 [INFO] data source instance 'openapiexample_cdn_v1_instance' successfully registered in the provider (time:576.494µs)
...
2021/06/13 17:14:31 [INFO] resource 'openapiexample_cdn_v1_firewall' successfully registered in the provider (time:537.96µs)
2021/06/13 17:14:31 [INFO] data source instance 'openapiexample_cdn_v1_firewall_instance' successfully registered in the provider (time:558.542µs)

This binary is a plugin. These are not meant to be executed directly.
Please execute the program that consumes these plugins, which will
load any plugins automatically
````

Since Terraform plugins are not expected to be run directly (like we did above), Terraform will stop its execution and at the
end will show something like `This binary is a plugin`. This is expected, and the purpose of this sanity check is to confirm
that the provider does indeed load the resources exposes by the OpenAPI document that are [OpenAPI Terraform compliant.](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md)

### Documenting your Provider

Terraform expects providers to be documented following a specific format and that the documentation follows a certain structure. This is required,
so the documentation follows the same convention as the other plugins registered in the Terraform registry. To assist with
this Terraform developed a helper tool called [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs) to auto-generate Terraform providers docs. 

You can download and install the `tfplugindocs` tool running the following:
````
$ go get github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs
````

Learn more about this in the [Documenting your Provider](https://www.terraform.io/docs/registry/providers/publishing.html#documenting-your-provider) docs.

### Creating a GitHub Release

The Terraform Registry looks up the provider's GitHub Releases to register new versions of the provider in the Terraform Registry. The way the
releases are done is totally up to the service provider. Terraform's preferred way for setting this up is using GitHub Actions and `goreleaser`. 
You can learn more on how to configure this [here](https://www.terraform.io/docs/registry/providers/publishing.html#github-actions-preferred-)

Learn more about this in the [Creating a GitHub Release](https://www.terraform.io/docs/registry/providers/publishing.html#creating-a-github-release) docs.

### Publishing to the Registry

The process for publishing the provider to the Terraform Registry is straightforward. Learn more about this following the
[Publishing to the Registry](https://www.terraform.io/docs/registry/providers/publishing.html#publishing-to-the-registry) docs.