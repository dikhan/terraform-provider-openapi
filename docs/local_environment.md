# Setting up a local environment

## Running docker compose

A [docker-compose](https://github.com/dikhan/terraform-provider-openapi/blob/master/docker-compose.yml) file has been created to ease the execution of an example. This will start up
a server that exposes its APIs in a swagger file.

```
docker-compose up --build --force-recreate
```

In addition, it will also render a UI from the swagger file exposed by the API server that can be accessed from the
browser at:

```
https://localhost:8443
```

The UI rendered feeds from the swagger file located at [docker-compose](https://github.com/dikhan/terraform-provider-openapi/blob/master/service_provider_example/resources/swagger.yaml)

## Trying out the service provider example

### Building the terraform provider plugin binary

Once docker-compose is done bringing up the API server, the following command will read the sample [main.tf](terraform_provider_api/main.tf)
file and build the terraform provider plugin:
```
$ cd terraform_provider_api
$ go build -o terraform-provider-sp
```

Looking carefully at the above command, the binary is named as 'terraform-provider-sp'. The reason for this is so
terraform knows what provider binary it should call when creating resources for 'sp' provider as defined in [main.tf](terraform_provider_api/main.tf)
file.

### Running the terraform plan

With the new provider binary in place, we can now execute terraform commands.

```
$ terraform init && OTF_VAR_sp_SWAGGER_URL="https://localhost:8443/swagger.yaml" terraform plan
```

The OpenAPI terraform provider expects as input the URL where the service provider is exposing the swagger file. This
can be passed in defining as an environment variable with a name tha follows "OTF_VAR_{PROVIDER_NAME}_SWAGGER_URL" being '{PROVIDER_NAME}'
the name of the provider specified in the binary when compiling the plugin - 'sp' in the example above.

When defining the env variable, {PROVIDER_NAME} can be lower case or upper case.

After executing terraform plan, the expected output should be:

```
....

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + sp_cdns_v1.my_cdn
      id:          <computed>
      hostnames.#: "1"
      hostnames.0: "origin.com"
      ips.#:       "1"
      ips.0:       "127.0.0.1"
      label:       "label"


Plan: 1 to add, 0 to change, 0 to destroy.

....

```

This means that the plugin was able to read the swagger file exposed by the service provider example, load it
up and set up the terraform provider dinamically with the resources exposed by 'cdn-service-provider-api' being one of
them 'cdns'.

Now we can run apply to see the plugin do its magic:

```
$ go build -o terraform-provider-sp && terraform init && terraform apply

Initializing provider plugins...

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
sp_cdns_v1.my_cdn: Creating...
  hostnames.#: "0" => "1"
  hostnames.0: "" => "origin.com"
  ips.#:       "0" => "1"
  ips.0:       "" => "127.0.0.1"
  label:       "" => "label"
sp_cdns_v1.my_cdn: Creation complete after 0s (ID: 80514498-a4d0-44e6-ad0d-22ac1023fdae)

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

And a 'terraform.tfstate' should have been created by terraform containing the state of the new resource created.


## Using the Makefile

A convenient [Makefile](https://github.com/dikhan/terraform-provider-openapi/blob/master/terraform_provider_api/Makefile)
is provided allowing the user to execute the above in just one command as follows:
```
make run_terraform TF_CMD=plan
```

The above command will build the binary with a default value - terraform-provider-sp and then it will export the
OTF_VAR_{PROVIDER_NAME}_SWAGGER_URL with a default value pointing to the local example API server that should be
running already after having run docker compose as specified. The value passed in to the TF_CMD parameter will be
the command terraform will execute (e,g: apply/plan/refresh etc)

