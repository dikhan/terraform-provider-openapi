# terraform-provider-api
PoC terraform provider that configures itself based on the resources exposed by the api. The service provider has to 
expose an end point that returns the contents of the swagger file (OpenApi specification) file containing the API resources 
following the .

### Pre-requirements

The service provider API has to expose a discovery endpoint that serves the API's swagger definition. The swagger file
must follow the OpenAPI spec.

## How to run the example?

A docker compose file has been created to ease the execution of the example. In order to bring up both service providers
and also expose their swagger documentation in two separate UIs, please run the following command:
```
docker-compose up --build --force-recreate
```

Once the backend is up, the following command will read main.tf file and execute terraform plan:  
```
go build -o terraform-provider-sp1 && go build -o terraform-provider-sp2 && terraform init && terraform plan
```

The expected output would be:

```

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  + sp1_cdns.my_cdn
      id:          <computed>
      hostnames.#: "1"
      hostnames.0: "origin.com"
      ips.#:       "1"
      ips.0:       "127.0.0.1"
      label:       "label"

  + sp2_users.my_user
      id:          <computed>
      email:       "info@server.com"
      first_name:  "Daniel"
      last_name:   "Khan"
      password:    "password1"
      phone:       "6049991234"
      username:    "dikhan"


Plan: 2 to add, 0 to change, 0 to destroy.

```