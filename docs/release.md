# Releasing terraform-provider-openapi

Currently, the release process for new versions of the terraform-provider-openapi involves some manual steps. The process
is described in this document.

## Pre-requirements

- [Go 1.10.1](https://golang.org/)
- [GoReleaser](https://goreleaser.com/)

## How to release a new version

This project uses [semantic version 2.0.0](https://semver.org/) and tags are created following this pattern (MAJOR.MINOR.PATCH). To
create a new version follow the steps below:

- Check the latest version released:

````
$ make latest-tag
[INFO] Latest tag released...
v0.1.0
````

- Update [install script](https://github.com/dikhan/terraform-provider-openapi/blob/master/scripts/install.sh#L61) with the 
latest version created, commit and push to origin:

````
# installation variables
LATEST_RELEASE_VERSION=0.1.1
````

- Create a new tag that will be associated with the new release. Based on the tag version displayed from the previous
command, depending on the type of release go ahead and change the major, minor or patch number. In this example, we are
releasing a patch version, hence that last digit will be pumped up:

````
$ git tag -a v0.1.1 -m "Release message"
$ git push origin v0.1.1
````

- Perform the release by running goreleaser:

````
GITHUB_TOKEN="YOUR_TOKEN" goreleaser --rm-dist
````

The file [.goreleaser.yml](../.goreleaser.yml) contains the configuration used by gorelease to build and release the
new version. Current configuration will build one binary in 64bit architecture for each of the operating systems supported
Linux and MacOS.

Alternately, you can run the following command to perform the release in one liner:

````
RELEASE_TAG="v0.1.1" RELEASE_MESSAGE="Release message" GITHUB_TOKEN="YOUR_TOKEN" make release-version
````

If the command above fails for some reason most likely the new tag would have been created and pushed. Hence in order to run
the command again the latest tag would need to be cleaned up as follows:

````
RELEASE_TAG=v0.1.1 make delete-tag
````