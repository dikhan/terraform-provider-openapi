# OpenAPI Terraform Provider v3.0.0 release notes

This version of the OpenAPI Terraform provider continues to integrate Terraform SDK 2.0; however there's been some non backwards
compatible changes in functionality that rendered a major upgrade.

This document describes the changes applied in the plugin.

## Terraform CLI Compatibility

Same as stated in [OpenAPI Terraform Provider v2.0.0 release notes - Terraform CLI Compatibility](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/major_version_release_notes/openapi_terraform_provider_v2.0.0.md#terraform-cli-compatibility)

## Go Compatibility

The OpenAPI Terraform Plugin is built in Go. Currently, that means Go 1.17 or later must be used when building this provider.

## What's changed?

### Version 3 of the 'github.com/dikhan/terraform-provider-openapi' Module

As part of the breaking changes, the OpenAPI Terraform Plugin SDK Go Module has been upgraded to v3. This involves changing 
import paths from `github.com/dikhan/terraform-provider-openapi/v2` to `github.com/dikhan/terraform-provider-openapi/v3` for the
custom terraform providers repositories that use the `github.com/dikhan/terraform-provider-openapi/v2` as the parent. This is done
usually to leverage the OpenAPI Terraform plugin capabilities using this repo as a library that can be imported and customizing the release cycle
of the provider as well as the name etc.

### Deprecated multi-region on resource name level

As notified on the previous major release v2, this major upgrade removes support for [Multi-region on resource name level](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/major_version_release_notes/openapi_terraform_provider_v2.0.0.md#deprecation-notices).

This feature was introduced in early days of the provider to support multi-region. However, this model was not compatible 
with Terraform modules since the region was attached to the resource name. Later, multi-region support was added on the provider 
level as expected by Terraform enabling the provider to be used in modules and enabling users to specify the regions on the provider level. 
The multi-region support on the resource name level has been still supported till now but it's time to say goodbye to features 
that had some value in the past but now don't make sense and makes the code more difficult to read and maintain.

**Note:** Please note that the above only refers to the multi-region on a **resource level**. Multi-region on the **provider configuration level** 
is supported and documented [here](https://github.com/dikhan/terraform-provider-openapi/blob/master/docs/how_to.md#multi-region-configuration)  

### Deprecated plugin_version property from OpenAPI config file SchemaV1

The OpenAPI Terraform config file enabled the user to specify the plugin_version as a way to ensure the config is targeted 
towards a specific version of the OpenAPI plugin binary. This was a feature added in early stages of the provider as a security 
mechanism to validate that the expected version of the binary was being used or fail close at runtime otherwise. 
This mechanism is no longer working as intended causing issues for users when they have different Terraform configurations 
each pointing at different versions of the OpenAPI Terraform provider but yet the global OpenAPI Terraform plugin configuration 
is set with one specific version.

## Deprecation notices

The following functionality will be deprecated in future versions of the OpenAPI Terraform Provider.
