# Using the OpenAPI Terraform provider

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

The OpenAPI terraform provider relies on the swagger file exposed by the service provider to
configure itself dynamically at runtime. This information can be provided to the plugin in two
different ways:

## OTF_VAR_<provider_name>_SWAGGER_URL

Terraform will need to be executed passing in the OTF_VAR_<provider_name>_SWAGGER_URL environment variable pointing at the location
where the swagger file is hosted, where````<your_provider_name>```` should be replaced with your provider's name.

```
$ terraform init && OTF_VAR_goa_SWAGGER_URL="https://some-domain-where-swagger-is-served.com/swagger.yaml" terraform plan
```

## OpenAPI plugin configuration file

A configuration file can be used to describe multiple OpenAPI service configurations
including where the swagger file is hosted as well as other meatada (e,g: insecure_skip_verify). This
plugin configuration file needs to be placed at ```~/.terraform.d/plugins```. For more information
 about the plugin configuration file read the [OpenAPI v1 plugin configuration specification](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md).

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

# Examples

Two API examples compliant with terraform are provided to make it easier to play around with this terraform provider. This
examples can be found in the [examples folder](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples)
and each of them provides details on how to bring up the service and run this provider against the APIs using terraform.

- [goa](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/goa): Example created using goa framework.
This API exposes a resource called 'bottles'
- [swaggercodegen](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/swaggercodegen): Example
created using swaggercodegen. This API exposes a resource called 'cdns'

Additionally, a convenient make target ``make examples-container`` is provided to bring up a container initialised with terraform and
the example OpenAPI terraform providers (goa and swaggercodegen) already installed. This enables users of this provider to
play around with the OpenAPI providers without messing with their local environments. The following
command will bring up the example APIs, and a container that you can interact with:

 ````
$ make examples-container
$ root@6d7ac292eebd:/openapi# cd goa/
$ root@6d7ac292eebd:/openapi/goa# terraform init && terraform plan
````

For more information refer to [How to set up the local environment?](./docs/local_environment.md) which contains instructions
for learning how to bring up the example APIs and run terraform against them.