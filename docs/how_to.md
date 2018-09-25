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
[x-terraform-resource-name](#xTerraformResourceName) | string | Only available in resource root's POST operation. Defines the name that will be used for the resource in the Terraform configuration. If the extension is not preset, default value will be the name of the resource in the path. For instance, a path such as /v1/users will translate into a terraform resource name users_v1

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
[string] | schema.TypeList (schema.TypeString) | list of string values
integer | schema.TypeInt | int value
number | schema.TypeFloat | float value
boolean | schema.TypeBool | boolean value

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

The API terraform provider supports apiKey type authentication in the header as well as a query parameter.

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
Note that the TF property name inside the provider's configuration is exactly the same as the one configured in the swagger
file.

##### <a name="swaggerSecurityDefinitionsRequirements">Requirements</a>

- Terraform requires field names to be lower case and follow the snake_case pattern (my_sec_definition). Thus, security definitions 
 must follow this naming convention.

## What is not supported yet?

- Response definitions: [Responses Definitions Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responsesDefinitionsObject)
- Oauth2 authentication 
