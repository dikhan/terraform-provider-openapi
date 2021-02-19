package main

import (
	"context"
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"

	"fmt"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi"
	"github.com/dikhan/terraform-provider-openapi/v2/openapi/version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"os"
	"regexp"
)

// Source addresses consist of three parts delimited by slashes (/), as follows: [<HOSTNAME>/]<NAMESPACE>/<TYPE>
// - Hostname (optional): The hostname of the Terraform registry that distributes the provider. If omitted, this defaults to registry.terraform.io, the hostname of the public Terraform Registry.
// - Namespace: An organizational namespace within the specified registry. For the public Terraform Registry and for Terraform Cloud's private registry, this represents the organization that publishes the provider. This field may have other meanings for other registry hosts.
// - Type: A short name for the platform or system the provider manages. Must be unique within a particular namespace on a particular registry host.

var otfProviderSourceAddressVar = "OTF_PROVIDER_SOURCE_ADDRESS"

func main() {

	log.Printf("Running OpenAPI Terraform Provider v%s-%s; Released on: %s", version.Version, version.Commit, version.Date)

	var debugMode bool
	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	binaryName, err := os.Executable()
	log.Printf("[INFO] Terraform is executing the following OpenAPI Terraform provider plugin: %s", binaryName)
	if err != nil {
		log.Fatalf("[ERROR] There was an error when getting the provider binary name: %s", err)
	}

	provider, err := initProvider(binaryName)
	if err != nil {
		log.Fatalf("[ERROR] %s", err)
	}

	if debugMode {
		// A provider's source address is its global identifier. It also specifies the primary location where Terraform can download it.
		// The value of this variable must match the source value provided in the terraform's configuration. For instance,
		// the below example declares a local provider name 'openapi' where the source is 'terraform.example.com/examplecorp/openapi'
		// terraform {
		//  required_providers {
		//    openapi = {
		//      source  = "terraform.example.com/examplecorp/openapi"
		//      version = "~> 1.0"
		//    }
		//  }
		//}
		// With the above terraform configuration the expected value for OTF_PROVIDER_SOURCE_ADDRESS environment variable is 'terraform.example.com/examplecorp/openapi'
		// and the plugin should be installed in any of the expected implied local mirror directories (https://www.terraform.io/docs/commands/cli-config.html#implied-local-mirror-directories).
		// For example, in a linux distribution that would be: $HOME/.terraform.d/plugins/terraform.example.com/examplecorp/openapi/1.0.0/linux_amd64/terraform-provider-openapi
		// For more info about source address read https://www.terraform.io/docs/configuration/provider-requirements.html#source-addresses
		// and In-house providers installation read https://www.terraform.io/docs/configuration/provider-requirements.html#in-house-providers
		providerSourceAddress := os.Getenv(otfProviderSourceAddressVar)
		if providerSourceAddress == "" {
			log.Fatalf("[ERROR] Could not start the provider '%s' in debug mode due to missing required environment variable %s", binaryName, otfProviderSourceAddressVar)
		}

		err := plugin.Debug(context.Background(), providerSourceAddress,
			&plugin.ServeOpts{
				ProviderFunc: func() *schema.Provider {
					return provider
				},
			})
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: func() *schema.Provider {
				return provider
			}})
	}
}

func initProvider(binaryName string) (*schema.Provider, error) {
	providerName, err := getProviderName(binaryName)
	if err != nil {
		return nil, fmt.Errorf("error getting the provider's name from the binary '%s': %s", binaryName, err)
	}
	log.Printf("[INFO] Initializing '%s' provider", providerName)
	p := openapi.ProviderOpenAPI{ProviderName: providerName}
	provider, err := p.CreateSchemaProvider()
	if err != nil {
		return nil, fmt.Errorf("error initialising the terraform provider: %s", err)
	}
	return provider, nil
}

func getProviderName(binaryName string) (string, error) {
	r, err := regexp.Compile("\\bterraform-provider-([a-zA-Z0-9]+)(?:_v[\\d]+\\.[\\d]+\\.[\\d]+)?\\b$")
	if err != nil {
		return "", err
	}

	match := r.FindStringSubmatch(binaryName)
	if len(match) != 2 {
		return "", fmt.Errorf("provider binary name (%s) does not match terraform naming convention 'terraform-provider-{name}', please rename the provider binary", binaryName)
	}
	return match[1], nil
}
