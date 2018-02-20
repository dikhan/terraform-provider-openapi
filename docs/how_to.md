# How to

The following describe some guidelines in terms of how to define the swagger file to be able to run this terraform
provider. These guidelines not only aim to encourage service providers to follow good practises when defining and 
exposing APIs but also and more importantly serve as a reference on how the different sections of a swagger file
are interpreted and translated into terraform idioms.

## Best Practises

- Resource names should be plural as per [Google API Resource name Guidelines](https://cloud.google.com/apis/design/resource_names).
This means that paths should try and keep their naming plural to refer to the collections. For instance, `/api/users` as opposed to
`/api/user`.
- Property names should be lower case and separated by underscore (e,g: my_property)
- Swagger Tags should be used to group resources by name (several version can be under the same tag)

Refer to [What's supported](#what's-supported?) section to learn more about specific best practises/requirements on
the different OpenAPI Swagger spec fields in the root document. 

## Versioning

API Terraform provider supports resource path versioning, which means that terraform will treat each resource version as
it was a different resource. Refer to the [FAQ](https://github.com/dikhan/terraform-provider-api/blob/master/docs/faq.md#versioning) to get more info about how versioning is handled

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
- **Required:** True
- **Description:**  The host (name or ip) serving the API. This MUST be the host only and does not include the scheme nor sub-paths. 
It MAY include a port. 

The terraform provider uses the host value to configure the internal http/s client used for the CRUD operations.

```yml
host: "api.server.com"
```

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
If both are present, default value is set to https

```yml
schemes:
    - http
    - https
```

#### <a name="globalSecuritySchemes">Global Security Schemes</a>

- **Field Name:** schemes
- **Type:** [string]
- **Required:** No
- **Description:** Applies the specified security schemes, corresponding to a security scheme defined in [securityDefinitions](#swaggerSecurityDefinitions)),
globally to all API operations unless overridden on the operation level.

Global security can be overridden in individual operations to use a different authentication type or no authentication at all:

```yml
security:
  - apiKeyAuth: []
```

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

The Open API terraform provider only considers as a terraform resource, API resource paths that have both the Create (POST)
and Read (GET) operations available in the form of:
```
POST /api/resource
GET /api/resource/{id}
```
Update (PUT) and Delete (DELETE) are optional operations.

The following can be used as a reference to help understand the expected structure.

If a given resource is missing any of the aforementioned required operations, the resource will not be available
as a terraform resource.

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

When the terraform provider is reading the different paths, it will only consider those that match the following criteria:

- In order for an endpoint to be considered as a terraform resource, it must expose a `POST /{resourceName}` and 
`GET,PUT,DELETE /{resourceName}/{id}` operations as shown in the example above. Paths can also be versioned, refer
to [versioning](#versioning) to learn more about it.

- The schema object definition must be described on the root level [definitions](#swaggerDefinitions) section and must 
not be embedded within the API definition. This is enforced to keep the swagger file well structured and to encourage
object re-usability across the CRUD operations. Operations such as POST/GET/PUT are expected to have a 'schema' property
with a link to the actual definition (e,g: `$ref: "#/definitions/resource`)

#### <a name="swaggerDefinitions">Definitions</a>

- **Field Name:** definitions
- **Type:** [Definitions Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#definitionsObject)
- **Required:** Yes
- **Description:** An object to hold data types produced and consumed by operations.

The API Terraform provider uses the object definition used to Create (POST) a resource as the object definition for the all the
CRUD operations. This means that, it is expected that the rest of the operations Read (GET), Update (PUT) and Delete (DELETE)
 will use the same payload and therefore they will all share the same object definition.

##### <a name="supportedTypes">Requirements</a>

The following properties are mandatory when defining the object schema:

- **id**: Object schemas must contain a property called Id which will be used internally to uniquely identify the resource. 

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
readOnly | boolean |  The field will not be considered when updating the resource
default | primitive (int, bool, string) | Default value that will be applied to the property if value is not provided by the user (this attribute can not coexist with readOnly)
x-terraform-immutable | boolean |  The field will be used to create a brand new resource; however it can not be updated. Attempts to update this value will result into terraform aborting the update.
x-terraform-force-new | boolean |  If the value of this property is updated; terraform will delete the previously created resource and create a new one with this value
x-terraform-sensitive | boolean |  If this meta attribute is present in an object definition property, it will be considered sensitive as far as terraform is concerned, meaning that its value will not be disclosed in the TF state file
x-terraform-id | boolean | If this meta attribute is present in an object definition property, the value will be used as the resource identifier when performing the read, update and delete API operations. The value will also be stored in the ID field of the local state file.

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

## What is not supported yet?

- Response definitions: [Responses Definitions Object](https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#responsesDefinitionsObject)
- Oauth2 authentication 
