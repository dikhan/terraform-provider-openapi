PROVIDER_NAME?="sp"
OTF_VAR_SWAGGER_URL?="https://localhost:8443/swagger.yaml"
OTF_INSECURE_SKIP_VERIFY?="false"
TF_CMD?="plan"
TF_CONFIGURATION_FILE?="$$(pwd)/examples/cdn"

TF_INSTALLED_PLUGINS_PATH="$(HOME)/.terraform.d/plugins"

TEST_PACKAGES?=$$(go list ./... | grep -v vendor/)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v 'vendor')

default: build

all: test build

build: fmt
	@echo "[INFO] Building terraform-provider binary"
	@go build -o terraform-provider
	@echo "[INFO] Creating a symlink to the specific provider name"
	@ln -sF terraform-provider terraform-provider-$(PROVIDER_NAME)

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
	@echo "[INFO] Testing terraform-provider-openapi"
	@go test -v -cover $(TEST_PACKAGES) ; if [ $$? -eq 1 ]; then \
		echo "[ERROR] Test returned with failures. Please go through the different scenarios and fix the tests that are failing"; \
		exit 1; \
	fi

deps:
	@echo "[INFO] Creating $(TF_INSTALLED_PLUGINS_PATH) if it does not exist"
	@[ -d $(TF_INSTALLED_PLUGINS_PATH) ] || mkdir -p $(TF_INSTALLED_PLUGINS_PATH)

install-no-tests: build deps
	@echo "[INFO] Installing terraform-provider binary in -> $(TF_INSTALLED_PLUGINS_PATH)"
	@mv ./terraform-provider $(TF_INSTALLED_PLUGINS_PATH)
	@rm -f terraform-provider-$(PROVIDER_NAME)
	@echo "[INFO] Creating a symlink to the specific provider name"
	@ln -sF $(TF_INSTALLED_PLUGINS_PATH)/terraform-provider $(TF_INSTALLED_PLUGINS_PATH)/terraform-provider-$(PROVIDER_NAME)

install: test deps
	@echo "[INFO] Installing terraform-provider binary in -> $(TF_INSTALLED_PLUGINS_PATH)"
	@mv ./terraform-provider $(TF_INSTALLED_PLUGINS_PATH)
	@rm -f terraform-provider-$(PROVIDER_NAME)
	@echo "[INFO] Creating a symlink to the specific provider name"
	@ln -sF $(TF_INSTALLED_PLUGINS_PATH)/terraform-provider $(TF_INSTALLED_PLUGINS_PATH)/terraform-provider-$(PROVIDER_NAME)

local-env-down: fmt
	@echo "[INFO] Tearing down local environment (clean up task)"
	@docker-compose -f ./build/docker-compose.yml down

local-env: fmt
	@echo "[INFO] Bringing up local environment"
	@docker-compose -f ./build/docker-compose.yml up --build --force-recreate

run_terraform: install-no-tests
	@echo "[INFO] Performing sanity check against the service provider's swagger endpoint '$(OTF_VAR_SWAGGER_URL)'"
	@$(eval SWAGGER_HTTP_STATUS := $(shell curl -s -o /dev/null -w '%{http_code}' $(OTF_VAR_SWAGGER_URL) -k))
ifeq ($(PROVIDER_NAME),"sp")
	echo "[INFO] Setting OTF_INSECURE_SKIP_VERIFY value to true as example server uses self-signed certificate"
	$(eval override OTF_INSECURE_SKIP_VERIFY="true")
endif
	@if [ "$(SWAGGER_HTTP_STATUS)" = 200 ]; then\
        echo "[INFO] Terraform Configuration file located at $(TF_CONFIGURATION_FILE)";\
        echo "[INFO] Service provider swagger end point '$(OTF_VAR_SWAGGER_URL)' is reachable";\
        echo "[INFO] Executing TF command: OTF_INSECURE_SKIP_VERIFY=$(OTF_INSECURE_SKIP_VERIFY) OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init && terraform ${TF_CMD}";\
        cd $(TF_CONFIGURATION_FILE) && export OTF_INSECURE_SKIP_VERIFY="$(OTF_INSECURE_SKIP_VERIFY)" OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init && terraform ${TF_CMD};\
    else\
        echo "[ERROR] Sanity check against swagger endpoint[$(OTF_VAR_SWAGGER_URL)] failed...Please make sure the service provider API is up and running and exposes swagger APIs on '$(OTF_VAR_SWAGGER_URL)'";\
    fi

.PHONY: all build fmt vet lint test run_terraform
