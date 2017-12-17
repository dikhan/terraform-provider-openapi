package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"log"
)

func main() {
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: func() terraform.ResourceProvider {
				provider, err := ApiProvider()
				if err != nil {
					log.Printf("[ERROR] There was an error initialising the terraform provider: %s", err)
				}
				return provider
			},
		})
}
