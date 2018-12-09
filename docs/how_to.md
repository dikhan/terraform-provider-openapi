# How to

The following describe some guidelines in terms of how to define the swagger file to be able to run this terraform
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

- Paths should be versioned as described in the [versioning](#versioning) document following ‘/v{number}/resource’ pattern 
(e,g: ‘/v1/resource’). A version upgrade (e,g: v1 -> v2) will be needed when the interface of the resource changes, hence 
the new version is non backwards compatible. See that only the 'Major' version is considered in the path, this is recommended 
as the service provider will have less paths to maintain overall. if there are minor/patches applied to the backend that 
should not affect the way consumer interacts with the APIs whatsoever and the namespace should remain as is.

- POST operation should have a body payload referencing a schema object (see example below) defined at the root level 
[definitions](#swaggerDefinitions) section. Payload schema should not be defined inside the path’s configuration. This 
is so the same definition can be shared across different operations for a given version (e,g: $ref: "#/definitions/resource) 
and consistency in terms of data model for a given resource version is maintained throughout all the operations. 
This helps keeping the swagger file well structured and encourages object definition re-usability. 
Different end point versions should their own payload definitions as the example below, path ```/v1/resource``` has a corresponding
```resourceV1``` definition object:

````
  /v1/resource:
    post:
      - in: "body"
        name: "body"
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

Refer to [readOnly](#attributeDetails) attributes to learn more about how to define an object that has computed properties 
(value auto-generated by the API).

- The schema object definition must be described on the root level [definitions](#swaggerDefinitions) section and must 
not be embedded within the API definition. This is enforced to keep the swagger file well structured and to encourage
object re-usability across the CRUD operations. Operations such as POST/GET/PUT are expected to have a 'schema' property
with a link to the actual definition (e,g: `$ref: "#/definitions/resource`)

- The schema object must have a property that uniquely identifies the resource instance. This can be done by either
having a computed property (readOnly) called ```id``` or by adding the ```x-terraform-id``` extension to one of the
existing properties. Read 

##### Extensions

The following extensions can be used in path operations. Read the according extension section for more information

Extension Name | Type | Description
---|:---:|---
[x-terraform-exclude-resource](#xTerraformExcludeResource) | bool | Only available in resource root's POST operation. Defines whether a given terraform compliant resource should be exposed to the OpenAPI Terraform provider or ignored.
[x-terraform-resource-timeout](#xTerraformResourceTimeout) | string | Only available in operation level. Defines the timeout for a given operation. This value overrides the default timeout operation value which is 10 minutes.
[x-terraform-header](#xTerraformHeader) | string | Only available in operation level parameters at the moment. Defines that he given header should be passed as part of the request.
[x-terraform-resource-poll-enabled](#xTerraformResourcePollEnabled) | bool | Only supported in operation responses (e,g: 202). Defines that if the API responds with the given HTTP Status code (e,g: 202), the polling mechanism will be enabled. This allows the OpenAPI Terraform provider to perform read calls to the remote API and check the resource state. The polling mechanism finalises if the remote resource state arrives at completion, failure state or times-out (60s)
[x-terraform-resource-name](#xTerraformResourceName) | string | Only available in resource root's POST operation. Defines the name that will be used for the resource in the Terraform configuration. If the extension is not preset, default value will be the name of the resource in the path. For instance, a path such as /v1/users will translate into a terraform resource name users_v1
[x-terraform-resource-host](#xTerraformResourceHost) | string | Only supported in resource root's POST operation. Defines the host that should be used when managing this specific resource. The value of this extension effectively overrides the global host configuration, making the OpenAPI Terraform provider client make thje API calls against the host specified in this extension value instead of the global host configuration. The protocols (HTTP/HTTPS) and base path (if anything other than "/") used when performing the API calls will still come from the global configuration.
[x-terraform-resource-regions-%s](#xTerraformResourceRegions) | string | Only supported in the root level. Defines the regions supported by a given resource identified by the %s variable. This extension only works if the ```x-terraform-resource-host``` extension contains a value that is parametrized and identifies the matching ```x-terraform-resource-regions-%s``` extension. The values of this extension must be comma separated strings.


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

This extension allows service providers to override the default timeout value for CRUD operations with a different value.

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
    post:
      x-terraform-resource-name: "cdn"
````

In the example above, the resource POST operation contains the extension ``x-terraform-resource-name`` with value ``cdn``.
This value will be the name used in the terraform configuration``cdn``.

````
resource "swaggercodegen_cdn" "my_cdn" {...} # ==> 'cdn' name is used as specified by the `x-terraform-resource-name` extension
````

The preferred name only applies to the name itself, if the resource is versioned like the example below
using version path ``/v1/cdns``, the appropriate postfix including the version will be attached automatically to the resource name.

````
paths:
  /v1/cdns:
    post:
      x-terraform-resource-name: "cdn"
````

The corresponding terraform configuration in this case will be (note the ``_v1`` after the resource name):

````
resource "swaggercodegen_cdn_v1" "my_cdn" {...} # ==> 'cdn' name is used instead of 'cdns'
````

If the ``x-terraform-resource-name`` extension is not present in the resource root POST operation, the default resource
name will be picked from the resource root POST path. In the above example ``/v1/cdns`` would translate into ``cdns_v1``
resource name.

*Note: This extension is only interpreted and handled in resource root POST operations (e,g: /v1/resource) in the
above example*


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

###### <a name="xTerraformResourceRegions">Multi-region resources</a>

Additionally, if the resource is using multi region domains, meaning there's one sub-domain for each region where the resource
can be created into (similar to how aws resources are created per region), this can be configured as follows:

````
swagger: "2.0"
host: "some.domain.com"
x-terraform-resource-regions-cdn: "dub1,sea1"
paths:
  /v1/cdns:
    post:
      x-terraform-resource-host: cdn.${cdn}.api.otherdomain.com
````

If the ``x-terraform-resource-host`` extension has a value parametrised in the form where the following pattern ```${identifier}```
 is found (identifier being any string with no whitspaces - spaces,tabs, line breaks, etc) AND there is a matching
 extension 'x-terraform-resource-regions-**identifier**' defined in the root level that refers to the same identifier
 then the resource will be considered multi region.
For instance, in the above example, the ```x-terraform-resource-host``` value is parametrised as the ```${identifier}``` pattern
is found, and the identifier in this case is ```cdn```. Moreover, there is a matching ```x-terraform-resource-regions-cdn```
extension containing a list of regions where this resource can be created in.

The regions found in the ```x-terraform-resource-regions-cdn``` will be used as follows:

- The OpenAPI Terraform provider will expose one resource per region enlisted in the extension. In the case above, the
following resources will become available in the Terraform configuration (the provider name chosen here is 'swaggercodegen'):

````
resource "swaggercodegen_cdn_v1_dub1" "my_cdn" {
  label = "label"
  ips = ["127.0.0.1"]
  hostnames = ["origin.com"]
}

resource "swaggercodegen_cdn_v1_sea1" "my_cdn" {
  label = "label"
  ips = ["127.0.0.1"]
  hostnames = ["origin.com"]
}
````

As shown above, the resources that are multi-region will have extra information in their name that identifies the region
where tha resource should be managed.

- The OpenAPI Terraform provider client will make the API call against the specific resource region when the resource
is configured with multi-region support.

- As far as the resource configuration is concerned, the swagger configuration remains the same for that specific resource
(parameters, operations, polling support, etc) and the same configuration will be applicable to all the regions that resource
supports.

*Note: This extension is only supported at the root level and can be used exclusively along with the 'x-terraform-resource-host'
extension*

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
[object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#object-definitions) | schema.TypeMap | map value
[array](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#array-definitions) | schema.TypeList | list of values of the same type. The list item types can be primitives (string, integer, number or bool) or complex data structures (objects)

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
  array_of_objects_example = [
    {
      protocol = "http"
    },
    {
      protocol = "tcp"
    }
  ]
  ...
````

**Note**: The items support both nested object definitions (in which case the type **must** be object) and ref to other schema
definitions as described in the [Object definitions](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#object-definitions) 
section.

###### Object definitions

Object types can be defined in two fashions:

####### Nested properties

Properties can have their schema definition in place or nested; and they must be of type 'object'.

````
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    ...
    properties:
      ...
      object_nested_scheme_property:
        type: object # nested properties required type equal object to be considered as object
        properties:
          name:
            type: string
````

####### Ref schema definition

A property that has a $ref attribute is considered automatically an object so defining the type 'object' is optional (although
it's recommended).

````
definitions:
  ContentDeliveryNetworkV1:
    type: "object"
    ...
    properties:
      ...
      object_property:
        #type: object - type is optional for properties of object type that use $ref
        $ref: "#/definitions/ObjectProperty"
  ObjectProperty:
    type: object
    required:
    - message
    properties:
      message:
        type: string
````

Additionally, properties can be flagged as required as follows:

```
    required:
      - mandatoryProperty
```
The provider will configure these properties as required accordingly. Any other property not enlisted in the required field
will be considered optional.

##### <a name="attributeDetails">Attribute details</a>

The following is a list of attributes that can be added to each property to define its behaviour:

Attribute Name | Type | Description
---|:---:|---
readOnly | boolean |  A property with this attribute enabled will be considered a computed property. Hence, it will not be expected from the consumer of the API when posting the resource. However; it will be expected that the API will return tthe property with the computed value in the response payload.
default | primitive (int, bool, string) | Default value that will be applied to the property if value is not provided by the user (this attribute can not coexist with readOnly)
x-terraform-immutable | boolean |  The field will be used to create a brand new resource; however it can not be updated. Attempts to update this value will result into terraform aborting the update.
x-terraform-force-new | boolean |  If the value of this property is updated; terraform will delete the previously created resource and create a new one with this value
x-terraform-sensitive | boolean |  If this meta attribute is present in an object definition property, it will be considered sensitive as far as terraform is concerned, meaning that its value will not be disclosed in the TF state file
x-terraform-id | boolean | If this meta attribute is present in an object definition property, the value will be used as the resource identifier when performing the read, update and delete API operations. The value will also be stored in the ID field of the local state file.
x-terraform-field-name | string | This enables service providers to override the schema definition property name with a different one which will be the property name used in the terraform configuration file. This is mostly used to expose the internal property to a more user friendly name. If the extension is not present and the property name is not terraform compliant, an automatic conversion will be performed by the OpenAPI Terraform provider to make the name compliant (following Terraform's field name convention to be snake_case) 
x-terraform-field-status | boolean | If this meta attribute is present in an object definition property, the value will be used as the status identifier when executing the polling mechanism on eligible async operations such as POST/PUT/DELETE.

##### <a name="definitionExample">Full Example</a>


```yml
definitions:
  resource:
    type: object
    required:
      - mandatoryField
    properties:
      id:
        type: string
        readOnly: true
      
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
x-terraform-authentication-scheme-bearer | boolean |  A security definition with this attribute enabled will enable the Bearer auth scheme. This means that the provider will automatically add to the value of the header/query the Bearer scheme. Note when using this extension the 'name' param will be ignored as this will automatically use the Bearer specification names behind the scenes, that being "Authorization" for header type and "access_token" for the query type.

###### ApiKey Bearer Header Auth

The Bearer scheme for header authentication can be used as follows:

```yml
securityDefinitions:
  apikey_auth:
    type: "apiKey"
    in: "header"
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

###### ApiKey Bearer Query Auth

The Bearer scheme for header authentication can be used as follows:

```yml
securityDefinitions:
  apikey_auth:
    type: "apiKey"
    in: "query"
    x-terraform-authentication-scheme-bearer: true
```

In the example above, the 'name' property does not need to be specified
as using the 'x-terraform-authentication-scheme-bearer' extension the provider internally will take care of
honoring the Bearer specification attaching the right query name to the API request ("access_token")
and the value provided by the user.

Note that the TF property name inside the provider's configuration is exactly the same as the one configured in the swagger
file.

##### <a name="swaggerSecurityDefinitionsRequirements">Requirements</a>

- Terraform requires field names to be lower case and follow the snake_case pattern (my_sec_definition). Thus, security definitions 
 must follow this naming convention.

## What is not supported yet?

- Response definitions: [Responses Definitions Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responsesDefinitionsObject)
- Oauth2 authentication 
