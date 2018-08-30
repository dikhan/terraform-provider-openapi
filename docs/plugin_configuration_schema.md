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

#### Example

````
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
````
