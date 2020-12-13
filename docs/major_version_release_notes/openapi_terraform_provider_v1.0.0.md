# OpenAPI Terraform Provider v1.0.0 release notes

This version of the OpenAPI Terraform provider integrates Terraform SDK 2.0 major release which includes many breaking changes
as specified in the [Terraform Plugin SDK v2 Upgrade Guide](https://www.terraform.io/docs/extend/guides/v2-upgrade-guide.html).

This document describes the changes applied in the plugin.

## Terraform CLI Compatibility
Terraform 0.12.0 or later is needed for version 1.0.0 and later of the OpenAPI Terraform provider.

When running provider tests, Terraform 0.12.26 or later is needed for version 1.0.0 and later of the OpenAPI Terraform plugin.

## Go Compatibility
The OpenAPI Terraform Plugin is built in Go. Currently, that means Go 1.14 or later must be used when building this provider.

## What's changed?

### Version 1 of the 'github.com/dikhan/terraform-provider-openapi' Module

As part of the breaking changes, the OpenAPI Terraform Plugin SDK Go Module has been upgraded to v1. This involves changing 
import paths from github.com/dikhan/terraform-provider-openapi to github.com/dikhan/terraform-provider-openapi/v1 for the
custom terraform providers repositories that use the `github.com/dikhan/terraform-provider-openapi` as the parent. This is done
usually to leverage the OpenAPI Terraform plugin capabilities using this repo as a library that can be imported and customizing the release cycle
of the provider as well as the name etc.

### Dropped Support for Terraform 0.11 and Below

Terraform Plugin SDK 2.0 only support Terraform 0.12 and higher.

### To be removed

- Multi-region on resource name level
- Assumption that a property that has a $ref attribute is considered automatically an object so defining the type 'object' is optional (although it's recommended).