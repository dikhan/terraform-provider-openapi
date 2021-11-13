# OpenAPI plugin configuration schema

The OpenAPI terraform plugin can be configured externally by defining a configuration file following one of the specifications
below. The configuration files have to be written using yaml 2.0 specification.

## Specification

### Format
The files describing the OpenAPI plugin configuration are represented as YAML objects and conform to the YAML standards.

### File Structure
The OpenAPI plugin configuration is made of a single file.

By convention, the Swagger specification file is named terraform-provider-openapi.yaml.

### File Location

The OpenAPI plugin configuration file has to be placed in the terraform's plugins folder ```~/.terraform.d/plugins``` along
with the terraform-provider-openapi binary.

```
$ pwd
/Users/dikhan/.terraform.d/plugins
$ ls -la
total 44112
drwxr-xr-x  6 dikhan  staff       192  4 Jul 17:12 .
drwxr-xr-x  5 dikhan  staff       160  4 Jul 13:19 ..
lrwxr-xr-x  1 dikhan  staff        63  4 Jul 17:12 terraform-provider-goa -> /Users/dikhan/.terraform.d/plugins/terraform-provider-openapi
-rwxr-xr-x  1 dikhan  staff  22257828  4 Jul 17:12 terraform-provider-openapi
-rw-r--r--  1 dikhan  staff       127  4 Jul 17:12 terraform-provider-openapi.yaml
```

Alternately, the location of the file can be specified by setting the
following environment variable OTF_VAR_%s_PLUGIN_CONFIGURATION_FILE
where '%s' should be replaced with your provider's name.

````
$ export OTF_VAR_myprovider_PLUGIN_CONFIGURATION_FILE="/Users/user/myprovider_config.yaml"
````

### Data Types
Primitive data types in the OpenAPI plugin configuration specification are based on the types supported by the YAML-Schema 2.0.

### Schema V1

#### PluginConfigSchema Object

This is the root document object for the plugin configuration specification.

##### Fixed Fields

