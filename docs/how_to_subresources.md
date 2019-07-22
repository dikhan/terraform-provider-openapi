# How to describe sub-resources

The following describe the guidelines on how to define a swagger file containing sub-resources; endpoints that 
live under other resources.

## Best Practises

The OpenAPI Terraform plugin aims to provide an easy way to expose API endpoints via Terraform. However, it also tries
to encourage good API design following good Restful practises. Given that, it is recommended that sub-resources endpoints
are formed in such a way that the sub-resource can always be referenced from the parent using the URL.

### How can sub-resources be described in the OpenAPI document?

The description of a sub-resource in the OpenAPI document is no different than any other resource and must meet the same [Terraform compliance requirements](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#terraform-compliant-resource-requirements).

If the parent paths are not described in the OpenAPI doc or they are not [Terraform compliant](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#terraform-compliant-resource-requirements)
then you will not be able to manage the sub-resource in the provider.

Nevertheless, the sub-resource endpoint path must contain the path parameters referring to the parents where they live under and
the path parameters must be named the same.

For instance, in the following example the ```/v1/firewalls/``` sub-resource lives under the ```/v1/cdns/{cdn_id}``` resource and
both the parent and the sub-resource instance paths contain the same path parameter name that refers to the parent ```{cdn_id}```.

````
paths:

  /v1/cdns:
    post:
      parameters:
      - in: "body"
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkV1"
      responses:
        201:
      ...
    ...
  /v1/cdns/{cdn_id}:
    get:
      ...
    put:
      ...
    delete:
      ....

  /v1/cdns/{cdn_id}/v1/firewalls:
    post:
      parameters:
      - name: "parent_id"
        in: "path"
        type: "string"
      - in: "body"
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallV1"    
      ....
  /v1/cdns/{parent_id}/v1/firewalls/{id}:
    get:
      ...
      
definitions:
  ContentDeliveryNetworkFirewallV1:
    type: "object"
    properties:
      id:
        type: "string"
        readOnly: true
      label:
        type: "string"

  ContentDeliveryNetworkV1:
    ...   
````

Note: At the moment any path that contains path parameters (e,g: {id}) is considered a sub-resource.

### How do sub-resources look like in the terraform configuration file?

When the OpenAPI Terraform provider would configured itself at runtime based on the above OpenAPI document, it will end up
exposing two resources. The cdns and the firewalls. 

The corresponding sub-resource Terraform configuration will contain the properties as defined in the model definition
configured in the POST operation AND the parent properties. The parent property/properties are added automatically by the
provider when creating the resource schema. This enables the user to link subresources to parent resources from the terraform
configuration file. The following example describes how the above OpenAPI document will be translated into the Terraform 
configuration: 

````
provider "openapi" {}

# Corresponding URI /v1/cdns/
resource "openapi_cdns_v1" "my_cdn_v1" {
  ....
}

# Corresponding URI /v1/cdns/{parent_id}/v1/firewalls/
resource "openapi_cdns_v1_firewalls_v1" "my_firewall_v1" {
   cdns_v1_id = openapi_cdns_v1.my_cdn_v1.id
   ...
}
````

Note the property ```cdns_v1_id``` is not described in the definition ```ContentDeliveryNetworkFirewallV1```. This property was
added automatically by the provider. The parent property name will be built based on the parent URI path including the version 
(must be next to the resource name - e,g: /v1/cdns) if applicable. In this example the parent URI is ```/v1/cdns```, 
hence the parent property generated will be ```cdns_v1_id```. The ```_id``` is appended after the resource parent name.

If the cdn endpoint was not using versioning in the path (e,g: ```/cdns```), then the automatically generated property would
not have the version in the name either. The parent property name generated in this case would be ```cdns_id```.

### How will the API requests for sub-resources look like?

As mentioned previously, the schema for sub-resources contain also properties to refer to the parents. This properties will 
hold the values with the actual IDs. These IDs are then used internally by the provider to build the right URI to make the API
calls against to. 

For instance, let's say that terraform has provisioned the ```openapi_cdns_v1.my_cdn_v1``` just fine and the ID generated
by the API for the resource was ```1234```. Now it's time to provision ```openapi_cdns_v1_firewalls_v1.my_firewall_v1```
so when the OpenAPI plugin is called the ```cdns_v1_id``` property would have been populated with ```1234``` and internally
the provider will make use of this value to build the firewall URI accordingly ```/v1/cdns/1234/v1/firewalls/``` and perform
the corresponding operation (e,g: POST). The same thing applies to any other operation exposed by the resource like
GET, PUT or DELETE.

### Are multiple level subsources also supported?

The provider also supports multiple level sub-resources. For instance, if there was a nested resource under firewalls like 
the example below (just showing the second level sub-resource here):

````
  /v1/cdns/{parent_id}/v1/firewalls/{firewall_id}/rules:
    post:
      parameters:
      - name: "parent_id"
        in: "path"
        type: "string"
      - name: "firewall_id"
        in: "path"
        type: "string"        
      - in: "body"
        schema:
          $ref: "#/definitions/ContentDeliveryNetworkFirewallRulesV1"    
      ....
  /v1/cdns/{cdn_id}/v1/firewalls/{firewall_id}/rules/{id}:
    get:
      - name: "cdn_id"
        in: "path"
        type: "string"
      - name: "firewall_id"
        in: "path"
        type: "string"
      - name: "id"
        in: "path"
        type: "string"                 
      ...
````

The corresponding terraform configuration file will be:

````
provider "openapi" {}

# Corresponding URI /v1/cdns/
resource "openapi_cdns_v1" "my_cdn_v1" {
  ....
}

# Corresponding URI /v1/cdns/{parent_id}/v1/firewalls/
resource "openapi_cdns_v1_firewalls_v1" "my_firewall_v1" {
   cdns_v1_id = openapi_cdns_v1.my_cdn_v1.id
   ...
}

# Corresponding URI /v1/cdns/{parent_id}/v1/firewalls/{firewall_id}/rules
resource "openapi_cdns_v1_firewalls_v1" "my_firewall_v1" {
   cdns_v1_id = openapi_cdns_v1.my_cdn_v1.id
   cdns_v1_firewalls_v1_id = openapi_cdns_v1_firewalls_v1.my_firewall_v1.id
   ...
}
````

Note that the parent property name for firewall contained not only the firewall but also the combination of the parent resource
name ```cdns_v1_firewalls_v1_id```. This is intentional to make it explicit what the hierarchy looks like and also to avoid
any potential conflict with the model definition containing a property with the same name.