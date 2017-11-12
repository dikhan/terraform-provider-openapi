# API Terraform Provider

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

## Give it a try

Refer to [How to set up the local environment](./docs/local_environment.md) to give it a try

## How to use it

Refer to [How to](docs/how_to.md) to learn more about how to define the swagger file, good practises
and what is currently supported.

## Contributing
Please follow the guidelines from:

 - [Contributor Guidelines](.github/CONTRIBUTING.md)

## FAQ

Refer to [FAQ](./docs/faq.md) document to get answers for the most frequently asked questions.

## References

- [go-swagger](https://github.com/go-swagger/go-swagger): Api terraform provider makes extensive use of this library 
which offers a very convenient implementation to serialize and deserialize swagger specifications.

## Authors

- Daniel I. Khan Ramiro

See also the list of [contributors](https://github.com/dikhan/terraform-provider-api/graphs/contributors) who participated in this project.