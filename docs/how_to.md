# How to

The following describe some guidelines in terms of how to define the OpenAPI file to be able to run this terraform
provider. These guidelines not only aim to encourage service providers to follow good practises when defining and 
exposing APIs but also and more importantly serve as a reference on how the different sections of a swagger file
are interpreted and translated into terraform idioms.


## Best Practises

- Resource names should be plural as per [Google API Resource name Guidelines](https://cloud.google.com/apis/design/resource_names).
This means that paths should try and keep their naming plural to refer to the collections. For instance, `/api/users` as opposed to
`/api/user`. More granular access to the resource should be permitted exposing /resource/{id} endpoints.

- Swagger Tags should be used to group resources by name (several version can be under the same tag)

Refer to [What's supported](#what's-supported?) section to learn more about specific best practises/requirements on
the different OpenAPI Swagger spec fields in the root document. 

## Versioning

API Terraform provider supports resource path versioning, which means that terraform will treat each resource version as
it was a different resource. Refer to the [FAQ](https://github.com/dikhan/terraform-provider-api/blob/master/docs/faq.md#versioning) 
to get more info about how versioning is handled.

## What's supported?

#### <a name="swaggerVersion">Swagger Version</a>

- **Field Name:** swagger
- **Type:** String
- **Required:** True
- **Description:**  Specifies the Swagger Specification version being used. 

This property is used by the provider to validate that the api is compatible with the swagger version supported. 
Version `"2.0"` is the only version supported at the moment.

```yml
swagger: '2.0'
```

#### <a name="swaggerHost">Host</a>

- **Field Name:** host
- **Type:** String
- **Required:** False
- **Description:**  The host (name or ip) serving the API. This MUST be the host only and does not include the scheme nor sub-paths. 
It may include the port number if different from the scheme’s default port (80 for HTTP and 443 for HTTPS).

**Note:** FQDNs using non standard HTTP ports (e,g: api.server.com:8080) is currently not supported, it is assumed that
FQDNs will use standard ports. However, if localhost is present in the host value it may contain non standard ports 
(e,g: localhost:8080). Keep in mind which schemes (http/https) are in use though as the host field can only specify
specific port for either or not both.

The terraform provider uses the host value to configure the internal http/s client used for the CRUD operations.

```yml
host: "api.server.com"
```

If host is not specified, it is assumed to be the same host where the API documentation is being served. This is handy
when multiple environments are supported (e,g: dev, stage and prod) and all of them share the same swagger file. If the
host field was present, then one swagger file will be required per environment supported pointing at the specific env
domain where the file is hosted. However, not having the host field specified will simplify that setup allowing service
provider to just have one swagger file to maintain. At runtime, API calls will be make against the FQDN where the swagger
file is hosted.

#### <a name="swaggerBasePath">Base Path</a>

- **Field Name:** host
- **Type:** String
- **Required:** No
- **Description:**  The base path on which the API is served, which is relative to the [`host`](#swaggerHost). 
If it is not included, the API is served directly under the `host`. The value MUST start with a leading slash (`/`).

```yml
basePath: "/"
```

#### <a name="swaggerSchemes">Schemes</a>

- **Field Name:** schemes
- **Type:** [string]
- **Required:** Yes
- **Description:**  The transfer protocol of the API. Values MUST be from the list: `"http"`, `"https"`. 
If both are present, the OpenAPI Terraform provider will always use HTTPs as default scheme for API calls.

```yml
schemes:
    - http
    - https
```

#### <a name="globalSecuritySchemes">Global Security Schemes</a>

- **Field Name:** security
- **Type:** [string]
- **Required:** No
- **Description:** Applies the specified security schemes, corresponding to a security scheme defined in [securityDefinitions](#swaggerSecurityDefinitions)),
globally to all API operations unless overridden on the operation level.

Global security can be overridden in individual operations to use a different authentication type or no authentication at all:

```yml
security:
  - api_key_auth: []
```

If multiple authentication is required, that can be achieved as follows:

```yml
security:
  - api_key_auth: []
    api_key_auth2: []
```

The above means that **both** authentication schemes, ```api_key_auth``` and ```api_key_auth2``` will be used when calling 
the APIs.

Alternatively, the example below means that **either** of the authentication schemes defined will be used. By default, the
OpenAPI Terraform provider picks the first one in the list by order of appearance, in this case ```api_key_auth``` will be
used as the global authentication mechanism.

```yml
security:
  - api_key_auth: []
  - api_key_auth2: []
```

More information about multiple API keys can be found [here](https://swagger.io/docs/specification/authentication/api-keys/#multiple).

#### <a name="swaggerConsumes">Consumes</a>

- **Field Name:** consumes
- **Type:** [string]
- **Required:** No
- **Description:**  A list of MIME types the APIs can consume. This is global to all APIs but can be overridden on specific API calls. 
Values MUST include application/json

*This value is currently not validated in the terraform provider; the provider assumes that the APIs accept json.*

```yml
consumes:
    - application/json
```

#### <a name="swaggerProduces">Produces</a>

- **Field Name:** produces
- **Type:** [string]
- **Required:** No
- **Description:**  A list of MIME types the APIs can produce. This is global to all APIs but can be overridden on specific API calls. 
Values MUST include application/json

*This value is currently not validated in the terraform provider; the provider assumes that the APIs return json.*

```yml
produces:
    - application/json
```

#### <a name="swaggerPaths">Paths</a>

- **Field Name:** paths
- **Type:** [Path Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#pathsObject)
- **Required:** Yes
- **Description:** The available paths and operations for the API.

The following can be used as a reference to help understand the expected structure.

```yml
paths:
  /resource:
    post:
      ...
      - in: "body"
        name: "body"
        required: true
        schema:
          $ref: "#/definitions/resource"       
      ...
  /resource/{id}:
    get:
      ...
    put:
      ...
    delete:
      ...
```

##### Terraform compliant resource requirements

A resource to be considered terraform compliant must meet the following criteria:

- The resource must have at least a POST and a GET operations defined as shown in the example below. Update (PUT) and 
Delete (DELETE) operations are optional.

```
paths:
  /{resourceName}:
    post:
      ...
  /resource/{id}:    
    get:
      ...
```

If a given resource is missing any of the aforementioned required operations, the resource will not be available
as a terraform resource.

- POST operation may have a request body payload either defined inside the path’s configuration or referencing a schema object (using $ref) 
defined at the root level [definitions](#swaggerDefinitions) section. The $ref can be a link to a local model definition or a definition hosted
externally. The request payload schema may be the same as the response schema, including the expected input properties (required and optional) as well 
as the computed ones (readOnly):

````
  /v1/resource:
    post:
      parameters:
      - in: "body"
        name: "body"
        schema:
          $ref: "#/definitions/resourceV1" 
      responses:
        201:
          schema:
            $ref: "#/definitions/resourceV1" 
definitions:   
  resourceV1:     
    type: object    
    required:       
      - name    
    properties:
      id:         
        type: string        
        readOnly: true
      name:         
        type: string        
````

- If the POST operation does not share the same definition for the request payload and the response payload, in order for the endpoint to be considered terraform compliant:
  - the POST request payload must contain only input properties, that is required and optional properties. 
  - the POST operation responses must contain at least one 'successful' response which can be either a 200, 201 or 202 response. The schema associated with
  the successful response will need to contain all the input properties from the request payload (but configured as readOnly) as well as any other
  computed property (readOnly) auto-generated by the API. 
  - the POST response schema must contain only readOnly properties. If the response schema properties are not explicitly configured as readOnly, the provider will automatically convert them as computed (readOnly).
  - If the POST response schema does not contain any property called 'id', at least one property must contain the [x-terraform-id](#attributeDetails) extension which
  serves as the resource identifier.
  
The following shows an example of a compatible terraform resource that expects in the request body only the input properties (`label` required property and `optional_property` an optioanl property)
 and returns in the response payload both the inputs as well as any other output (computed properties) generated by the API:

````
  /v1/resource:
    post:
      parameters:
      - in: "body"
        name: "body"
        required: true
        schema:
          $ref: "#/definitions/ResourceRequestPayload"
      responses:
        201:
          schema:
            $ref: "#/definitions/ResourceResponsePayload"
  /v1/resource/{resource_id}:
    get:
      parameters:
      - name: "resource_id"
        in: "path"
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/ResourceResponsePayload"
          
definitions:
  ResourceRequestPayload:
    type: "object"
    required:
      - label
    properties:
      label:
        type: "string"
      optional_property:
        type: "string"
  ResourceResponsePayload:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
        readOnly: true
      optional_property:
        type: "string"
        readOnly: true    
````

The resulted resource's terraform schema configuration will contain the combination of the request and response schemas keeping
the corresponding input configurations as is (eg: required and optional properties will still be required and optional in the resulted final schema) as well as the output computed properties (readOnly properties from the response). 
Note, if both the request and response schema properties contain extensions and their values are different, the extension value kept
for the property will be the one in the response.

- If the POST operation does not contain a request body payload, in order for the endpoint to be considered terraform compliant:
  - the POST operation responses must contain at least one 'successful' response which can be either a 200, 201 or 202 response. The schema associated with
   the successful response will be the one used as the resource schema. Note if more than one successful response is present in the
   responses the first one found (in no particular order) will be used.  
  - the POST response schema must contain only readOnly properties.
  - If the POST response schema does not contain any property called 'id', at least one property must contain the [x-terraform-id](#attributeDetails) extension which
  serves as the resource identifier.  
  
The following shows an example of a compatible terraform resource that does not expect any input upon creation but does return
computed data:

````
paths:
  /v1/deployKey:
    post:
      responses:
        201:
          schema:
            $ref: "#/definitions/DeployKeyV1"
  /v1/deployKey/{id}:
    get:
      parameters:
      - name: "id"
        in: "path"
        required: true
        type: "string"
      responses:
        200:
          schema:
            $ref: "#/definitions/DeployKeyV1"
definitions:
  DeployKeyV1: # All the properties are readOnly
    type: "object"
    properties:
      id:
        readOnly: true
        type: string
      deploy_key:
        readOnly: true
        type: string
````

Refer to [readOnly](#attributeDetails) attributes to learn more about how to define an object that has computed properties 
(value auto-generated by the API).

- The resource's POST, GET and PUT (if exposed) operations must have the same response schema configuration. This is required to ensure
the resource state is consistent.

- The resource's PUT operation (if exposed) request payload must be the same as the resources POST's request schema. The OpenAPI provider expects this
 to ensure the update operation enables replacement of the representation of the target resource.

- The resource's PUT operation may return one of the following successful responses:
  - 200 OK with a response payload containing the final state of the resource representation in accordance with the state of the enclosed representation and any other computed property. The response schema must be the same as the GET operation response schema.
  - 202 Accepted for async resources. Refer to [asynchronous resources](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#xTerraformResourcePollEnabled) for more info.
  - 204 No Content with an empty response payload.
  
- The schema object must have a property that uniquely identifies the resource instance. This can be done by either
having a computed property (readOnly) called ```id``` or by adding the [x-terraform-id](#attributeDetails) extension to one of the
existing properties.
  
- The resource schema object definition should be described on the root level [definitions](#swaggerDefinitions) section and should not 
 be embedded within the API definition. This is recommended to keep the OpenAPI document well structured and to encourage
object re-usability across the resource CRUD operations. The resource POST/GET/PUT operations are expected to have a 'schema' property
for both the request and response payloads with a link to the same definition (e,g: `$ref: "#/definitions/resource`). The ref can 
be a link to an external source as described in the [OpenAPI documentation for $ref](https://swagger.io/docs/specification/using-ref/).

- Paths should be versioned as described in the [versioning](#versioning) document following ‘/v{number}/resource’ pattern 
(e,g: ‘/v1/resource’). A version upgrade (e,g: v1 -> v2) will be needed when the interface of the resource changes, hence 
the new version is non backwards compatible. See that only the 'Major' version is considered in the path, this is recommended 
as the service provider will have less paths to maintain overall. if there are minor/patches applied to the backend that 
should not affect the way consumer interacts with the APIs whatsoever and the namespace should remain as is.
  - Different end point versions should their own payload definitions as the example below, path ```/v1/resource``` has a corresponding ```resourceV1``` definition object:  

###### Data source instance

Any resources that are deemed terraform compatible as per the previous section, will also expose a terraform data source 
that internally will be mapped to the GET operation (in the previous example that would be GET ```/resource/{id}```).

This type of data source is named data source instance. The data source name will be formed from the resource name 
plus the ```_instance``` string attach to it.

###### Argument Reference

````
data "openapi_resource_v1_instance" "my_resource_data_source" {
   id = "resourceID"
}
````  

- id: string value of the resource instance id to be fetched

###### Attributes Reference

The data source state will be filled with the corresponding properties defined in the resource model definition, in the 
example above that would be ```resourceV1```. Please note that all the properties from the model will be configured as computed 
in the data source schema and will be available as attributes. 



##### Terraform data source compliant requirements

The OpenAPI provider is able to export data sources from paths that are data source compatible.

An endpoint (path) to be considered terraform data source compliant must meet the following criteria:

- The path must be a root level path (e,g: /v1/cdns) not an instance path (e,g: /api/v1/cdn/{id}). Subresource data source paths are also supported (e,g: /v1/cdns/{id}/firewalls)
- The path must contain a GET operation with a response 200 which contains a schema of type 'array'. The items schema must be of type 'object' and must specify at least one property.
- The items schema object definition must contain a property called ```id``` which will be used internally to uniquely identify the data source. If
the object schema does not have a property called ```id```, then at least one property should have the ```x-terraform-id``` extension 
so the OpenAPI Terraform provider knows which property should be used to unique identifier instead.

The following snipped of code shows a valid terraform compliant data source endpoint ```/v1/cdns```:

````
paths:
  /v1/cdns:
    get:
      summary: "Get all cdns"
      responses:
        200:
          description: "successful operation"
          schema:
            $ref: "#/definitions/ContentDeliveryNetworkCollectionV1"
definitions:
  ContentDeliveryNetworkCollectionV1:
    type: "array"
    items:
      $ref: "#/definitions/ContentDeliveryNetworkV1"
  ContentDeliveryNetworkV1:
    type: "object"
    required:
      - label
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"
      computed_property:
        type: "string"
        readOnly: true
````

The corresponding Terraform data source configuration will be:

````
data "openapi_cdns_v1" "my_data_source" {
  filter {
    name = "label"
    values = ["my_label"]
  }
}
````

Check out the argument and attributes references below to learn more about how the input expected and ouput produced by
the data sources.

*Refer to [x-terraform-resource-name](#xTerraformResourceName) to learn more about how the data source name (```cdns_v1```) type is built.*

###### Argument Reference

filter - (Optional) One or more name/value pairs to filter off of. The keys allowed to filter by will depend on the properties
 exposed in the swagger model definition for the data source path. In the example above the corresponding model definition
 for ```/v1/cdns``` was the ```ContentDeliveryNetworkV1```, which exposed three properties - id, label and computed_property. These
 become automatically available as filter for the data source. 

**NOTE**: Currently, only primitive properties are supported as filters. If the model definition contains properties that are
not primitive (e,g: arrays or objects), these will not be available as filters.
**NOTE**: If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single result only.

###### Attributes Reference

id is set to the ID of the found result. In addition, the properties defined in the swagger model definition of the data
source will be exported as attributes. 

In the previous example, let's pretend the ```GET /v1/cdns``` API returned a list of cdns:

````
[
    {
      "id":"someID",
      "label":"someLabel",
      "computed_property":"computed property"
    },
    {
      "id":"someOtherID",
      "label":"someOtherLabel",
      "computed_property":"computed property 2"
    },
]
````

From the above list, the filter set up was to retrieve the cdn with ```label = my_label```, hence the matching item would
be:

````
    {
      "id":"someID",
      "label":"someLabel",
      "computed_property":"computed property"
    }
````

Considering the above result, the openapi plugin will then go ahead and start setting the data source terraform state with
the properties and values of the matching result.

##### Extensions

The following extensions can be used in path operations. Read the according extension section for more information

Extension Name | Type | Description
---|:---:|---
[x-terraform-exclude-resource](#xTerraformExcludeResource) | bool | Only available in resource root's POST operation. Defines whether a given terraform compliant resource should be exposed to the OpenAPI Terraform provider or ignored.
[x-terraform-resource-timeout](#xTerraformResourceTimeout) | string | Only available in operation level. Defines the timeout for a given operation. This value overrides the default timeout operation value which is 10 minutes.
[x-terraform-header](#xTerraformHeader) | string | Only available in operation level parameters at the moment. Defines that he given header should be passed as part of the request.
[x-terraform-resource-poll-enabled](#xTerraformResourcePollEnabled) | bool | Only supported in operation responses (e,g: 202). Defines that if the API responds with the given HTTP Status code (e,g: 202), the polling mechanism will be enabled. This allows the OpenAPI Terraform provider to perform read calls to the remote API and check the resource state. The polling mechanism finalises if the remote resource state arrives at completion, failure state or times-out (60s)
[x-terraform-resource-name](#xTerraformResourceName) | string | Only supported in resource root level. Defines the name that will be used for the resource in the Terraform configuration. If the extension is not preset, default value will be the name of the resource in the path. For instance, a path such as /v1/users will translate into a terraform resource name users_v1
[x-terraform-resource-host](#xTerraformResourceHost) | string | Only supported in resource root's POST operation. Defines the host that should be used when managing this specific resource. The value of this extension effectively overrides the global host configuration, making the OpenAPI Terraform provider client make thje API calls against the host specified in this extension value instead of the global host configuration. The protocols (HTTP/HTTPS) and base path (if anything other than "/") used when performing the API calls will still come from the global configuration.

###### <a name="xTerraformExcludeResource">x-terraform-exclude-resource</a>
 
Service providers might not want to expose certain resources to Terraform (e,g: admin resources). This can be achieved 
by adding the following swagger extension to the resource root POST operation (in the example below ```/v1/resource:```):

````
paths:
  /v1/resource:
    post:
      ...
      x-terraform-exclude-resource: true
      ...
  /v1/resource/{id}:
    get:
      ...     
````

The resource root POST operation is a mandatory operation for a resource to be terraform compliant; hence if the resource
is deemed Terraform compliant an extra validation is performed to check if the resource is meant to be exposed by checking
this extension. If the extension is not present or has value 'false' then the resource will be exposed as usual.

*Note: This extension is only interpreted and handled in resource root POST operations (e,g: /v1/resource) in the
above example*

###### <a name="xTerraformResourceTimeout">x-terraform-resource-timeout</a>

This extension allows service providers to override the default timeout value for CRUD operations with a different value
for both operations that are synchronous and [asynchronous](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#xTerraformResourcePollEnabled).

The value must comply with the duration type format. A duration string is a sequence of decimal positive numbers (negative numbers are not allowed),
each with optional fraction and a unit suffix, such as "300s", "20.5m", "1.5h" or "2h45m".

Valid time units are "s", "m", "h".

````
paths:
  /v1/resource:
    post:
      ...
      x-terraform-resource-timeout: "15m" # this means the max timeout for the post operation to finish is 15 minutes. This overrides the default timeout per operation which is 10 minutes
      ...
  /v1/resource/{id}:
    get: # will have default value of 10 minutes as the 'x-terraform-resource-timeout' is not present for this operation
      ...
    delete:
      x-terraform-resource-timeout: "20m" # this means the max timeout for the delete operation to finish is 20 minutes. This overrides the default timeout per operation which is 10 minutes
      ...
````

This extension will also enable users to specify a different value from the terraform configuration file. For instance, the
example above will expose the timeouts property in the resource_v1 but only for the create and delete operations enabling the
user to override the default values in the OpenAPI document with different ones:

````
resource "openapi_resource_v1" "my_resource" {
  timeouts {
    create = "10s"
    delete = "5s"
  }
}
````

Hence, overriding the default timeout value set in the swagger document for the ```/v1/resource``` post operation from 15m to 10s
and the default timeout value set in the swagger document for the ```/v1/resource/{id}``` delete operation from 20m to 5s.

*Note: This extension is only supported at the operation level*

###### <a name="xTerraformHeader">x-terraform-header</a>  

Certain operations may specify other type of parameters besides a 'body' type parameter which defines the payload expected 
by the API. One example is 'header' type parameters which are also supported by the openapi terraform provider, meaning that when
a request is performed against an operation that requires headers, these will be sent along the payload. In the following
example, a body payload (defined at #/definitions/resource) along with the header 'X-Request-ID' will be sent when performing
the POST request. 

````
paths:    
/resource:     
  post:       
  ...      
  - in: "body"        
    name: "body"        
    required: true        
    schema:           
      $ref: "#/definitions/resource"
    responses:
      ...
  - in: "header"            
    name: "X-Request-ID" # This header will be send along with the request when making the POST request against the '/resource' API            
    required: true            
    x-terraform-header: x_request_id 
    ...         
  ...  
````

The value of the header will be defined by the end user in the terraform configuration file as follows:

````
provider "swaggercodegen" {
  x_request_id = "request header value for POST /resource"
}
````

The field name ```x_request_id``` is defined by the extension property ```x-terraform-header: x_request_id``` when defining
the header parameter; however, if the extension is not present the terraform provider will fall back to its default behaviour
and will convert the name of the header into a field name that is terraform compliant (Field name may only contain lowercase 
alphanumeric characters & underscores.). Hence, the result in this case will be the same ```x_request_id```. The value of 
the header will be the one specified in the terraform configuration ```request header value for POST /resource```.

*Note: Currently, parameters of type 'header' are only supported on an operation level*

###### <a name="xTerraformResourcePollEnabled">x-terraform-resource-poll-enabled</a>

This extension allows the service provider to enable the polling mechanism in the OpenAPI Terraform provider for asynchronous
operations. In order for this to work, the following must be met:

- The resource response status code must have the 'x-terraform-resource-poll-enabled' present and set to true.
- The resource definition must have a **read-only** field that defines the status of the resource. By default, if a string field
named 'status' is present in the resource schema definition that field will be used to track the different statues of the resource. Alternatively,
a field can be marked to serve as the status field adding the 'x-terraform-field-status'. This field will be used as the status
field even if there is another field named 'status'. This gives service providers flexibility to name their status field the
way they desire. More details about the 'x-terraform-field-status' extension can be found in the [Attribute details](#attributeDetails) section.
- The polling mechanism requires two more extensions to work which define the expected 'status' values for both target and
pending statuses. These are:

  - **x-terraform-resource-poll-completed-statuses**: (type: string) Comma separated values - Defines the statuses on which the resource state will be considered 'completed'
*Note: For DELETE operations, the expected behaviour is that when the resource has been deleted, GET requests to the deleted
resource would return a 404 HTTP response status code back. This means that no payload will be returned in the response,
and hence there won't be any status field to check against to. Therefore, the OpenAPI Terraform provider handle deletes
target statuses in a different way not expecting the service provide to populate this extension. Behind the scenes, the
OpenAPI Terraform provider will handle the polling accordingly until the resource is no longer available at which point
the resource will be considered destroyed. If the extension is present with a value, it wil be ignored in the backend.*
  - **x-terraform-resource-poll-pending-statuses**: (type: string) Comma separated values - Defines the statuses on which the resource state will be considered 'in progress'.
Any other state returned that returned but is not part of this list will be considered as a failure and the polling mechanism
will stop its execution accordingly.

**If the above requirements are not met, the operation will be considered synchronous and no polling will be performed.**

In the example below, the response with HTTP status code 202 has the extension defined with value 'true' meaning
that the OpenAPI Terraform provider will treat this response as asynchronous. Therefore, the provider will perform
continues calls to the resource's instance GET operation and will use the value from the resource 'status' property to
determine the state of the resource:

````
  /v1/lbs:
    post:
      ...
      responses:
        202: # Accepted
          x-terraform-resource-poll-enabled: true # [type (bool)] - this flags the response as trully async. Some resources might be async too but may require manual intervention from operators to complete the creation workflow. This flag will be used by the OpenAPI Service provider to detect whether the polling mechanism should be used or not. The flags below will only be applicable if this one is present with value 'true'
          x-terraform-resource-poll-completed-statuses: "deployed" # [type (string)] - Comma separated values with the states that will considered this resource creation done/completed
          x-terraform-resource-poll-pending-statuses: "deploy_pending, deploy_in_progress" # [type (string)] - Comma separated values with the states that are "allowed" and will continue trying
          schema:
            $ref: "#/definitions/LBV1"
definitions:
  LBV1:
    type: "object"
    required:
      - name
      - backends
    properties:
      ...
      status:
        x-terraform-field-status: true # identifies the field that should be used as status for async operations. This is handy when the field name is not status but some other name the service provider might have chosen and enables the provider to identify the field as the status field that will be used to track progress for the async operations
        description: lb resource status
        type: string
        readOnly: true
        enum:
          - deploy_pending
          - deploy_in_progress
          - deploy_failed
          - deployed
          - delete_pending
          - delete_in_progress
          - delete_failed
          - deleted
````

Alternatively, the status field can also be of 'object' type in which case the nested properties can be defined in place or
the $ref attribute can be used to link to the corresponding status schema definition. The nested properties are considered
computed automatically even if they are not marked as readOnly.

````
definitions:
  LBV1:
    type: "object"
    ...
    properties:
      newStatus:
        $ref: "#/definitions/Status"
        x-terraform-field-status: true # identifies the field that should be used as status for async operations. This is handy when the field name is not status but some other name the service provider might have chosen and enables the provider to identify the field as the status field that will be used to track progress for the async operations
        readOnly: true
      timeToProcess: # time that the resource will take to be processed in seconds
        type: integer
        default: 60 # it will take two minute to process the resource operation (POST/PUT/READ/DELETE)
      simulate_failure: # allows user to set it to true and force an error on the API when the given operation (POST/PUT/READ/DELETE) is being performed
        type: boolean
  Status:
    type: object
    properties:
      message:
        type: string
      status:
        type: string
````

*Note: This extension is only supported at the operation's response level.*


###### <a name="xTerraformResourceName">x-terraform-resource-name</a>

This extension enables service providers to write a preferred resource name for the terraform configuration.

````
paths:
  /cdns:
    x-terraform-resource-name: "cdn"
````

In the example above, the resource  contains the extension ``x-terraform-resource-name`` with value ``cdn``.
This value will be the name used in the terraform configuration``cdn``.

````
resource "swaggercodegen_cdn" "my_cdn" {...} # ==> 'cdn' name is used as specified by the `x-terraform-resource-name` extension
````

The preferred name only applies to the name itself, if the resource is versioned like the example below
using version path ``/v1/cdns``, the appropriate postfix including the version will be attached automatically to the resource name.

````
paths:
  /v1/cdns:
    x-terraform-resource-name: "cdn"
````

The corresponding terraform configuration in this case will be (note the ``_v1`` after the resource name):

````
resource "swaggercodegen_cdn_v1" "my_cdn" {...} # ==> 'cdn' name is used instead of 'cdns'
````

If the ``x-terraform-resource-name`` extension is not present in the resource root level operation or the resource POST level operation*, 
the default resource name will be picked from the resource root path. In the above example ``/v1/cdns`` would translate into ``cdns_v1``
resource name.

*Note: Support for this extension on the resource root POST operation is still currently supported but 
will be deprecated in the future, so users are encouraged to use the extension on the resource root level.


###### <a name="xTerraformResourceHost">x-terraform-resource-host</a>

This extension allows resources to override the global host configuration with a different host. This is handy when
a given swagger file may combine resources provided by different service providers.

````
swagger: "2.0"
host: "some.domain.com"
paths:
  /v1/cdns:
    post:
      x-terraform-resource-host: cdn.api.otherdomain.com
````

The above configuration will make the OpenAPI Terraform provider client make API CRUD requests (POST/GET/PUT/DELETE) to
the overridden host instead, in this case ```cdn.api.otherdomain.com```.

*Note: This extension is only supported at the operation's POST operation level. The other operations available for the
resource such as GET/PUT/DELETE will used the overridden host value too.*

#### <a name="swaggerDefinitions">Definitions</a>

- **Field Name:** definitions
- **Type:** [Definitions Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#definitionsObject)
- **Required:** Yes
- **Description:** An object to hold data types produced and consumed by operations.

The API Terraform provider uses the object definition used to Create (POST) a resource as the object definition for the all the
CRUD operations. This means that, it is expected that the rest of the operations Read (GET), Update (PUT) and Delete (DELETE)
 will use the same payload and therefore they will all share the same object definition.

##### <a name="definitionRequirements">Requirements</a>

- Terraform requires field names to be lower case and follow the snake_case pattern (my_property). Thus, definition object 
fields must follow this naming convention.
  
- Object schemas must contain a property called ```id``` which will be used internally to uniquely identify the resource. If
the object schema does not have a property called ```id```, then at least one property should have the ```x-terraform-id``` extension 
so the OpenAPI Terraform provider knows which property should be used to unique identifier instead. This property must
have the readOnly attribute present with value equal to true.

```yml
      id:
        type: "string"
        readOnly: true
```
*Refer to [Attribute details](#attributeDetails) for more info about readOnly properties*


##### <a name="supportedTypes">Supported types</a>

The following property types will be translated into their corresponding terraform types.

Swagger Type | TF Type | Description
---|:---:|---
string | Type: schema.TypeString | string value
integer | schema.TypeInt | int value
number | schema.TypeFloat | float value
boolean | schema.TypeBool | boolean value
[object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#object-definitions) | schema.TypeList with MaxItems 1 and Elem *Resource | The list will contain only one element. The element will be the object with its corresponding properties which can be primitives as well as objects or lists.
[array](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#array-definitions) | schema.TypeList | list of values of the same type. The list item types can be primitives (string, integer, number or bool) or complex data structures (objects)

###### Object definitions

Object types can be defined using `type: "object"` and the internal Terraform's schema representation will be a TypeList with MaxItems 1 and Elem *Resource
and an HCL representation of block type. This is due to the current version of Terraform SDK, at the time of writing 2.0, not supporting helper/schema.TypeMap with Elem *helper/schema.Resource and
follows Hashi Terraform maintainers' recommendations:

- [Issue 616](https://github.com/hashicorp/terraform-plugin-sdk/issues/616): Upgrading OpenAPI Terraform provider to Terraform SDK 2.0: TypeMap with Elem*Resource not supported
- [Issue 22511](https://github.com/hashicorp/terraform/issues/22511): Objects that contain properties with different types (e,g: string, integer, etc) and configurations (e,g: some of them being computed)
- [Issue 21217](https://github.com/hashicorp/terraform/issues/21217): Objects that contain nested objects

The OpenAPI Terraform provider supports among others the following scenarios when handling properties of type object:

- Scenario 0: Simple objects where all the object properties are the same type and are required or optional.
- Scenario 1: Complex objects that contain properties with different types (e,g: string, integer, etc) and configurations (e,g: some of them being computed)
- Scenario 2: Complex objects that contain nested objects

The following example shows how to define an OpenAPI definition `ContentDeliveryNetworkV1` that contains an object type property
and what will be the internal Terraform schema representation as well as the user facing tf configuration using the block type.

````
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    properties:
      ...
      object_property: 
        type: "object"
        properties:
          name:
            type: "string"
          nested_object_property:
            type: "object"
            properties:
              account:
                type: string
      ...
````

The above will be translated into the following Terraform schema internally:

````
&schema.Resource{
        # This will be the Terraform schema of the resource using the ContentDeliveryNetworkV1 model definition
		Schema: map[string]*schema.Schema {
		    ...
		    "object_property": *schema.Schema {
                Type:TypeList 
                Optional:true 
                Required:false 
                ...
                Elem: &{
                          Schema:map[name:0xc0005ee700 nested_object_property:0xc0005ee800] 
                          SchemaVersion:0 
                          MigrateState:<nil> 
                          StateUpgraders:[] 
                          Create:<nil> 
                          Read:<nil> 
                          Update:<nil> 
                          Delete:<nil> 
                          Exists:<nil> 
                          CustomizeDiff:<nil> 
                          Importer:<nil> 
                          DeprecationMessage: 
                          Timeouts:<nil>
                } 
                MaxItems:1 
                MinItems:0 
                ...		    
		    }
		},
		...
	}
````

This would translate into the following terraform configuration:

````
resource "openapi_cdn_v1" "my_cdn" {
  ....
  object_property {
    name = ""
    nested_object_property {
      account = ""
    }
  }
  ....
}
````

The above OpenAPI spec could have also been defined using the $ref directive which would have resulted into the same result
as far as the Terraform schema is concerned:

````
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    properties:
      ...
      object_property:
        $ref: "#/definitions/ObjectProperty"
  ObjectProperty:
    type: object
    properties:
      name:
        type: string
      nested_object_property:
        type: "object"
        properties:
          account:
            type: string
````

Remember that due to the internal schema representation of object properties being of `helper/schema.TypeList with Elem *helper/schema.Resource and MaxItems 1`
if the object property needs to be referenced from other places in the terraform configuration the list syntax needs to be used indexing
on the zero element. Example: `openapi_cdn_v1.my_cdn.object_property[0].name`. Similarly to reference the nested_object_property which
is an object property you would do `openapi_cdn_v1.my_cdn.object_property[0].nested_object_property[0].account`.

###### Array definitions

Arrays can be constructed containing simple values like primitive types (string, integer, number or bool) or complex
types defined by the object definition. In any case, the swagger property 'items' must be populated when describing
an array property.

- Arrays of primitive values (string, integer, number or bool primitives):

````
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    ...
    properties:
      arrayOfOStringsExample: # This is an example of an array of strings
        type: "array"
        items:
          type: "string"
````

The above OpenAPI configuration would translate into the following Terraform configuration:

````

resource "swaggercodegen_cdn_v1" "my_cdn" {
  ...
  array_of_strings_example = ["somevalue", "some-other-value"]
  ...
````

- Arrays of complex values (objects):

The example below shows how the property named 'arrayOfObjectsExample' is configured with type 'array' and the 'items'
are of type object, meaning that the array holds objects inside as described in the object 'properties' section.

````
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    ...
    properties:
      arrayOfObjectsExample: # This is an example of an array of objects
        type: "array"
        items:
          type: "object"
          properties:
            protocol:
              type: string
````

The above OpenAPI configuration would translate into the following Terraform configuration:

````
resource "swaggercodegen_cdn_v1" "my_cdn" {
  ...
  array_of_objects_example {
    protocol = "http"
  }

  array_of_objects_example {
    protocol = "tcp"
  }
  ...
````

**Note**: The items support both nested object definitions (in which case the type **must** be object) and ref to other schema
definitions as described in the [Object definitions](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#object-definitions)
section.

##### <a name="attributeDetails">Attribute details</a>

The following is a list of attributes that can be added to each property to define its behaviour:

Attribute Name | Type | Description
---|:---:|---
readOnly | boolean |  A property with this attribute enabled will be considered a computed property. readOnly properties are included in responses but not in requests. Hence, it will not be expected from the consumer of the API when posting the resource. However; it will be expected that the API will return tthe property with the computed value in the response payload.
description | string | A description for property. 
default | primitive (int, bool, string) | Documents what will be the default value generated by the API for the given property
x-terraform-immutable | boolean |  The field will be used to create a brand new resource; however it can not be updated. Attempts to update this value will result into terraform aborting the update. This applies also to properties of type object and also list of objects. If an object property contains this attribute, any update to its child properties will result  terraform aborting the update too. Also, if an object property is does not contain this flag, but any of its child properties, the same principle applies and updates to the values of those properties will not be allowed.
x-terraform-force-new | boolean |  If the value of this property is updated; terraform will delete the previously created resource and create a new one with this value
x-terraform-sensitive | boolean | If this meta attribute is present in a definition property, it will be considered sensitive as far as terraform is concerned, meaning that the attribute's value does not get displayed in logs or regular output. It should be used for passwords or other secret fields.
x-terraform-id | boolean | If this meta attribute is present in an object definition property, the value will be used as the resource identifier when performing the read, update and delete API operations. The value will also be stored in the ID field of the local state file.
x-terraform-field-name | string | This enables service providers to override the schema definition property name with a different one which will be the property name used in the terraform configuration file. This is mostly used to expose the internal property to a more user friendly name. If the extension is not present and the property name is not terraform compliant (following snake_case), an automatic conversion will be performed by the OpenAPI Terraform provider to make the name compliant (following Terraform's field name convention to be snake_case) 
x-terraform-field-status | boolean | If this meta attribute is present in a definition property, the value will be used as the status identifier when executing the polling mechanism on eligible async operations such as POST/PUT/DELETE.
[x-terraform-ignore-order](#xTerraformIgnoreOrder) | boolean | If this meta attribute is present in a definition property of type list, when the plugin is updating the state for the property it will inspect the items of the list received from remote and compare with the local values and if the lists are the same but unordered the state will keep the users input. Please go to the `x-terraform-ignore-order` section to learn more about the different behaviours supported. 
x-terraform-write-only | boolean | If this meta attribute is present in a definition property, when the plugin is reading or updating the state for the property it will always take the local state's value as the the value of the property. Any changes in the remote state for such properties will be ignored. 

###### <a name="xTerraformIgnoreOrder">x-terraform-ignore-order</a>

This extension enables the service providers to setup the 'ignore order' behaviour for a property of type list defined in
the object definition. For instance, the API may be returning the array items in lexical order but that behaviour might
not be the desired one for the terraform plugin since it would cause DIFFs for users that provided the values of the array
property in a different order. Hence, ensuring as much as possible that the order for the elements in the input list 
provided by the user is maintained.

Given the following terraform snippet where the members values are in certain desired order and assuming that the members 
property is of type 'list' AND has the `x-terraform-ignore-order` extension set to true in the OpenAPI document for the `group_v1` 
resource definition:

````
resource "openapi_group_v1" "my_iam_group_v1" {
  members = ["user1", "user2", "user3"]
}
````

The following behaviour is applied depending on the different scenarios when processing the response received by the API
and saving the state of the property.

- Use case 0: If the remote value for the property `members` contained the same items in the same order (eg: `{"members":["user1", "user2", "user3"]}`) as the tf input then the state saved for the property would match the input values. That is: ``members = ["user1", "user2", "user3"]``
- Use case 1: If the remote value for the property `members` contained the same items as the tf input BUT the order of the elements is different (eg: `{"members":["user3", "user2", "user1"]}`) then state saved for the property would match the input values. That is: ``members = ["user1", "user2", "user3"]``
- Use case 2: If the remote value for the property `members` contained the same items as the tf input in different order PLUS new ones (eg: `{"members":["user2", "user1", "user3", "user4"]}`) then state saved for the property would match the input values and also add to the end of the list the new elements received from the API. That is: ``members = ["user1", "user2", "user3", "user4"]``
- Use case 3: If the remote value for the property `members` contained a shorter list than items in the tf input (eg: `{"members":["user3", "user1"}`) then state saved for the property would contain only the matching elements between the input and remote. That is: ``members = ["user1", "user3"]``
- Use case 4: If the remote value for the property `members` contained the same list size as the items in the tf input but some elements inside where updated (eg: `{"members":["user1", "user5", "user9"]}`) then state saved for the property would contain the matching elements  between the input and output and also keep the remote values. That is: ``members = ["user1", "user5", "user9"]``

##### <a name="propertyUseCasesSupport">Property use cases</a>

Properties can be defined with different behaviours and constraints. As far as properties for definitions go, the following 
use cases are supported as per OpenAPI spec 2.0:

- Required properties

Properties that are required are the ones enlisted in the ```required``` section of the definition. These properties will 
be considered ```Required=true``` in the Terraform resource property schema. Hence, input will be expected from the user. See how the
**mandatory_property** property is defined in the example below.

- Optional properties: Properties that are not required, are automatically considered optional. The Terraform resource
property schema will have ```Optional=true``` for these properties. There are few use cases to consider under this section:

  - Purely optional properties: This means that if the value is not provided by the client, the property will not be pushed as part of the API request, and
will not be stored in the terraform state file either. Note: the API is not expected to return a computed value for the property
if input is not given, that's the purpose of ```Optional computed``` properties. See example below **optional_property**

  -  Optional computed: These properties combine both worlds. If the client does not provide a value for the property, 
  the API is expected to generate one by default. Otherwise, the API should take the input given by the client. Due to the nature
  of this behaviour, the Terraform resource schema property will be set as ```Computed=true```. There are 
  two use cases under the optional-computed scenario:
     - Optional computed with default: The default value is known at plan time. In this case, the service provider should document 
     what the default value is specifying the default attribute with the known value. See example below **optional_computed_with_default**
     - Optional computed: The default value is NOT known at plan time: In this case, the value is probably autogenerated by 
     the API and therefore the value is not known at plan time. Hence, the user should attach to the property the 
     *‘x-terraform-computed’* attribute so the OpenAPI Terraform provider will understand the behaviour expected. More info
     about this extension can be found in the [FAQ](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/faq.md#why-the-need-for-the-x-terraform-computed-extension) section.
     See example below **optional_computed**

**NOTE**: 
  - Object properties containing optional computed child properties will also need to include the extension ```x-terraform-computed```. Otherwise
  the Terraform schema for the object will not be marked as computed and any non expected value change in the child properties will result into diffs.
  - Optional properties that are of type `array` that contain a default value are not supported at the moment and the provider
  will ignore the Default value when creating the schema for the property. This is due to Terraform not supporting at the moment
  [default values to be set in the schema's Default field for TypeList properties](https://github.com/hashicorp/terraform-plugin-sdk/blob/v2.5.0/helper/schema/schema.go#L763).  

- Computed properties: These properties must contain the ```readOnly``` attribute set. These properties are included 
in responses but not in requests, and the value is automatically assigned by the API. See example below **computed**
     - Computed with default: These properties must have also the ```default``` attribute present. This will represent that the 
     computed value is known at plan time and enables the service provider to document the behaviour of the property and be 
     transparent with the client. See example below **computed_with_default**

More info about what led to the above designs here:
- [How to configure optional-computed properties?](https://github.com/hashicorp/terraform/issues/21278)
- [OpenAPI 2.0 Read-Only properties explained](https://swagger.io/docs/specification/data-models/data-types#readonly-writeonly)
- [OpenAPI 2.0 Default attribute](https://swagger.io/docs/specification/describing-parameters#default)

##### <a name="definitionExample">Full Example</a>


```yml
definitions:
  resource:
    type: object
    required:
      - mandatory_property
    properties:
      id: # the value of this computed property is not known at plan time (e,g: uuid, etc)
        type: string
        description: "some description for the property..."
        readOnly: true
 
      mandatory_property: # this property is required
        type: string  
              
      # Primitives  
      string_prop:
        type: string          
      
      integer_prop:
        type: integer
      
      number_prop:
        type: number
      
      boolean_prop:
        type: boolean        
      
      string_array_prop:
        type: "array"
        items:
          type: "string"

      optional_property: # this property is optional as far as input from user is concerned. If the API will compute a value when it's not provided by the user, use instead 'optional_computed' or 'optional_computed_with_default' property definitions.
        type: "string"

      computed: # computed property that the value is NOT know at plan time, and the API will compute one and return it in the response
        type: "string"
        readOnly: true 
  
      computed_with_default: # computed property that the default value is known at plan time
        type: "string"
        readOnly: true
        default: "computed value known at plan time" # this computed value happens to be known before hand, the default attribute is just for documentation purposes
      
      optional_computed: # optional property that the default value is NOT known at plan time
        type: "string"
        x-terraform-computed: true
      
      optional_computed_with_default: # the value happens to be known at plan time, so the service provider decides to document what the default value will be if the client does not provide a value
        type: "string"
        default: "some computed value known at plan time"

      # Properties with attributes that define behaviour

      computed_prop:
        type: boolean
        readOnly: true
                
      immutable_prop:
        type: string
        x-terraform-immutable: true
        
      force_new_prop:
        type: number
        x-terraform-force-new: true

      sensitive_prop:
        type: string
        x-terraform-sensitive: true        
        
      someNonUserFriendlyPropertyName:  # If this property did not have the 'x-terraform-field-name' extension, the property name will be automatically converted by the OpenAPI Terraform provider into a name that is Terraform field name compliant. The result will be:  some_non_user_friendly_propertyName
        type: string
        x-terraform-field-name: property_name_more_user_friendly
```


#### <a name="swaggerSecurityDefinitions">Security Definitions</a>

- **Field Name:** securityDefinitions
- **Type:** [Security Definitions](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#securityDefinitionsObject)
- **Required:** This configuration is up to the user
- **Description:**  Security scheme definitions that can be used across the specification. After you have defined the 
security schemes in securityDefinitions, you can apply them to the whole API or individual operations by adding the 
security section on the root level (global security schemes) or operation level, respectively.

The API terraform provider supports apiKey type authentication in the header as well as a query parameter. The
location can be specified in the 'in' parameter of the security definition.

If an API has a security policy attached to it (as shown below), the API provider will use the corresponding policy
when performing the HTTP request to the API.

```yml
paths:
  /resource:
    post:
      ...
      security:
        - apikey_auth: []
      ...
```

```yml
securityDefinitions:
  apikey_auth:
    type: "apiKey"
    name: "Authorization"
    in: "header"
```

The provider automatically identifies header/query based auth policies and exposes them as part of the provider
TF configuration so the actual token can be injected into the HTTP calls. The following is an example on how a user would
be able to configure the provider with the auth header key. Internally, the provider will use this value for every API that has
 the 'apikey_auth' attach to it. Moreover, the name of the header/query parameter will be the one specified in the
 'name' property of the security definition, in the above example 'Authorization'.

Below is the corresponding TF configuration, for a provider that has a header based authentication in the swagger file
(as the example above):
```
provider "sp" {
  apikey_auth = "apiKeyValue"
}
```

##### Security Definitions extensions

The following terraform specific extensions are supported to complement the lack of support
for authentication schemes in the OpenAPI 2.0 specification. To use them, just add the extension
to the security definition as the example below.

Attribute Name | Type | Description
---|:---:|---
[x-terraform-authentication-scheme-bearer](#xTerraformAuthenticationSchemeBearer) | boolean |  A security definition with this attribute enabled will enable the Bearer auth scheme. This means that the provider will automatically use the header/query names specified in the Auth Bearer specification. Note when using this extension the 'name' param will be ignored as this will automatically use the Bearer specification names behind the scenes, that being "Authorization" for header type and "access_token" for the query type.
[x-terraform-refresh-token-url](#xTerraformAuthenticationRefreshToken) | string |  The URL that will be used to post the refresh token (provided in the plugin config input - using the sed def name) and will return an access token that then will be used in every API call made by the plugin. This is useful specially for resource that take a long time to complete and the token may expire before they finish.

###### <a name="xTerraformAuthenticationRefreshToken">x-terraform-refresh-token-url</a>

This extension is used as an alternative authenticator mechanism from the default authenticator used in the provider (where 
the same security definition token provided in the plugin configuration is always posted to the API). This extension enables the 
user to provide a refresh token URL where instead, the user will provide a refresh token in the configuration and the URL
provided in the ```x-terraform-refresh-token-url``` extension value will be the one where the refresh token will be posted. Then
the response from the API should contain a header called ```Authorization``` containing the access token tobe used in the
subsequent resource API calls.

The following example shows how this authorization works behind the scenes:

- The following security definition will be exposed in the plugin config as a string input ```apikey_auth``` where the user
wold provide the value for the refresh token.

  - Swagger config
```yml
securityDefinitions:
  apikey_auth:
    type: "apiKey"
    in: "header"
    x-terraform-refresh-token-url: https://api.iam.com/auth/token
```

  - Provider plugin config with the value of the refresh token
```
provider "sp" {
  apikey_auth = "refresh token value"
}
```

- Behind the scenes then the provider will go ahead and perform the following before making any API request to the resource
endpoints:
  - Perform a POST request to ```http://api.iam.com/auth/token``` passing in a header with name `Authorization` and as value
  the refresh token value `refresh token value`
  - The response is expected to have a status code 200 or 204, and the response header must contain an `Authorization` header
  containing the session token generated. This session token will be the one used for any API request made to the resource
  endpoints. Note: the whole contained in the header value will be used as the session token, hence if the value contains
  the Bearer scheme that will also get send to the API endpoints.

###### <a name="xTerraformAuthenticationSchemeBearer">x-terraform-authentication-scheme-bearer</a>

The 'x-terraform-authentication-scheme-bearer' extension can be applied to
a security definition of type 'apiKey' in both header as well as query locations
 (as described in the 'in' parameter). The extension enables the bearer scheme authentication
 following the the [OAuth 2.0 Authorization Framework: Bearer Token Usage RFC](https://tools.ietf.org/html/rfc6750#page-5)

The following sections describe more details about header and query options:

- ApiKey Bearer Header Auth

The Bearer scheme for header authentication can be used as follows:

```yml
securityDefinitions:
  apikey_auth:
    type: "apiKey"
    in: "header"
    # name: "something" the name paramter will be ignored when using the 'x-terraform-authentication-scheme-bearer' extension, "Authorization" name will be use as default value
    x-terraform-authentication-scheme-bearer: true
```

In the example above, the 'name' property does not need to be specified
as using the 'x-terraform-authentication-scheme-bearer' extension the provider internally will take care of
honoring the Bearer specification attaching the right header name to the API request ("Authorization")
and adding the 'Bearer' keyword before the JWT token for the header value. In this
case the user is expected to add just the value of the JWT token. For compatibility reasons
the implementation also handles the case where the user has provided as the
auth value the Bearer plus the token in which case no extra Bearer will be added to the header
value avoiding duplications.

- ApiKey Bearer Query Auth

The 'x-terraform-authentication-scheme-bearer' extension can be applied to
 an 'apiKey' type authentication of type query (as described in the 'in'
 parameter). The extension enables the bearer scheme authentication
 following the the [OAuth 2.0 Authorization Framework: Bearer Token Usage RFC](https://tools.ietf.org/html/rfc6750#page-5)

The Bearer scheme for header authentication can be used as follows:

```yml
securityDefinitions:
  apikey_auth:
    type: "apiKey"
    in: "query"
    # name: "something" the name paramter will be ignored when using the 'x-terraform-authentication-scheme-bearer' extension, "access_token" name will be use as default value
    x-terraform-authentication-scheme-bearer: true
```

In the example above, the 'name' property does not need to be specified
as using the 'x-terraform-authentication-scheme-bearer' extension the provider internally will take care of
honoring the Bearer specification attaching the right query name to the API request ("access_token")
and the value provided by the user.

Note that the TF property name inside the provider's configuration is exactly the same as the one configured in the swagger
file.

#### <a name="subresource-configuration">Sub-resource configuration</a>

Refer to the [sub-resource documentation](https://github.com/dikhan/terraform-provider-openapi/tree/master/docs/how_to_subresources.md) to learn more about this.

#### <a name="multiRegionConfiguration">Multi-region configuration</a>

This section describes how to configure the swagger file for a service that operates multi-region, meaning there's an API for each region.

The example below shows how terraform configuration will look like if the swagger file contains multiregion support:

Assuming the following swagger configuration:

````
swagger: 2.0
...
x-terraform-provider-multiregion-fqdn: "service.api.${region}.hostname.com"
x-terraform-provider-regions: "rst, dub"
....
````

The above will be translated into the following terraform configuration:

````
provider "provider" {}

provider "provider" {
  alias = "dub"
  region = "dub"
}


## this resource will be managed in the default provider region, in this case rst (as it's the first element in the 'x-terraform-provider-regions' comma separated value) and API calls will be made against service.api.rst.hostname.com
resource "provider_resource" "my_resource_rst" {
  name = "resource in rst"
}

## this resource will be managed with the provider with alias dub, hence the region will be dub and API calls will be made against service.api.dub.hostname.com
resource "provider_resource" "my_resource_dub" {
  provider = "provider.dub"
  name = "resource in dub"
}

````

In order to support multi-region configuration, the following extensions must be set with the right values:

#### Multi-region Extensions

The following extensions can be used in the root level. Read the according extension section for more information

Extension Name | Type | Description
---|:---:|---
[x-terraform-provider-multiregion-fqdn](#xTerraformProviderMultiregionFQDN) | string | Defines the host that should be used when managing the resources exposed. The value of this extension effectively overrides the global host configuration, making the OpenAPI Terraform provider client make the API calls against the host specified in this extension value instead of the global host configuration. The protocols (HTTP/HTTPS) and base path (if anything other than "/") used when performing the API calls will still come from the global configuration. The value must be parameterised following the expected format (regex: (S+)(${(S+)})(S+)) where the ${region} section identifies the spot that will be replaced by the region value. E,g: service.api.${region}.hostname.com.
[x-terraform-provider-regions](#xTerraformProviderRegions) | string | Defines the regions the service has APIs exposed and will be translated into the terraform provider 'region' property. The value must be a comma separated list of strings. The default region value set in the provider will be the first element in the comma separated string. The value set, either the default or the one provider by the user, will be used to build the right FQDN based on the 'x-terraform-provider-multiregion-fqdn' value. In the example above, if the region value was 'uswest1', the API calls will be made against the following hostL: service.api.uswest1.hostname.com 

##### <a name="xTerraformProviderMultiregionFQDN">x-terraform-provider-multiregion-fqdn</a>

This extension defines the FQDN to be used by Terraform when managing the service resources. The value must be parameterised
following the pattern (S+)(${(S+)})(S+) where the ${} section identifies the location that will be replaced by the region value. 

````
x-terraform-provider-multiregion-fqdn: "service.api.${region}.hostname.com"
````

This extension must be present with the correct parameterised value in order for multi-region to be enabled.

##### <a name="xTerraformProviderRegions">x-terraform-provider-regions</a>

This extension defines the different regions supported by the service provider. The values will be used in the 'x-terraform-provider-multiregion-fqdn'
value to build the final FQDN with the right region. The default value set in the terraform provider will be the first
element in the command separated list, in the example below that will be 'rst':

````
x-terraform-provider-regions: "rst, dub"
````

Note: This extension will be ignored if the ``x-terraform-provider-multiregion-fqdn`` is not present.

### <a name="swaggerSecurityDefinitionsRequirements">Requirements</a>

- Terraform requires field names to be lower case and follow the snake_case pattern (my_sec_definition). Thus, security definitions 
 must follow this naming convention.

## Path collisions

_If one or more resources have the same path, then one of them will be accessible in the provider and the others will 
not, and which one is available will be determined at run time and may change from one invocation to the 
next, such that different resource types may be created, updated, and destroyed with different terraform plan, apply and 
destory invocations._

## Resource naming collisions

When resource names collide, the provider is unable to determine which resource the name refers to in tf files, so it 
will not provide access to either resource (unless there is a path collision, as documented above).  

Here are some scenarios that will result in naming collisions such that the resources will not available in the 
provider: 
- Two or more resources with the same `x-terraform-resource-name` when both are versioned or neither is versioned.  
  - Example 1: A swagger document defines one resource with a path of `/abc` and a `x-terraform-resource-name` of 
  `something` and another resource with a path of `/xyz` and a `x-terraform-resource-name` of `something`.  The  
  resource names for both of them would be `something`.
  - Example 2: A swagger document defines one resource with a path of `/v1/abc` and a `x-terraform-resource-name` of 
  `something` and another resource with a path of `/v1/xyz` and a `x-terraform-resource-name` of `something`.  The  
  resource names for both of them would be `something_v1`.
- Versioned resources with non-versioned resources having version-like patterns in the paths.  For example, if a swagger 
document defines a path for one resource of `/v1/abc` and  a path for another resource of `/abc_v1`, then the  resource 
names for both of them would be `abc_v1`.
- Resources with `x-terraform-resource-name` name values matching the path of another resource without a 
`x-terraform-resource-name`.
  - Example 1: One resource has a path of `/abc` while another has a `x-terraform-resource-name` value of `abc`.  The 
  resource name for both will be `abc`.
  - Example 2: One resource has a path of `/v1/abc` while another has a path of `/abc` and a `x-terraform-resource-name`
  value of `abc_v1`.  The resource name for both will be `abc_v1`.
  
Note that none these scenarios above involve duplicate paths, which is addressed above in the "Path collisions" section. 

## What is not supported yet?

- Response definitions: [Responses Definitions Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responsesDefinitionsObject)
- Oauth2 authentication 

