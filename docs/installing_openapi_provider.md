# OpenAPI Terraform provider installation

This document describes the different ways on how the OpenAPI Terraform
provider can be installed.

## OpenAPI Terraform provider 'manual' installation

- Download most recent release for your architecture (macOS/Linux) from [release](https://github.com/dikhan/terraform-provider-openapi/releases)
page.
- Extract contents of tar ball and copy the terraform-provider-openapi binary into your  ````~/.terraform.d/plugins````
folder as described in the [Terraform documentation on how to install plugins](https://www.terraform.io/docs/extend/how-terraform-works.html#discovery).
- After installing the plugin, rename the binary file to have your provider's name:

    ````
    $ cd ~/.terraform.d/plugins
    $ mv terraform-provider-openapi terraform-provider-<your_provider_name>
    $ ls -la
    total 29656
    drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
    drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
    -rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-<your_provider_name>
    ````

Where ````<your_provider_name>```` should be replaced with your provider's name. This is the name that will also be used
in the Terraform tf files to refer to the provider resources.

````
$ cat main.tf
resource "<your_provider_name>_<resource_name>" "my_resource" {
    ...
    ...
}
````

## OpenAPI Terraform provider 'script' installation

In order to simplify the installation process for this provider, a convenient install script is provided and can also be
used as follows:

- Check out this repo and execute the install script:

````
$ git clone git@github.com:dikhan/terraform-provider-openapi.git
$ cd terraform-provider-openapi/scripts
$ PROVIDER_NAME=goa ./install.sh --provider-name $PROVIDER_NAME
````

- Or directly by downloading the install script using curl:

````
$ export PROVIDER_NAME=goa && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME
````

The install script will download the most recent [terraform-provider-openapi release](https://github.com/dikhan/terraform-provider-openapi/releases)
and install it in the terraform plugins folder ````~/.terraform.d/plugins```` as described above. The terraform plugins folder should contain the newly
installed open api customer terraform provider with the name provider in the installation (PROVIDER_NAME=goa) ```terraform-provider-goa```.

````
$ ls -la ~/.terraform.d/plugins
total 29656
drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
-rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-goa
````
