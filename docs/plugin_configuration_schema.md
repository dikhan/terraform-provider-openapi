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
plugin_version | `string` | Defines the plugin version. If this value is specified, the openapi plugin version executed must match this value; otherwise an error will be thrown at runtime.
insecure_skip_verify | `string` | Defines whether a certificate verification should be performed when retrieving ```swagger-url``` from the server. This is **not recommended** for regular use and should only be set when the server hosting the swagger file is known and trusted but does not have a cert signed by the usually trusted CAs.
schema_configuration | [][Schema Configuration Object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md#schema-configuration-object) |  | Schema Configuration Object

##### Schema Configuration Object

Describes the schema configuration for the service provider:

Field Name | Type | Description
---|:---:|---
schema_property_name | `string` | Defines the name of the provider's schema property. For more info refer to [OpenAPI Provider Configuration](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#configuration)
cmd | `[]string` | Defines the command to execute (using exec form: ```["executable","param1","param2"]```) before the value is assigned to the schema property. This command can be used for example to refresh non static tokens before the value is assigned. Note, there must be at least one value in the array for the cmd to be executed.
cmd_timeout | `int` | Defines the max timeout, in seconds, for the command to execute. If the timeout is not specified the default value is 10s.
default_value | `string` | Defines the default value for the property. If ```schema_property_external_configuration``` is defined, it takes preference over this value.
schema_property_external_configuration | [Schema Property External Configuration Object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/plugin_configuration_schema.md#schema-property-external-configuration) | Schema Property External Configuration Object

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
