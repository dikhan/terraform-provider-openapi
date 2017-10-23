# terraform-provider-api
PoC terraform provider that configures itself based on the resources exposed by the api

## How to run the exaple?

The backend needs to be up and running. This can be done executing the following command:
```
docker-compose up --build
```

Once the backend is up, the following command will read main.tf file and execute terraform plan:  
```
go build -o terraform-provider-sp && terraform init && TF_LOG=DEBUG API_DISCOVERY_URL="http://localhost:8080/v2/" terraform plan
```
