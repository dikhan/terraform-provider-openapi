PROVIDER_NAME?=""
TF_CMD?="plan"

TF_INSTALLED_PLUGINS_PATH="$(HOME)/.terraform.d/plugins"

TEST_PACKAGES?=$$(go list ./... | grep -v "/examples\|/vendor")
GOFMT_FILES?=$$(find . -name '*.go' | grep -v 'examples\|vendor')

TF_PROVIDER_NAMING_CONVENTION="terraform-provider-"
TF_OPENAPI_PROVIDER_PLUGIN_NAME="$(TF_PROVIDER_NAMING_CONVENTION)openapi"

default: build

all: test build

build:
	@echo "[INFO] Building $(TF_OPENAPI_PROVIDER_PLUGIN_NAME) binary"
	@go build -o $(TF_OPENAPI_PROVIDER_PLUGIN_NAME)

fmt:
	@echo "[INFO] Running gofmt on the current directory"
	gofmt -w $(GOFMT_FILES)

vet:
	@echo "[INFO] Running go vet on the current directory"
	@go vet $(TEST_PACKAGES) ; if [ $$? -eq 1 ]; then \
		echo "[ERROR] Vet found suspicious constructs. Please fix the reported constructs before submitting code for review"; \
		exit 1; \
	fi

lint:
	@echo "[INFO] Running golint on the current directory"
	@go get -u github.com/golang/lint/golint
	@golint -set_exit_status $(TEST_PACKAGES)

test: fmt vet lint
	@echo "[INFO] Testing $(TF_OPENAPI_PROVIDER_PLUGIN_NAME)"
	@go test -v -cover $(TEST_PACKAGES) ; if [ $$? -eq 1 ]; then \
		echo "[ERROR] Test returned with failures. Please go through the different scenarios and fix the tests that are failing"; \
		exit 1; \
	fi

pre-requirements:
	@echo "[INFO] Creating $(TF_INSTALLED_PLUGINS_PATH) if it does not exist"
	@[ -d $(TF_INSTALLED_PLUGINS_PATH) ] || mkdir -p $(TF_INSTALLED_PLUGINS_PATH)

install: build pre-requirements
	$(call install_plugin,$(PROVIDER_NAME))

local-env-down: fmt
	@echo "[INFO] Tearing down local environment (clean up task)"
	@docker-compose -f ./build/docker-compose.yml down

local-env: fmt
	@echo "[INFO] Bringing up local environment"
	@docker-compose -f ./build/docker-compose.yml up --detach --build --force-recreate

run-terraform-example-swaggercodegen: build pre-requirements
	$(call run_terraform_example,"https://localhost:8443/swagger.yaml",swaggercodegen)

run-terraform-example-goa: build pre-requirements
	$(call run_terraform_example,"http://localhost:9090/swagger/swagger.yaml",goa)

define install_plugin
    @$(eval TF_PROVIDER_PLUGIN_NAME := $(TF_PROVIDER_NAMING_CONVENTION)$(1))

	@echo "[INFO] Installing $(TF_PROVIDER_PLUGIN_NAME) binary in -> $(TF_INSTALLED_PLUGINS_PATH)"
	@mv ./$(TF_OPENAPI_PROVIDER_PLUGIN_NAME) $(TF_INSTALLED_PLUGINS_PATH)
	@ln -sF $(TF_INSTALLED_PLUGINS_PATH)/$(TF_OPENAPI_PROVIDER_PLUGIN_NAME) $(TF_INSTALLED_PLUGINS_PATH)/$(TF_PROVIDER_PLUGIN_NAME)
endef

define run_terraform_example
    @$(eval OTF_VAR_SWAGGER_URL := $(1))
    @$(eval PROVIDER_NAME := $(2))

	$(call install_plugin,$(PROVIDER_NAME))

    @$(eval TF_EXAMPLE_FOLDER := ./examples/$(PROVIDER_NAME))

	@echo "[INFO] Performing sanity check against the service provider's swagger endpoint '$(OTF_VAR_SWAGGER_URL)'"
	@$(eval SWAGGER_HTTP_STATUS := $(shell curl -s -o /dev/null -w '%{http_code}' $(OTF_VAR_SWAGGER_URL) -k))
	@rm -f ./examples/$(PROVIDER_NAME)/terraform.tfstate
	@if [ "$(SWAGGER_HTTP_STATUS)" = 200 ]; then\
        echo "[INFO] Terraform Configuration file located at $(TF_EXAMPLE_FOLDER)";\
        echo "[INFO] Executing TF command: OTF_INSECURE_SKIP_VERIFY=true OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init && terraform ${TF_CMD}";\
        cd $(TF_EXAMPLE_FOLDER) && export OTF_INSECURE_SKIP_VERIFY=true OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init && terraform ${TF_CMD};\
    else\
        echo "[ERROR] Sanity check against swagger endpoint[$(OTF_VAR_SWAGGER_URL)] failed...Please make sure the service provider API is up and running and exposes swagger APIs on '$(OTF_VAR_SWAGGER_URL)'";\
    fi
endef

.PHONY: all build fmt vet lint test run_terraform
