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

## Local environment

A [docker-compose](docker-compose.yml) file has been created to ease the execution of an example. In order to bring up 
the service provider example and also render a UI from the swagger file that can be accessed from the browser, please 
run the following command from the root folder:

```
docker-compose up --build --force-recreate
```

Once docker-compose is done bringing up both services, the following command will read the sample [main.tf](terraform_provider_api/main.tf) 
file and execute terraform plan:  
```
go build -o terraform-provider-sp && terraform init && terraform plan
```

Looking carefully at the above command, the binary is named as 'terraform-provider-sp'. The reason for this is so
terraform knows what provider binary it should call when creating resources for 'sp' provider as defined in [main.tf](terraform_provider_api/main.tf) 
file. 

After executing terraform plan, the expected output should be:

```

$ go build -o terraform-provider-sp && terraform init && terraform plan

....

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + sp_cdns_v1.my_cdn
      id:          <computed>
      hostnames.#: "1"
      hostnames.0: "origin.com"
      ips.#:       "1"
      ips.0:       "127.0.0.1"
      label:       "label"


Plan: 1 to add, 0 to change, 0 to destroy.

....

```

This means that the plugin was able to read the swagger file exposed by the service provider example, load it
up and set up the terraform provider on the fly with the resources exposed by 'cdn-service-provider-api' being one of
them 'cdns'.

Now we can run apply to see the plugin do its magic:

```
$ go build -o terraform-provider-sp && terraform init && terraform apply

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
sp_cdns_v1.my_cdn: Creating...
  hostnames.#: "0" => "1"
  hostnames.0: "" => "origin.com"
  ips.#:       "0" => "1"
  ips.0:       "" => "127.0.0.1"
  label:       "" => "label"
sp_cdns_v1.my_cdn: Creation complete after 0s (ID: 80514498-a4d0-44e6-ad0d-22ac1023fdae)

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

And a 'terraform.tfstate' should have been created by terraform containing the state of the new resource created.

## Contributing
Please follow the guidelines from:

 - [Contributor Guidelines](./docs/contributing.md)

## FAQ

Refer to [FAQ](./docs/faq.md) document to get answers for the most frequently asked questions.

## References

- [go-swagger](https://github.com/go-swagger/go-swagger): Api terraform provider makes extensive use of this library 
which offers a very convenient implementation to serialize and deserialize swagger specifications.

## Authors

- Daniel I. Khan Ramiro

See also the list of [contributors](https://github.com/dikhan/terraform-provider-api/graphs/contributors) who participated in this project.