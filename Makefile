PROVIDER_NAME ?=  sp

local_deploy:
	docker-compose up -d --build --force-recreate
	build

test: 
	cd "terraform_provider_api" && go vet -v  $$(go list ./... | grep -v /vendor/) && \
	go get -u github.com/golang/lint/golint && \
	golint -set_exit_status  $$(go list ./... | grep -v /vendor/ ) && \
	go test -v -cover  $$(go list ./... | grep -v /vendor/) 

build: 
	cd "terraform_provider_api" && go build -o terraform-provider-$(PROVIDER_NAME)
