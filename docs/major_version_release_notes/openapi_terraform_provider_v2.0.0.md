# OpenAPI Terraform Provider v2.0.0 release notes

This version of the OpenAPI Terraform provider integrates Terraform SDK 2.0 major release which includes significant breaking changes
as specified in the [Terraform Plugin SDK v2 Upgrade Guide](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html).

This document describes the changes applied in the plugin.

## Terraform CLI Compatibility
Terraform 0.12.0 or later is needed for version v2.0.0 and later of the OpenAPI Terraform provider.

When running provider tests, Terraform 0.12.26 or later is needed for version v2.0.0 and later of the OpenAPI Terraform plugin.

## Go Compatibility
The OpenAPI Terraform Plugin is built in Go. Currently, that means Go 1.14 or later must be used when building this provider.

## What's changed?

### Version 1 of the 'github.com/dikhan/terraform-provider-openapi' Module

As part of the breaking changes, the OpenAPI Terraform Plugin SDK Go Module has been upgraded to v1. This involves changing 
import paths from `github.com/dikhan/terraform-provider-openapi` to `github.com/dikhan/terraform-provider-openapi/v2` for the
custom terraform providers repositories that use the `github.com/dikhan/terraform-provider-openapi` as the parent. This is done
usually to leverage the OpenAPI Terraform plugin capabilities using this repo as a library that can be imported and customizing the release cycle
of the provider as well as the name etc.

### Dropped Support for Terraform 0.11 and Below

Terraform Plugin SDK 2.0 only supports Terraform 0.12 and higher.

### Deprecated Support for OpenAPI property type object with internal Terraform schema representation of helper/schema.TypeMap with Elem *helper/schema.Resource

