# FAQ

This document aims to provide more insight about the general functionality of this terraform provider.
 
## <a name="howToIntegrate">I am a service provider, does my API need to follow any specification to be able to integrate with this tool?</a> 

Short answer, yes. 

This terraform provider relies on the service provider to follow certain API standards in terms of how the end points 
should be defined. These end points should be defined in a [swagger file](https://swagger.io/specification/) 
that complies with the OpenAPI Specification (OAS) and contains the definition of all the resources supported by the service. 

Please note that swagger is currently a bit behind the latest version of [OpenAPI 3.0](https://swagger.io/specification/#securitySchemeObject). 
Hence, for more information about currently supported features refer to 
[Swagger RESTful API Documentation Specification](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md) 

Additionally, to achieve some consistency across multiple service providers in the way the APIs are structured, it is expected 
the APIs to follow [Google APIs Design guidelines](https://cloud.google.com/apis/design/).

## <a name="versioning">I am service provider and need to upgrade my APIs...How will this provider handle new versions?</a>

The version topic among software engineers is rather conflicting and often involves endless discussions that most of 
the times finish with a non deterministic conclusion. Not having an official guideline that expresses the best-practise 
approach to follow, more specifically **path versioning VS content-type negotiation**, makes it even more difficult to
stick to one or the other. Therefore, you need to make a personal call and in this case the decision has leaned towards
path versioning.

Why?

- Makes the endpoints immutable which means that new versions require a complete new namespace. This helps, as far as the 
API terraform provider is concerned, to handle different versions on the same resource and makes it more explicit from
the users point of view to decide what version to use on the terraform resource type level (see the example below).
- Swagger files are easier to configure using path versioning than content-type versioning
- Each version is associated with its own backend function rather than clugging support for different content types within
the same resource function.
 
This provider expects the service providers to follow [Google APIs Design guidelines](https://cloud.google.com/apis/design/)
so refer to the guideline for any questions related to 'how' the APIs should be structured.

This terraform provider is able to read the resources exposed and the versions they belong too. So, if a service
provider is exposing the following end point with two different versions their corresponding tf configuration would look
like:

- Let's say the service provider initially had support for version 1 and the sagger file looked as follows:

```
paths:
  /v1/cdns:
    post:
      tags:
      - "cdn"
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV1"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
    
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
      - ips
      - hostnames
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
      ips:
        type: "array"
        items:
          type: "string"
      hostnames:
        type: "array"
        items:
          type: "string"
```

The corresponding .tf resource definition would look like:

```
resource "sp_cdns_v1" "my_cdn_v1" {
  label = "label"
  ips = ["127.0.0.1"]
  hostnames = ["origin.com"]
}
```

- After a while, new functionality needs to be supported and the API has to change dramatically resulting into the new API
non being backwards compatible. Solution? Create a new version namespace with its own model object. For the sake of the 
example and to keep it small, V1 version is not shown below.

```
paths:
  .... (/v1/cdns path)
  
  /v2/cdns:
    post:
      tags:
      - "cdn"
      summary: "Create cdn"
      operationId: "ContentDeliveryNetworkCreateV2"
      parameters:
      - in: "body"
        name: "body"
        description: "Created CDN"
        required: true
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV2"
    
  ....
    
definitions:

  ... (ContentDeliveryNetworkV1 definition)

  ContentDeliveryNetworkV2:
    type: "object"
    required:
      - proxyDns
    properties:
      id:
        type: "string"
        readOnly: true
      proxyDns:
        type: "string"
        
  ...
  
```

And the corresponding .tf resource definition would look like:

```
resource "sp_cdns_v2" "my_cdn_v2"{
    proxy_dns = ""
}    
```

## <a name="optionalComputedProperties">Why optional properties with default attributes are translated to the Terraform resource schema as Optional = true and Default = (the default value)?</a>

This enables terraform to know about the default value at plan time. More info [here](https://github.com/hashicorp/terraform/issues/21278)

## <a name="xTerraformOptionalComputed">Why the need for the ‘x-terraform-computed’ extension?</a>

- Without this extension the OpenAPI terraform provider will not be able to identify whether the property is just optional or optional-computed.

### What use cases does the ‘x-terraform-computed’ extension cover?

Some property values that default to computed values may not be known at plan time such as:

- Interrelated properties: Property values can be set based on some other property value, their value depend on other property value. For instance, CPU and Memory could be two properties that the user can populated, however if the user decides to only populate CPU the API will return the preferred Memory based on the CPU units provided.
- Autogenerated computed values: Property values that may be autogenerated at runtime. For instance, a common example might be a property like an event object that contains type and name properties which are required but create_at would be optional-computed. Pseudo schema example below:

````
event:
  type (required)
  name (required)
  create_at (optional)
````
  
The create_at could be populated by the user providing a future date where the event should be created, or if value is not provided the API will compute the now() date automatically.

- Properties where default value is not a primitive: This will also fall under the umbrella of optional-computed since currently there's no way to represent default values in openapi that are not primitive.


## <a name="xTerraformOptionalComputed">What if I want a given property to have a default value that is not the one specified in the default attribute?</a>

- This could be achieved in the future with support for a new extension - something like: x-terraform-default. This will 
enable service providers to set the behaviour of default attributes to preferred ones in terraform if desired. 

Note: This is not supported at the moment.

## <a name="multipleEnvironments">I am service provider and currently support multiple environments. How will this provider handle that?</a>

To be decided...

