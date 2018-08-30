# Terraform Provider OpenAPI [![Build Status][travis-image]][travis-url]

This terraform provider aims to minimise as much as possible the efforts needed from service providers to create and
maintain custom terraform providers. This provider uses terraform as the engine that will orchestrate and manage the cycle
of the resources and depends on a swagger file (hosted on a remote endpoint) to successfully configure itself dynamically at runtime.

<center>
    <table cellspacing="0" cellpadding="0" style="width:100%; border: none;">
      <tr>
        <th align="center"><img src="https://cdn.rawgit.com/hashicorp/terraform-website/master/content/source/assets/images/logo-hashicorp.svg" width="400px"></th>
        <th align="center"><img src="https://www.openapis.org/wp-content/uploads/sites/3/2018/02/OpenAPI_Logo_Pantone-1.png" width="400px"></th> 
      </tr>
      <tr>
        <td align="center"><p>Powered by <a href="https://www.terraform.io">https://www.terraform.io</a></p></td>
        <td align="center"><p>Following <a href="https://github.com/OAI/OpenAPI-Specification">The OpenAPI Specification</a></td> 
      </tr>
    </table>
</center>

What are the main pain points that this terraform provider tries to tackle?

- As as service provider, you can focus on improving the service itself rather than the tooling around it.
- Due to the dynamic nature of this terraform provider, the service provider can continue expanding the functionality
of the different APIs by introducing new versions, and this terraform provider will be able to discover the new resource versions automatically without the need to add support for those as you would when mantining your own custom Terraform provider.
- Find consistency across APIs provided by different teams encouraging the adoption of OpenAPI specification for
describing, producing, consuming, and visualizing RESTful Web services.

## Overview

API terraform provider is a powerful full-fledged terraform provider that is able to configure itself at runtime based on 
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

- Terraform expects third party providers to be manually installed in the '.terraform.d/plugins' sub-path in your user's home directory.
- Terraform expects terraform provider names to follow a specific naming scheme. The naming scheme for plugins is 
terraform-<type>-NAME_vX.Y.Z, where type is either provider or provisioner. 

More information about how terraform discovers third party terraform providers and naming conventions [here](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery).

### OpenAPI Terraform provider installation

Installing the OpenAPI Terraform provider can be achieved in various ways, but for the sake of simplicity below are
the suggested options:

#### OpenAPI Terraform provider 'manual' installation

- Download most recent release for your architecture (macOS/Linux) from [release](https://github.com/dikhan/terraform-provider-openapi/releases) 
page.
- Extract contents of tar ball and copy the terraform-provider-openapi binary into your  ````~/.terraform.d/plugins````
folder as described in the [Terraform documentation on how to install plugins](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery).
- After installing the plugin, you have two options. Either:
 
  - Rename the binary file to have your provider's name:
 
    ````
    $ cd ~/.terraform.d/plugins
    $ mv terraform-provider-openapi terraform-provider-<your_provider_name>
    $ ls -la 
    total 29656
    drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
    drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
    -rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-<your_provider_name>
    ````
 
  - Create a symlink pointing at the terraform-provider-openapi binary. The latter is recommended so the same compiled binary 
'terraform-provider-openapi' can be reused by multiple openapi providers and also reduces the number of providers to support.

    ````
    $ cd ~/.terraform.d/plugins
    $ ln -sF terraform-provider-openapi terraform-provider-<your_provider_name>
    $ ls -la 
    total 29656
    drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
    drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
    -rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-openapi
    lrwxr-xr-x  1 dikhan  staff        63  3 Jul 15:11 terraform-provider-<your_provider_name> -> /Users/dikhan/.terraform.d/plugins/terraform-provider-openapi
    ````

Where ````<your_provider_name>```` should be replaced with your provider's name. This is the name that will also be used
in the Terraform tf files to refer to the provider resources. 

````
$ cat main.tf
resource "<your_provider_name>_<resource_name>" "my_resource" {
    ...
    ...
}
````

#### OpenAPI Terraform provider 'script' installation

In order to simplify the installation process for this provider, a convenient install script is provided and can also be 
used as follows:

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

### Using the OpenAPI Terraform provider

Once the OpenAPI terraform plugin is installed, you can go ahead and define a tf file that has resources exposed
by your service provider. 

The example below describes a resource of type 'bottle' provided by the 'goa' service provider. For full details about this
example refer to [goa example](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/goa).

````
$ cat main.tf
resource "goa_bottles" "my_bottle" {
  name = "Name of bottle"
  rating = 3
  vintage = 2653
}
````

The OpenAPI terraform provider relies on the swagger file exposed by the service provider. This information can be provided
to the plugin in two different ways:

#### OTF_VAR_<provider_name>_SWAGGER_URL

Terraform will need to be executed passing in the OTF_VAR_<provider_name>_SWAGGER_URL environment variable pointing at the location 
where the swagger file is hosted, where````<your_provider_name>```` should be replaced with your provider's name. 

    ```
    $ terraform init && OTF_VAR_goa_SWAGGER_URL="https://some-domain-where-swagger-is-served.com/swagger.yaml" terraform plan
    ```

#### OpenAPI plugin configuration file

A configuration file will need to be created in terraform plugins folder ```~/.terraform.d/plugins``` following [OpenAPI v1 plugin configuration specification](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md).
An example is described below:

    ```
    $ pwd
    /Users/dikhan/.terraform.d/plugins
    $ cat terraform-provider-openapi.yaml
    version: '1'
    services:
        monitor:
          swagger-url: http://monitor-api.com/swagger.json
          insecure_skip_verify: true
        cdn:
          swagger-url: /Users/user/go/src/github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen/api/resources/swagger.yaml
        vm:
          swagger-url: http://vm-api.com/swagger.json
        goa: 
          swagger-url: https://some-domain-where-swagger-is-served.com/swagger.yaml
    ```

This option is the recommended one when the user is managing resources provided by multiple OpenAPI providers (e,g: goa and swaggercodegen),
since it minimizes the configuration needed when calling terraform. Hence, terraform could be executed as usual without 
having to pass in any special environment variables like OTF_VAR_<provider_name>_SWAGGER_URL:

    ```
    $ terraform init && terraform plan
    ```

## Examples

Two API examples compliant with terraform are provided to make it easier to play around with this terraform provider. This
examples can be found in the [examples folder](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples) 
and each of them provides details on how to bring up the service and run this provider against the APIs using terraform.

- [goa](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/goa): Example created using goa framework. 
This API exposes a resource called 'bottles'
- [swaggercodegen](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/swaggercodegen): Example 
created using swaggercodegen. This API exposes a resource called 'cdns'

For more information refer to [How to set up the local environment?](./docs/local_environment.md) which contains instructions
for learning how to bring up the example APIs and run terraform against them.

## References

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
