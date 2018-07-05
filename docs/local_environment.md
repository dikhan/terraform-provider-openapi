# Setting up a local environment

The Makefile has some convenient targets configured to help you bring up/tear down the two example APIs provided for 
development purposes.

## Bringing up the example API servers

The following target can be executed to bring up the [example API services](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples)
which expose their APIs documentation in a swagger file:

```
$ make local-env
...
Successfully tagged goa-service-provider-api:latest
Recreating goa-service-provider-api            ... done
Recreating swagger-ui-swaggercodegen           ... done
Recreating swaggercodegen-service-provider-api ... done
```

Internally, the make target 'local-env' uses the following [docker-compose](https://github.com/dikhan/terraform-provider-openapi/blob/master/build/docker-compose.yml) 
file that contains the definitions for the [example API services](https://github.com/dikhan/terraform-provider-openapi/tree/master/examples). 

Additionally, it will also render a UI from the swagger file exposed by the 'swaggercodegen-service-provider-api' API 
server that can be accessed from the browser at ``https://localhost:8443``.

The UI rendered feeds from the swagger file located at [docker-compose](https://github.com/dikhan/terraform-provider-openapi/blob/master/service_provider_example/resources/swagger.yaml)

## Trying out the service provider example

### Installing the openapi terraform provider plugin binary

Once docker-compose is done bringing up the example API servers, the following command can be executed to compile and install 
the openapi terraform provider binary:

```
$ PROVIDER_NAME="<provider_name>" make install
[INFO] Building terraform-provider-openapi binary
[INFO] Creating /Users/dikhan/.terraform.d/plugins if it does not exist
[INFO] Installing terraform-provider-<provider_name> binary in -> /Users/dikhan/.terraform.d/plugins
```

Where ````<your_provider_name>```` should be replaced with your provider's name.

The above ```make install``` command will compile the provider from the source code, install the compiled binary terraform-provider-openapi 
in the terraform plugin folder ````~/.terraform.d/plugins```` and create a symlink from terraform-provider-goa to the
binary compiled. The reason why a symlink is created is so the same compiled binary can be reused by multiple openapi providers 
and also reduces the number of providers to support.

````
$ ls -la ~/.terraform.d/plugins
total 29656
drwxr-xr-x  4 dikhan  staff       128  3 Jul 15:13 .
drwxr-xr-x  4 dikhan  staff       128  3 Jul 13:53 ..
-rwxr-xr-x  1 dikhan  staff  15182644 29 Jun 16:21 terraform-provider-openapi
lrwxr-xr-x  1 dikhan  staff        63  3 Jul 15:11 terraform-provider-<provider_name> -> /Users/dikhan/.terraform.d/plugins/terraform-provider-openapi
````

For the sake of the example, let's pick the swaggercodegen service example and install the plugin:

````
$ PROVIDER_NAME="swaggercodegen" make install
[INFO] Building terraform-provider-openapi binary
[INFO] Creating /Users/dkhanram/.terraform.d/plugins if it does not exist
[INFO] Installing terraform-provider-swaggercodegen binary in -> /Users/dkhanram/.terraform.d/plugins
````

And now let's check that the binaries are indeed installed in the terraform plugins folder:

````
ls -la /Users/dikhan/.terraform.d/plugins
total 44120
drwxr-xr-x  5 dikhan  staff       160  5 Jul 16:02 .
drwxr-xr-x  5 dikhan  staff       160  4 Jul 13:19 ..
-rwxr-xr-x  1 dikhan  staff  22270452  5 Jul 16:01 terraform-provider-openapi
lrwxr-xr-x  1 dikhan  staff        63  5 Jul 16:01 terraform-provider-swaggercodegen -> /Users/dkhanram/.terraform.d/plugins/terraform-provider-openapi
````

### Running the openapi terraform provider

Having the openapi provider binary installed, we can now execute terraform commands.
 
#### Executing terraform plan

First we need to access the folder where the the .tf file is localed. An example of swaggercodegen [main.tf](https://github.com/dikhan/terraform-provider-openapi/blob/master/examples/swaggercodegen/main.tf)
file is provided in the examples folder.

```
$ cd $GOPATH/github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen
$ terraform init && OTF_INSECURE_SKIP_VERIFY="true" OTF_VAR_swaggercodegen_SWAGGER_URL="https://localhost:8443/swagger.yaml" terraform plan

....

Initializing provider plugins...

Terraform has been successfully initialized!

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + swaggercodegen_cdns_v1.my_cdn
      id:              <computed>
      example_boolean: "true"
      example_int:     "12"
      example_number:  "1.12"
      hostnames.#:     "1"
      hostnames.0:     "origin.com"
      ips.#:           "1"
      ips.0:           "127.0.0.1"
      label:           "label"


Plan: 1 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

...
```

Notice that OTF_INSECURE_SKIP_VERIFY="true" is passed in to the command, this is needed due to the fact that the server
uses a self-signed cert. This will make the provider's internal http client skip the certificate verification. This is
**not recommended** for regular use and this env variable OTF_INSECURE_SKIP_VERIFY should only be set when the server hosting
the swagger file is known and trusted but does not have a cert signed by the usually trusted CAs. 

The OpenAPI terraform provider expects as input the URL where the service provider is exposing the swagger file. This
can be passed in defining as an environment variable with a name tha follows "OTF_VAR_{PROVIDER_NAME}_SWAGGER_URL" being '{PROVIDER_NAME}'
the name of the provider specified in the binary when compiling the plugin - 'swaggercodegen' in the example above.

When defining the env variable, {PROVIDER_NAME} can be lower case or upper case.

This means that the plugin was able to read the swagger file exposed by the service provider example, load it
up and set up the terraform provider dynamically with the resources exposed by 'cdn-service-provider-api' being one of
them 'cdns'.

#### Executing terraform apply

Now we can run terraform apply to see the plugin do its magic:

```
$ terraform init && OTF_INSECURE_SKIP_VERIFY="true" OTF_VAR_swaggercodegensp_SWAGGER_URL="https://localhost:8443/swagger.yaml" terraform apply

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + swaggercodegen_cdns_v1.my_cdn
      id:              <computed>
      example_boolean: "true"
      example_int:     "12"
      example_number:  "1.12"
      hostnames.#:     "1"
      hostnames.0:     "origin.com"
      ips.#:           "1"
      ips.0:           "127.0.0.1"
      label:           "label"


Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

swaggercodegen_cdns_v1.my_cdn: Creating...
  example_boolean: "" => "true"
  example_int:     "" => "12"
  example_number:  "" => "1.12"
  hostnames.#:     "" => "1"
  hostnames.0:     "" => "origin.com"
  ips.#:           "" => "1"
  ips.0:           "" => "127.0.0.1"
  label:           "" => "label"
swaggercodegen_cdns_v1.my_cdn: Creation complete after 0s (ID: 0d586a2b-f9d2-496d-a09d-a6ef06fca4a1)

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

And a 'terraform.tfstate' should have been created by terraform containing the state of the new resource created.

````
{
    "version": 3,
    "terraform_version": "0.11.7",
    "serial": 1,
    "lineage": "dd9d43ed-6f85-e4b2-e9f3-e31bce86707d",
    "modules": [
        {
            "path": [
                "root"
            ],
            "outputs": {},
            "resources": {
                "swaggercodegen_cdns_v1.my_cdn": {
                    "type": "swaggercodegen_cdns_v1",
                    "depends_on": [],
                    "primary": {
                        "id": "0d586a2b-f9d2-496d-a09d-a6ef06fca4a1",
                        "attributes": {
                            "example_boolean": "true",
                            "example_int": "12",
                            "example_number": "1.12",
                            "hostnames.#": "1",
                            "hostnames.0": "origin.com",
                            "id": "0d586a2b-f9d2-496d-a09d-a6ef06fca4a1",
                            "ips.#": "1",
                            "ips.0": "127.0.0.1",
                            "label": "label"
                        },
                        "meta": {},
                        "tainted": false
                    },
                    "deposed": [],
                    "provider": "provider.swaggercodegen"
                }
            },
            "depends_on": []
        }
    ]
}
````

## Running the example via Makefile

A convenient [Makefile](https://github.com/dikhan/terraform-provider-openapi/blob/master/Makefile) is provided allowing 
the user to execute the above in just one command as follows:
```
$ make local-env-down local-env run-terraform-example-swaggercodegen TF_CMD=plan
```

The above command will bring up the example server API and install the binary plugin in the terraform plugin folder. 

When calling terraform it will pass all the required environment variables mentioned above using the example values:

````
terraform init && OTF_INSECURE_SKIP_VERIFY="true" OTF_VAR_sp_SWAGGER_URL="https://localhost:8443/swagger.yaml" terraform plan
````