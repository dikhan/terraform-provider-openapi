# OpenAPI Terraform provider installation

This document describes the different ways on how the OpenAPI Terraform provider can be installed. Terraform expects plugins to
be installed in specific locations depending on the version of Terraform used. 

## OpenAPI Terraform provider 'manual' installation

### OpenAPI provider installation instructions for Terraform v0.12 

If you are using Terraform v0.12, the plugins must be installed following [Terraform's v0.12 installation instructions](https://www.terraform.io/docs/plugins/basics.html#installing-plugins). The following
steps describe how to install the OpenAPI Terraform provider when using Terraform 0.12.

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

### OpenAPI provider installation instructions for Terraform >= v0.13 

If you are using Terraform v0.13, the plugins must be installed following [Terraform's v0.13 installation instructions](https://www.terraform.io/docs/configuration/provider-requirements.html#in-house-providers). The following
steps describe how to install the OpenAPI Terraform provider when using Terraform v0.13 or greater.

- Download most recent release for your architecture (macOS/Linux) from [release](https://github.com/dikhan/terraform-provider-openapi/releases)
page.
- Terraform >= v0.13 introduced the concept of plugin's source address, a string composed by `[<HOSTNAME>/]<NAMESPACE>/<TYPE>` which 
tells Terraform the primary location where the plugin can be downloaded and installed. Terraform automatically tries to download
providers from the Terraform's public registry unless the source address configuration for the provider specifies a different `hostname`. This hostname
does not actually need to resolve in DNS, Terraform will use the placeholder hostname to look up the provider binary in the local file system. In order to
tell Terraform where to look up the provider, you will need to first update your .tf file to include the providers source configuration as shown below:

````
$ cat main.tf

terraform {
  required_providers {
    <your_provider_name> = {
      source  = "<hostname>/<namespace>/<your_provider_name>"
      version = ">= <version>"
    }
  }
}

resource "<your_provider_name>_<resource_name>" "my_resource" {
    ...
}
````

Where 
- `<hostname>` is the placeholder hostname of your choice (eg: `terraform.example.com`)
- `<namespace>` is the namespace of your choice (eg: `examplecorp`)
- `<your_provider_name>` should be replaced with your provider's name. This is the name that will also be used in the Terraform 
tf files to refer to the provider resources ``resource "<your_provider_name>_<resource_name>" "my_resource" {}``.
- `<version>` is the version of your provider. Since, the actual plugin used is the OpenAPI Terraform plugin it's recommended
 to pin the version of the OpenAPI plugin version used (eg: `2.0.0`)

- Now you can extract the contents of the tar ball downloaded previously and copy the terraform-provider-openapi binary into the 
following file path: ````~/.terraform.d/plugins/<hostname>/<namespace>/<your_provider_name>/<version>/darwin_amd64/terraform-provider-<your_provider_name>````. Note, this
instruction assumes you are installing the provider using a Mac with Darwin OS and `amd64` architecture.

## OpenAPI Terraform provider 'script' installation

In order to simplify the installation process the following convenient installation script can be used which can detect the version
of Terraform in use and download and install the OpenAPI Terraform provider in the expected Terraform installation path automatically. In order
to execute the script the `--provider-name` must be specified.

### If using Terraform v0.12

- Check out this repo and execute the installation script as follows:

````
$ git clone git@github.com:dikhan/terraform-provider-openapi.git
$ cd terraform-provider-openapi/scripts
$ PROVIDER_NAME=myprovidername ./install.sh --provider-name $PROVIDER_NAME
````

- Or directly by downloading the installation script using curl:

````
$ export PROVIDER_NAME=myprovidername && curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name $PROVIDER_NAME
````

The installation script will download the most recent [terraform-provider-openapi release](https://github.com/dikhan/terraform-provider-openapi/releases)
and install it in the terraform plugins folder ````~/.terraform.d/plugins````. The terraform plugins folder should contain the newly
installed open api customer terraform provider with the name provided in the installation (PROVIDER_NAME=myprovidername) ```terraform-provider-myprovidername```.

````
$ ls -la ~/.terraform.d/plugins
total 29656
drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
-rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-myprovidername
````

### If using Terraform >= v0.13

Beside the `--provider-name` argument, the installation script accepts also as input argument the provider's source address following arguments 
if Terraform >= v0.13 is in use. This enables the user to specify a custom <HOSTNAME>/<NAMESPACE>. The default value is `terraform.example.com/examplecorp` if the source
address is not provided. For instance:

````
$ curl -fsSL https://raw.githubusercontent.com/dikhan/terraform-provider-openapi/master/scripts/install.sh | bash -s -- --provider-name "myprovidername" --provider-source-address "terraform.example.com/examplecorp"
````

The above will download and install the latest release of the OpenAPI Terraform provider and install the plugin using the
provider name provided `myprovidername` in the expected Terraform's plugin installation directory based on the source address provided `terraform.example.com/examplecorp`:

````
$ ls -la ~/.terraform.d/plugins/terraform.example.com/examplecorp/myprovidername/1.0.0/darwin_amd64/terraform-provider-myprovidername
total 29656
drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
-rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-myprovidername
````