As per the addition of [more Robust Validation of helper/schema.TypeMap Elems](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#more-robust-validation-of-helper-schema-typemap-elems)
the OpenAPI Terraform plugin no longer supports properties of `type: "object"` with an internal Terraform schema representation
of helper/schema.TypeMap with Elem *helper/schema.Resource. This is due to the fact that inserting complex types into a HashMap field resulted into undefined behaviors as
documented in [Terraform Issue #22511](https://github.com/hashicorp/terraform/issues/22511#issuecomment-522609116).

Previous versions of the OpenAPI Terraform provider < v2.0.0 would translate OpenAPI definitions containing properties of `type: "object"`
that did not have the extension `x-terraform-complex-object-legacy-config` as helper/schema.TypeMap with Elem *helper/schema.Resource in the Terraform Schema 
and hence the HCL representation resulting into an argument.

The example below shows the OpenAPI definition `ContentDeliveryNetworkV1` containing an object property called `object_property` with some fields.

````
definitions:
  ContentDeliveryNetworkV1:
  ....
      object_property:
        type: "object"
        properties:
          account:
            type: string
          create_date:
            type: string
            readOnly: true
  ....
````

The resulted HCL configuration would look like:

````
resource "openapi_cdn_v1" "my_cdn" {
  ...
  object_property = {
    account = "my_account"
    create_date = "11/07/1988" // This is the workaround users will have to do in order to fix the diff issues with objects that contain readOnly properties
  }
}
````

Even though the `create_date` property was readOnly and therefore marked as computed in the Terraform schema, due to the undefined
behaviours resulting of using HashMaps with complex types, Terraform apply would work fine but subsequent plans would result into
diffs and as a workaround users would have to explicitly define the properties in the configuration files (as shown in the example above).

As of OpenAPI Terraform v2.0.0, the same object property `object_property` definition is now following the workaround suggested 
by Hashicorp Maintainers [Terraform SDK Issue #155](https://github.com/hashicorp/terraform-plugin-sdk/issues/155#issuecomment-489699737) 
as described in the [x-terraform-complex-object-legacy-config](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#x-terraform-complex-object-legacy-config) 
extension. This extension has been deprecated and property of `type: "object"` and now treated by default as `helper/schema.TypeList with Elem *helper/schema.Resource and MaxItems 1`.

The following example would be the corresponding HCL configuration for the `ContentDeliveryNetworkV1` definition above:
 
````
resource "openapi_cdn_v1" "my_cdn" {
  ...
  object_property {
   account = "my_account"
  }
}
````

Note that here `object_property` is using the [block](https://www.terraform.io/docs/configuration/syntax.html#blocks) syntax. 

It's important to remember that due to the internal schema representation of object properties being of `helper/schema.TypeList with Elem *helper/schema.Resource and MaxItems 1`
if the object property needs to be referenced from other places in the terraform configuration the list syntax needs to be used indexing
on the zero element. Example: `openapi_cdn_v1.my_cdn.object_property[0].account`

More context on this decision can be found at [Terraform SDK Issue #616](https://github.com/hashicorp/terraform-plugin-sdk/issues/616)

### Deprecated OpenAPI extension x-terraform-complex-object-legacy-config

This extension is deprecated and no longer drives special behaviour. The internal representation of the OpenAPI 'object' type
properties will always be TypeList with Elem *Resource and MaxItems 1 regardless whether the object is simple (only contains same type properties) 
or whether it's a complex object containing a mix of different types (string, int, nested objects, etc properties) and different 
configurations (eg: some properties being required, others optional, some computed etc). Refer to the [object](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#object-definitions) documentation instead. 

Previous versions of the OpenAPI Terraform provider < v2.0.0 would require this extension to be present to treat complex object properties as
recommended by Hashicorp Maintaners (see issues below) using the helper/schema.TypeList with *helper/schema.Resource as Elem and limiting the MaxItems to 1. This
behaviour is now the default for any property of type object.

- [Issue 22511](https://github.com/hashicorp/terraform/issues/22511): Objects that contain properties with different types (e,g: string, integer, etc) and configurations (e,g: some of them being computed)
- [Issue 21217](https://github.com/hashicorp/terraform/issues/21217): Objects that contain nested objects 
- [Issue 616](https://github.com/hashicorp/terraform-plugin-sdk/issues/616): Upgrading OpenAPI Terraform provider to Terraform SDK 2.0: TypeMap with Elem*Resource not supported

### Better support for resource operation timeouts and diagnostics

As part of the update of the CRUD functions to support better diagnosis via the context as documented in the [Support for Diagnostics](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-diagnostics)
upgrade guide, the OpenAPI Terraform provider now supports timeouts not only for async resource operations but also synchronous. The timeouts can
be specified in the OpenAPI document per resource operation using the [x-terraform-resource-timeout](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#xTerraformResourceTimeout)
extension.

### Support for Field-Level Descriptions

The [Terraform helper/schema.Schema](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-resource-level-and-field-level-descriptions) for OpenAPI Terraform 
compatible resources and data sources now contains a field Description which is populated based on the corresponding OpenAPI definition property's description. 

For instance, the following OpenAPI object definition would be translated into the below helper/schema.Schema
containing the `property_with_description` and its corresponding description as specified in the OpenAPI document.

```yml
definitions:
  ContentDeliveryNetworkV1:
    type: object
    properties:
      ...
      property_with_description: 
        type: string
        description: "some description for the property..."
      ... 
```

````
&schema.Resource{
        # This will be the Terraform schema of the resource using the ContentDeliveryNetworkV1 model definition
		Schema: map[string]*schema.Schema {
		    ...
		    "property_with_description": *schema.Schema {
                Type:TypeString 
                Description: "some description" 
                Optional:true 
                Required:false
                ...		    
		    }
		},
		...
	}
````

### Support for Debuggable Provider Binaries

As per [Support for Debuggable Provider Binaries](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html#support-for-debuggable-provider-binaries), the
OpenAPI Terraform binary now supports debuggers like delve attached to them. To learn more about how to enable this behaviour refer
to the [Using OpenAPI Provider documentation](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/using_openapi_provider.md#support-for-debuggable-provider-binaries)

## Deprecation notices

The following functionality will be deprecated in future versions of the OpenAPI Terraform Provider.

- [Multi-region on resource name level](./../how_to.md#xTerraformResourceRegions): This was introduced in early stages of the OpenAPI Terraform provider
to support resources that could be managed in multiple regions. However, having the region name attached to the resource name (eg: `myprovider_resource_rst1`) did
not play well with Terraform modules and as of [OpenAPI Terraform v0.10.0](https://github.com/dikhan/terraform-provider-openapi/releases/tag/v0.10.0) native [multi-region on the provider
level configuration](./../how_to.md#multiRegionConfiguration) was added. Hence, OpenAPI Terraform providers should make use of the 
latter approach when multi-region is needed and describe the OpenAPI document as described in the [multi region configuration instructions](./../how_to.md#multiRegionConfiguration).