# Releasing terraform-provider-openapi

Currently, the release process for new versions of the terraform-provider-openapi involves some manual steps. The process
is described in this document.

## Pre-requirements

- [Go 1.12.4](https://golang.org/)
- [GoReleaser](https://goreleaser.com/)

## How to release a new version

This project uses [semantic version 2.0.0](https://semver.org/) and tags are created following this pattern (MAJOR.MINOR.PATCH). To
create a new version follow the steps below:

### Check the latest version released:

````
$ make latest-tag
[INFO] Latest tag released...
v0.1.0
````

### Update version file

Update the [version file](https://github.com/dikhan/terraform-provider-openapi/blob/master/version) with the 
latest version created and then create a branch, commit to it and merge to master:

````
$ cat version 
0.1.0
````

### Create release

Run the release command in one liner. The release is performed by goreleaser which uses the [.goreleaser.yml](../.goreleaser.yml) file 
to perform the release according to the configuration. The current configuration will build one binary in 64bit architecture 
for each of the operating systems supported Linux, MacOS and Windows.

This will create the appropriate tag, push it to origin/master and create the release accordingly:

````
$ GITHUB_TOKEN="YOUR_TOKEN" make release-version
````

If needed, RELEASE_TAG and RELEASE_MESSAGE env variables can be passed in to oveerride the default values which
make use of the version file by default:

````
$ RELEASE_TAG="v0.1.1" RELEASE_MESSAGE="Release message" GITHUB_TOKEN="YOUR_TOKEN" make release-version
````

Alternatively, a more manual approach can be followed:

- Create a new tag that will be associated with the new release. Based on the tag version displayed from the previous
command, depending on the type of release go ahead and change the major, minor or patch number. In this example, we are
releasing a patch version, hence that last digit will be pumped up:

````
$ git tag -a v0.1.1 -m "Release message"
$ git push origin v0.1.1
````

- Perform the release by running goreleaser:

````
$ GITHUB_TOKEN="YOUR_TOKEN" goreleaser --rm-dist
````

### Troubleshooting

#### Tag already exists

If the command above fails for some reason most likely the new tag would have been created and pushed. Hence in order to run
the command again the latest tag would need to be cleaned up as follows:

````
$ RELEASE_TAG=v0.1.1 make delete-tag
````