Field Name | Type | Description
---|:---:|---
version | `string` | **Required.** Specifies the OpenAPI plugin configuration spec version being used. The value MUST be `'1'`.
services | [Services Object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md#services-object) | Specifies the service configurations

##### Services Object

Holds the configuration for individual services

Field Name | Type | Description
---|:---:|---
{service_name} | [Service Item Object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md#service-item-object) | Defines the configuration available for this service with name {service_name}. The {service_name} must match the name use for the terraform provider, being terraform-provider-{service_name}.
 
##### Service Item Object

Describes the configurations available on a single service.

Field Name | Type | Description
---|:---:|---
swagger-url | `string` | **Required.** Defines the location where the swagger document is hosted. The value must be either a valid formatted URL or a path to a swagger file stored in the disk
insecure_skip_verify | `string` | Defines whether a certificate verification should be performed when retrieving ```swagger-url``` from the server. This is **not recommended** for regular use and should only be set when the server hosting the swagger file is known and trusted but does not have a cert signed by the usually trusted CAs.
schema_configuration | [][Schema Configuration Object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md#schema-configuration-object) |  | Schema Configuration Object
telemetry | [Telemetry Object](#telemetry-object) | Telemetry configuration

##### Schema Configuration Object

Describes the schema configuration for the service provider:

Field Name | Type | Description
---|:---:|---
schema_property_name | `string` | Defines the name of the provider's schema property. For more info refer to [OpenAPI Provider Configuration](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#configuration)
cmd | `[]string` | Defines the command to execute (using exec form: ```["executable","param1","param2"]```) before the value is assigned to the schema property. This command can be used for example to refresh non static tokens before the value is assigned. Note, there must be at least one value in the array for the cmd to be executed. If the command fails to execute, the plugin will log the error and continue its execution.
cmd_timeout | `int` | Defines the max timeout, in seconds, for the command to execute. If the timeout is not specified the default value is 10s.
default_value | `string` | Defines the default value for the property. If ```schema_property_external_configuration``` is defined, it takes preference over this value.
schema_property_external_configuration | [Schema Property External Configuration Object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md#schema-property-external-configuration) | Schema Property External Configuration Object. If there is an error when retriving the info from the external source, the plugin will log the error and continue its execution and will set the default value as empty ultimately delegating the responsibility to the API to complain about any missing required property. 

##### Schema Property External Configuration Object

Describes the schema configuration for the service provider:

Field Name | Type | Description
---|:---:|---
file | `string` | Defines the location where the swagger document is hosted. The value must be either a valid formatted URL or a path to a swagger file stored on disk. Paths starting with `~` will be expanded to user's home directory
key_name | `string` | Defines the key name of the property to look for in the `file`. The file must be JSON formatted if this property is populated. The value must be formatted using the [JsonPath syntax](https://github.com/oliveagle/jsonpath)
content_type | `string` | Defines the type of content in the ```file```. Supported values are: raw, json

The [JSONPath online evaluator](http://jsonpath.com/) can be used to play around with the syntax
and validate right paths.

#### Example

````
version: '1'
services:
    monitor: # Basic example of service that has basic configuration
      swagger-url: http://monitor-api.com/swagger.json
      insecure_skip_verify: true
    cdn: # More advanced example of a service that has schema configuration for schema property 'apikey_auth', including a default value and also schema external configuration that will set as default value the 'raw' contents of the file located at '/Users/dikhanr/.terraform.d/plugins/swaggercodegen'
      swagger-url: /Users/user/go/src/github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen/api/resources/swagger.yaml
      schema_configuration:
      - schema_property_name: "apikey_auth"
        cmd: ["date"]
        cmd_timeout: 10
        default_value: "apiKeyValue"
        schema_property_external_configuration:
          content_type: raw
          file: /Users/dikhanr/.terraform.d/plugins/swaggercodegen
    vm: # More advanced example of a service that has schema configuration for schema property 'some_property', including a default value and also schema external configuration that will set as default value the 'raw' contents of the file located at '/Users/dikhanr/.terraform.d/plugins/swaggercodegen'
      swagger-url: http://vm-api.com/swagger.json
      schema_configuration:
      - schema_property_name: "some_property"
        default_value: "someDefaultValue"
        schema_property_external_configuration:
          content_type: json # This defines the content type of the 'file'
          key_name: $.token # This is the key to look for in the json file provided in the 'file' field, in this case as seen in the example below the default value will be 'superSecret'
          file: /Users/dikhanr/my_service/vm.json # The content of the file could looke like: {"token":"superSecret", "createdAt":"Mar.01,2000 15:45:17"}
    goa: 
      swagger-url: https://some-domain-where-swagger-is-served.com/swagger.yaml
````

##### Telemetry Object

Describes the telemetry providers configurations.

Field Name | Type | Description
---|:---:|---
graphite | [Graphite Object](#graphite-object) | Graphite Telemetry configuration
http_endpoint | [HTTP Endpoint Object](#http-endpoint-object) | HTTP Endpoint Telemetry configuration

###### Graphite Object

Describes the configuration for Graphite telemetry.

Field Name | Type | Description
---|:---:|---
host | `string` | **Required.** Graphite host to ship the metrics to
port | `integer` | **Required.** Graphite port to connect to
prefix | `string` | Some prefix to append to the metrics pushed to Graphite. If populated, metrics pushed to Graphite will be of the following form: `statsd.<prefix>.terraform....`. If the value is not provided, the metrics will not contain the prefix.

The following metrics will be shipped to the corresponding configured Graphite host upon plugin execution

  - Terraform OpenAPI version used by the user: `statsd.<prefix>.terraform.openapi_plugin_version.*.total_runs:1|c|#openapi_plugin_version:0_25_0` where the tagged `openapi_plugin_version` value would contain the corresponding OpenAPI terraform plugin version used by the user (e,g: v0_25_0, etc)
  - Service used by the user: `statsd.<prefix>.terraform.provider:1|c|#provider_name:myProviderName,resource_name:cdn_v1,terraform_operation:create` where the tagged `provider_name`, `resource_name` and `terraform_operation` values would contain the corresponding plugin name (service provider) used by the user (e,g: if the plugin name was terraform-provider-cdn the provider name in the metric would be 'cdn'), resource name being provisioned and operation performed (eg: create, read, update, delete)

###### HTTP Endpoint Object

Describes the configuration for HTTP endpoint telemetry.

Field Name | Type | Description
---|:---:|---
url | `string` | **Required.** URL endpoint to where the metrics will be sent to (eg: https://my-app.com/v1/metrics).
prefix | `string` | Some prefix to append to the metrics pushed to the http endpoint. If populated, metrics pushed to the endpoint will be of the following form: `<prefix>.terraform....`. If the value is not provided, the metrics will not contain the prefix. 
provider_schema_properties | `[]string` | Defines what specific provider configuration properties and their values will be injected into metric API request headers. This is useful in cases where you need the specified provider configuration's properties as part of for instance the metric tags. Values must match a real property name in provider schema configuration.

The following metrics will be shipped to the corresponding configured URL endpoint upon plugin execution:

  - Terraform OpenAPI version used by the user: `<prefix>.terraform.openapi_plugin_version.total_runs`. This metric is posted
  any time the plugin is executed.
  - Service used by the user: `<prefix>.terraform.provider`. This metric is posted any time the plugin is provisioning a resource
  via any of the CRUD operations. This metric will be submitted upon resource provisioning as well as data source.

The above will result into separate POST HTTP requests to the corresponding configured URL passing in a JSON payload 
containing the `metric_type` with value 'IncCounter' and the `metric_name` being one of the above values. The 'IncCounter' 
value describes an increase of 1 in the corresponding counter metric, the consumer (eg: API) then will decide how to handle this 
information. The request will also contain a `User-Agent` header identifying the OpenAPI Terraform provider as the client.

- Example of HTTP request sent to the HTTP endpoint increasing the `<prefix>.terraform.openapi_plugin_version.total_runs` counter:
````
curl -X POST https://my-app.com/v1/metrics -d '{"metric_type": "IncCounter", "metric_name":"<prefix>.terraform.openapi_plugin_version.total_runs", "tags": ["openapi_plugin_version:0_25_0"]}' -H "Content-Type: application/json" -H "User-Agent: OpenAPI Terraform Provider/v0.26.0-b8364420eb450a34ff02e4c7832ad52165cd05b4 (darwin/amd64)"
````  

Note the specific OpenAPI terraform plugin version used is passed in the `tags` property (e,g: 0_25_0, etc).

- Example of HTTP request sent to the HTTP endpoint increasing the `<prefix>.<prefix>.terraform.provider` counter:

````
curl -X POST https://my-app.com/v1/metrics -d '{"metric_type": "IncCounter", "metric_name":"<prefix>.terraform.provider", "tags": ["provider_name:cdn", "resource_name":"cdn_v1", "terraform_operation":"create"]}' -H "Content-Type: application/json" -H "User-Agent: OpenAPI Terraform Provider/v0.26.0-b8364420eb450a34ff02e4c7832ad52165cd05b4 (darwin/amd64)"
````

The `terraform_operation` value will correspond the specific operation executed by Terraform. That is: create, update, read or delete. 

Note the specific plugin name (service provider) used is passed in the `tags` property (e,g: if the plugin name was terraform-provider-cdn the provider name in the metric would be 'cdn')


- Example of HTTP request submitting the `terraform.providers.total_runs` metric where the `provider_schema_properties` property is populated with one of the properties exposed by the provider configuration:

````
provider openapi {
  billing_id = "some_id"
}

resource "openapi.cdn_v1" "my_cdn" {
...
}
````

````
curl -X POST https://my-app.com/v1/metrics -d '{"metric_type": "IncCounter", "metric_name":"<prefix>.terraform.provider", "tags": ["provider_name:openapi", "resource_name":"cdn_v1", "terraform_operation":"create"]}' -H "billing_id: some_id" -H "Content-Type: application/json" -H "User-Agent: OpenAPI Terraform Provider/v0.26.0-b8364420eb450a34ff02e4c7832ad52165cd05b4 (darwin/amd64)"
````

Note the provider configuration property and its value is attached to the header (following the OpenAPI plugin behaviour when appending to
the API requests the provider configuration properties) so the API will then be able to use this value for whatever it needs to.