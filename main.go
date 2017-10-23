package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"github.com/dikhan/terraform-provider-api/terraform_provider_api"
)

func main() {
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: func() terraform.ResourceProvider {
				return terraform_provider_api.ApiProvider()
			},
		})
}
