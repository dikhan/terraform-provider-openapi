package main

import (
	"log"

	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: func() terraform.ResourceProvider {
				provider, err := APIProvider()
				if err != nil {
					log.Fatalf("[ERROR] There was an error initialising the terraform provider: %s", err)
				}
				return provider
			},
		})
}
