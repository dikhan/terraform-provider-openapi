# Releasing a new version of terraform-provider-openapi

Terraform Provider OpenAPI performs releases automatically via Travis CI. 

## Pre-requirements

- [Go 1.12.4](https://golang.org/)
- [GoReleaser](https://goreleaser.com/)
- [Github Release Notes](https://github.com/buchanae/github-release-notes)

## How to release a new version

This project uses [semantic version 2.0.0](https://semver.org/) and tags are created following this pattern (MAJOR.MINOR.PATCH). To
create a new version follow the steps below:

### Check the latest version released

````
$ make latest-tag
[INFO] Latest tag released...
v0.1.0
````

### Update version file

Per the latest version released, determine whether the new version is a major, minor, or patch and update the 
[version file](https://github.com/dikhan/terraform-provider-openapi/blob/master/version) accordingly. For instance, given 
the above latest-tag output (v0.1.0), if we wanted to release a new minor version, the content of the version file should be 
`0.2.0`. Then follow the regular [process for contributing code](https://github.com/dikhan/terraform-provider-openapi/blob/master/.github/CONTRIBUTING.md#contributing-code)
to push the changes:

To confirm the new version to be released:
````
$ cat version 
0.2.0
````

Note - For new releases, the PR title should follow the below convention (replace 0.2.0 with the new version):

```
[NewRelease] v0.2.0
```

With the above title, the new version would be v0.2.0.

The PR will need one admin approval. Once approved and merged, the release will be automatically performed by Travis CI.

## How to release a new alpha version

Alpha means the features haven't been locked down, it's an exploratory phase. Releasing an alpha version enable users to 
start early adopting the version even though it may not be production ready yet and functionality might still change until
the final version released. The following targets have been created to help create alpha release versions:

- To create a new alpha release version run the following command:
````
RELEASE_ALPHA_VERSION=2.1.0 make release-alpha
````
This will create a local tag in the form v$(RELEASE_ALPHA_VERSION)-alpha.1. For the example above that would be `v2.1.0-alpha.1` and
push the tag to origin.

- To delete a previously create alpha version run the following command:
````
RELEASE_ALPHA_VERSION=2.1.0 make delete-release-alpha
````
This will delete a local tag in the form v$(RELEASE_ALPHA_VERSION)-alpha.1. For the example above that would be `v2.1.0-alpha.1` and
delete the tag in origin.