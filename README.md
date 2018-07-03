# Terraform Provider OpenAPI [![Build Status][travis-image]][travis-url]

This terraform provider aims to minimise as much as possible the efforts needed from service providers to create and
maintain custom terraform providers. This provider uses terraform as the engine that will orchestrate and manage the cycle
of the resources and depends on a swagger file (hosted on a remote endpoint) to successfully configure itself dynamically.

<center>
    <table cellspacing="0" cellpadding="0" style="width:100%; border: none;">
      <tr>
        <th align="center"><img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="600px"></th>
        <th align="center"><img src="https://goo.gl/QUpyCh" width="150px"></th> 
      </tr>
      <tr>
        <td align="center"><p>Powered by <a href="https://www.terraform.io">https://www.terraform.io</a></p></td>
        <td align="center"><p>Powered by <a href="swagger.io">swagger.io</a></td> 
      </tr>
    </table>
</center>

What are the main pain points that this terraform provider tries to tackle?

- As as service provider, you can focus on improving the service itself rather than the tooling around it.
- Due to the dynamic nature of this terraform provider, the service provider can continue expanding the functionality
of the different APIs by introducing new versions, and this terraform provider will do the rest configuring the
resources based on the resources exposed and their corresponding versions.
- Find consistency across APIs provided by different teams encouraging the adoption of OpenAPI specification for
describing, producing, consuming, and visualizing RESTful Web services.

## Overview

API terraform provider is a powerful full-fledged terraform provider that is able to configure itself on runtime based on 
a [Swagger](https://swagger.io/) specification file containing the definitions of the APIs exposed. The dynamic nature of 
this provider is what makes it very flexible and convenient for service providers as subsequent upgrades 
to their APIs will not require new compilations of this provider. 
The service provider APIs are discovered on the fly and therefore the service providers can focus on their services
rather than the tooling around it.  


### Pre-requirements

-   The service provider hosts APIs compliant with OpenApi and swagger spec file is available via a discovery endpoint.

### Requirements

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x (to execute the terraform provider plugin)
-	[Go](https://golang.org/doc/install) 1.9 (to build the provider plugin)
-	[Docker](https://www.docker.com/) 17.09.0-ce (to run service provider example)
-	[Docker-compose](https://docs.docker.com/compose/) 1.16.1 (to run service provider example)


## How to use Terraform Provider OpenAPI

### Things to know regarding custom terraform providers

- Terraform expects Third-party providers to be manually installed in the sub-path .terraform.d/plugins in your user's home directory.
- Terraform expects terraform provider names to follow a specific naming scheme. The naming scheme for plugins is 
terraform-<type>-NAME_vX.Y.Z, where type is either provider or provisioner. 

More information about how terraform discovers third-party terraform providers and naming conventions [here](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery).

### OpenAPI Terraform provider installation

 In order to simplify the installation process for this provider, a convenient install script is provided and can be used
as follows:

- Check out this repo and execute the install script:

````
$ git clone git@github.com:dikhan/terraform-provider-openapi.git
$ cd ./scripts
$ PROVIDER_NAME=goa ./install.sh --provider-name $PROVIDER_NAME
````

- Or directly by downloading the install script using curl:

````
$ export PROVIDER_NAME=goa && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME
````

The install script will download the most recent [terraform-provider-openapi release](https://github.com/dikhan/terraform-provider-openapi/releases) 
and install it in the terraform plugins folder ````~/.terraform.d/plugins```` as described above. The terraform plugins 
folder should contain both the terraform-provider-openapi provider and the provider's symlink pointing at it.

````
$ ls -la ~/.terraform.d/plugins
total 29656
drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
-rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-openapi
lrwxr-xr-x  1 dikhan  staff        63  3 Jul 15:11 terraform-provider-goa -> /Users/dikhan/.terraform.d/plugins/terraform-provider-openapi
````

### OpenAPI Terraform provider 'manual' installation

Alternatively, you can also follow a more manual approach and compile/install the provider yourself.

The make file has a target that simplifies this step in the following command:

```
$ git clone git@github.com:dikhan/terraform-provider-openapi.git
$ PROVIDER_NAME="goa" make install
[INFO] Building terraform-provider-openapi binary
[INFO] Creating /Users/dikhan/.terraform.d/plugins if it does not exist
[INFO] Installing terraform-provider-goa binary in -> /Users/dikhan/.terraform.d/plugins
```

The above ```make install``` command will compile the provider from the source code, install the compiled binary terraform-provider-openapi 
in the terraform plugin folder ````~/.terraform.d/plugins```` and create a symlink from terraform-provider-goa to the
binary compiled. The reason why a symlink is created is so the same compiled binary can be reused by multiple openapi providers 
and also reduces the number of providers to support.

### Creating resources using the OpenAPI Terraform provider

Once the OpenAPI terraform plugin is installed, you can go ahead and define a tf file that has resources exposed
by your service provider. 

The example below describes a resource of type bottle provided by the 'goa' service provider. For full details about this
example refer to [goa example](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/goa).

````
$ cat main.tf
resource "goa_bottles" "my_bottle" {
  name = "Name of bottle"
  rating = 3
  vintage = 2653
}
````

The OpenAPI terraform provider relies on the swagger file exposed by the service provider. In this example, the 'goa' service 
provider exposes an end point that returns the [swagger documentation](https://github.com/dikhan/terraform-provider-openapi/blob/master/examples/goa/api/swagger/swagger.yaml) 
and exposes an API to manage resources of type 'bottles'.

In order to run the OpenAPI terraform provider, simply pass in the OTF_VAR_```<provider_name>```_SWAGGER_URL environment variable
to terraform pointing at the URL where the swagger doc is exposed. ```<provider_name>``` is your provider's name.

```
$ terraform init && OTF_VAR_goa_SWAGGER_URL="https://some-domain-where-swagger-is-served.com/swagger.yaml" terraform plan
```

Below is an output of the execution using the example openapi terraform provider (named 'goa'). 

````

$ cd examples/goa && terraform init && OTF_VAR_goa_SWAGGER_URL="http://localhost:9090/swagger/swagger.yaml" terraform plan

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + goa_bottles.my_bottle
      id:      <computed>
      name:    "Name of bottle"
      rating:  "3"
      vintage: "2653"


Plan: 1 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.

````

For more information refer to [How to set up the local environment?](./docs/local_environment.md) which contains instructions
for learning how to bring up the example API service and run terraform against it.

Additionally, the following documents provide deep insight regarding OpenAPI and Terraform as well as frequently asked questions:

- [How to](docs/how_to.md) document contains information about how to define a swagger file following good practises that
make it work seamlessly with this terraform provider. Additionally, learn more about what is currently supported.
- [FAQ](./docs/faq.md) document answers for the most frequently asked questions.

## Contributing
Please follow the guidelines from:

 - [Contributor Guidelines](.github/CONTRIBUTING.md)
 - [How to set up the local environment?](./docs/local_environment.md)

## References

- [go-swagger](https://github.com/go-swagger/go-swagger): Api terraform provider makes extensive use of this library 
which offers a very convenient implementation to serialize and deserialize swagger specifications.

## Authors

- Daniel I. Khan Ramiro

See also the list of [contributors](https://github.com/dikhan/terraform-provider-api/graphs/contributors) who participated in this project.


[travis-url]: https://travis-ci.org/dikhan/terraform-provider-openapi
[travis-image]: https://travis-ci.org/dikhan/terraform-provider-openapi.svg?branch=master
