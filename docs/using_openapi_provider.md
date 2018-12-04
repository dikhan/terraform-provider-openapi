# Using the OpenAPI Terraform provider

## OpenAPI configuration

The OpenAPI terraform provider relies on the swagger file exposed by the service provider to
configure itself dynamically at runtime. This information can be provided to the plugin in two
different ways:

### OTF_VAR_<provider_name>_SWAGGER_URL

Terraform will need to be executed passing in the OTF_VAR_<provider_name>_SWAGGER_URL environment variable pointing at the location
where the swagger file is hosted, where````<your_provider_name>```` should be replaced with your provider's name.

```
$ terraform init && OTF_VAR_goa_SWAGGER_URL="https://some-domain-where-swagger-is-served.com/swagger.yaml" terraform plan
```

### OpenAPI plugin configuration file

A configuration file can be used to describe multiple OpenAPI service configurations
including where the swagger file is hosted as well as other metadata (e,g: insecure_skip_verify).

The plugin configuration file location by default is ```~/.terraform.d/plugins```. However,
this location can be overridden by setting the OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE
environment variable, where '%s' should be replaced with your provider's name.

````
$ export OTF_VAR_myprovider_PLUGIN_CONFIGURATION_FILE="/Users/user/myprovider_config.yaml"
````

The configuration file must comply with the [OpenAPI v1 plugin configuration specification](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md).

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

## OpenAPI Terraform provider configuration

Once the OpenAPI terraform plugin is installed, you can go ahead and define a tf file that has resources exposed
by your service provider.

### Example Usage

The example below describes a resource of type 'cdn_v1' provided by the 'swaggercodegen' service provider. For full details about this
example refer to [goa example](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples/swaggercodegen).

````
provider "swaggercodegen" {
  apikey_auth = "${var.apikey_auth}"
  x_request_id = "request header value for POST /v1/cdns"
}

resource "swaggercodegen_cdn_v1" "my_cdn" {
  label = "label" ## This is an immutable property (refer to swagger file)
  ips = ["127.0.0.1"] ## This is a force-new property (refer to swagger file)
  hostnames = ["origin.com"]
}
````

### Configuration

The OpenAPI provider offers a flexible means of providing credentials for authentication as well as any other header
that might be required by any resource exposed.

- [What can be configured?](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#what-can-be-configured)
- [How can it be configured?](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#how-can-it-be-configured) 

#### What can be configured?

- [Authentication](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#authentication-configuration)
- [Headers](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#headers-configuration)

##### Authentication configuration

The authentication is described in the swagger file under the security definitions section. Additionally,
the security definitions can be attached to the global security scheme making the
authentication mandatory for all the end points exposed (unless overridden in any operation).

In the below example, the swagger file has one security definition named ```apikey_auth```
that defines some basic authentication. On the other hand, there's also a global 'security'
definition which has the ```apikey_auth``` attached.

````
swagger: "2.0"

paths:
 ...
 ...

security:
  - apikey_auth: []

securityDefinitions:
  apikey_auth:
    type: "apiKey"
    name: "Authorization"
    in: "header"
````

The above will translate into the following configuration for the
terraform provider:

````
provider "swaggercodegen" {
  apikey_auth = "..."
}
````

As you can see above, the provider automatically detects that the swagger has a global scheme and automatically exposes that as part of the terraform
provider configuration allowing the user to provider the right values and therefore be able to authenticate properly when creating resources
provider by the aforementioned provider.

Note: A global security scheme makes the authentication required as far as the terraform provider is concerned. If there 
are no global security schemes defined and there are just security definitions, these can also be configured
via the terraform provider but will be optional.

##### Headers configuration

Similarly to the authentication configuration, the provider can also be
configured with headers required by certain operations as described in the [xTerraformHeader](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#xTerraformHeader)

````
swagger: "2.0"

paths:
/resource:
  post:
  ...
  - in: "body"
    ...
  - in: "header"
    name: "X-Request-ID" # This header will be send along with the request when making the POST request against the '/resource' API
    required: true
    x-terraform-header: x_request_id
    responses:
      ...
    ...
  ...
````

#### How can it be configured?

The following methods to configure the properties of the OpenAPI provider are supported, in this order, and explained below:

- [Static credentials](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#static-credentials)
- [Environment variables](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#environment-variables)
- [Shared OpenAPI Plugin Configuration file](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#shared-openapi-plugin-configuration-file)

The above will only be available if the swagger file describes the authentication and headers accordingly as described below.

##### Static credentials

Static credentials can be provided by adding the global security scheme or header name in-line in the OpenAPI provider block:

Usage:

````
provider "swaggercodegen" {
  apikey_auth = "apiKeyValue"
  x_request_id = "some value for the header"
}
````

##### Environment variables

You can provide the auth credentials and headers via environment variables representing the security definition in the swagger file. 

- In the case of where security definition name is ```apikey_auth```, the corresponding environment
variable name would be ```APIKEY_AUTH```.

````
provider "swaggercodegen" {}
````

Usage:

````
$ export APIKEY_AUTH="apiKeyValue"
$ terraform plan
````

- In the case where the header name is ```x_request_id``` as defined in the extension ```x-terraform-header```
value and the corresponding environment variable name would be ```X_REQUEST_ID```.

Note: if the extension ```x-terraform-header``` is not present, the name of the header will be translated in a terraform 
compliant name (snake_case pattern) and the environment variable name will match that name in upper case.

````
provider "swaggercodegen" {}
````

Usage:

````
$ export X_REQUEST_ID="some value for the header"
$ terraform plan
````

##### Shared OpenAPI Plugin Configuration file

The OpenAPI plugin configuration file may contain schema configuration
for the provider. Read more about how to configure the OpenAPI provider
schema in the [Schema Configuration Object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md#schema-configuration-object)
documentation.


## Examples

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