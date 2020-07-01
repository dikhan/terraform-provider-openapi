# OpenAPI Terraform Documentation Renderer

This library generates the Terraform documentation automatically given an already Terraform compatible OpenAPI document. 

## How to use this library?

The main.go show cases how the openapi.TerraformProviderDocGenerator can be configured so it's able to render the markdown
given a swagger file.

## How to run the example?

The main.go file is configured with a sample swagger file. The demo can be executed simply by running the following command:

````
$ go run main.go
````

The program will spit out the Terraform documentation for the sample swagger file in the standard output

````
## Provider Configuration
The provider (cloudflare) provider is used to interact with the many resources supported by cloudflare's API. The provider may need to be configured with the proper credentials before it can be used. Refer to the provider confgiguration's arguments to learn more about the required configuration properties.

#### Example usage

provider "cloudflare" {
    region = "sea"
    api_auth = "value"
    required_header_example = "value"
}


#### Arguments Reference
The following arguments are supported:

- region [string] (required): The core data center location to be used ([sea dub rst fra]). If region isn't specified, the default is 'sea'.
- api_auth [string] (required): 
- required_header_example [string] (required): 

## Provider Resources

### cloudflare_cdn

#### Example usage

resource "cloudflare_cdn" "my_cdn" {
    label = "string value"
}


#### Arguments Reference (input)
The following arguments are supported:

- label [string] (required): Label to use for the CDN

#### Attributes Reference (output)
In addition to all arguments above, the following attributes are exported:

- id [string]: System generated identifier for the CDN
````
