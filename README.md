# Terraform Provider OpenAPI [![Build Status][travis-image]][travis-url] [![GoDoc][godoc-badge]][godoc-url] [![GoReportCard][goreportcard-badge]][goreportcard-url] [![CodeCov][codecov-badge]][codecov-url]

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

- The service provider hosts APIs documented using [OpenApi 2.0 specification](https://swagger.io/specification/v2/) and the APIs
comply with the OpenAPI Terraform Provider [How to](docs/how_to.md) guidelines. The service provider API's OpenAPI document must also 
be available via a discovery endpoint served through HTTP/s or the file system.

### Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= v0.12.0 (to execute the terraform provider plugin)
  - If you are using Terraform 0.11, refer to the latest [OpenAPI Terraform provider v0.13.1 released](https://github.com/dikhan/terraform-provider-openapi/releases/tag/v0.31.1) compatible with Terraform 0.11.
- [Go](https://golang.org/doc/install) >=1.14 (to build the provider plugin)
  - This project uses [go modules](https://github.com/golang/go/wiki/Modules) for dependency management
- [Docker](https://www.docker.com/) 17.09.0-ce (to run service provider example)
- [Docker-compose](https://docs.docker.com/compose/) 1.16.1 (to run service provider example)


## How to use Terraform Provider OpenAPI

### Things to know regarding custom terraform providers

- Terraform expects third party (in-house) providers to be manually installed in a specific directory. Refer to the [OpenAPI Terraform provider installation instructions](#openapi-terraform-provider-installation) to
learn more about this.
- Terraform expects terraform provider names to follow a specific naming scheme. The naming scheme for plugins is 
``terraform-<type>-<name>_vX.Y.Z``, where `<type>` is either provider or provisioner, `<name>` is the provider's name and `X.Y.Z` is the version of the plugin.

More information about how terraform discovers third party terraform providers and naming conventions [here](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery).

### OpenAPI Terraform provider installation

There are multiple ways how the OpenAPI Terraform provider can be installed. Please refer to the [OpenAPI Terraform provider installation document](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/installing_openapi_provider.md)
to learn more about it.

### OpenAPI Terraform provider in action

After having provisioned your environment with the OpenAPI Terraform provider you can now write Terraform configuration files using resources provided
by the OpenAPI service. Refer to [Using the OpenAPI Terraform Provider doc](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md) for more details.

### Terraform provider documentation

You can generate the Terraform documentation automatically given an already Terraform compatible OpenAPI document using the The [OpenAPI Terraform Documentation Renderer](https://github.com/dikhan/terraform-provider-openapi/tree/master/pkg/terraformdocsgenerator) 
library. The OpenAPI document is the source of truth for both the OpenAPI Terraform provider as well as the user facing documentation.

## References

Additionally, the following documents provide deep insight regarding OpenAPI and Terraform as well as frequently asked questions:

- [How to](docs/how_to.md) document contains information about how to define a OpenAPI document following good practises that
make it work seamlessly with this terraform provider. Additionally, learn more about what is currently supported.
- [Migrating to Terraform 0.12](./docs/terraform_version_upgrades/upgrading_to_terraform_0.12.md). This document describes
how to update configuration created using Terraform v0.11 to v0.12.
- [FAQ](./docs/faq.md) document answers for the most frequently asked questions.

## Contributing

Please follow the guidelines from:

 - [Contributor Guidelines](.github/CONTRIBUTING.md)
 - [How to set up the local environment?](./docs/local_environment.md)

## References

- [go-swagger](https://github.com/go-swagger/go-swagger): Api terraform provider makes extensive use of this library 
which offers a very convenient implementation to serialize and deserialize swagger specifications.
- [JsonPath](https://github.com/oliveagle/jsonpath): Json path is used in
the plugin external configuration file to define values for provider schema
properties that are coming from external files.

## Authors

- Daniel I. Khan Ramiro

See also the list of [contributors](https://github.com/dikhan/terraform-provider-api/graphs/contributors) who participated in this project.


[travis-url]: https://travis-ci.org/dikhan/terraform-provider-openapi
[travis-image]: https://travis-ci.org/dikhan/terraform-provider-openapi.svg?branch=master

[godoc-url]: https://godoc.org/github.com/dikhan/terraform-provider-openapi
[godoc-badge]: http://img.shields.io/badge/godoc-reference-5272B4.svg?style=flat-square

[goreportcard-url]: https://goreportcard.com/report/github.com/dikhan/terraform-provider-openapi
[goreportcard-badge]: https://goreportcard.com/badge/github.com/dikhan/terraform-provider-openapi?style=flat-square

[codecov-url]: https://codecov.io/gh/dikhan/terraform-provider-openapi
[codecov-badge]: https://codecov.io/gh/dikhan/terraform-provider-openapi/branch/master/graph/badge.svg