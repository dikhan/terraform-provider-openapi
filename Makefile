PROVIDER_NAME?=""
TF_CMD?="plan"

TF_INSTALLED_PLUGINS_PATH="$(HOME)/.terraform.d/plugins"

TEST_PACKAGES?=$$(go list ./... | grep -v "/examples\|/vendor|/integration")
INT_TEST_PACKAGES?=$$(go list ./... | grep "/tests/integration")
GOFMT_FILES?=$$(find . -name '*.go' | grep -v 'examples\|vendor')

TF_PROVIDER_NAMING_CONVENTION="terraform-provider-"
TF_OPENAPI_PROVIDER_PLUGIN_NAME="$(TF_PROVIDER_NAMING_CONVENTION)openapi"

default: build

all: test build

# make build
build:
	@echo "[INFO] Building $(TF_OPENAPI_PROVIDER_PLUGIN_NAME) binary"
	@go build -ldflags="-s -w" -o $(TF_OPENAPI_PROVIDER_PLUGIN_NAME)

# make fmt
fmt:
	@echo "[INFO] Running gofmt on the current directory"
	gofmt -w $(GOFMT_FILES)

# make vet
vet:
	@echo "[INFO] Running go vet on the current directory"
	@go vet $(TEST_PACKAGES) ; if [ $$? -eq 1 ]; then \
		echo "[ERROR] Vet found suspicious constructs. Please fix the reported constructs before submitting code for review"; \
		exit 1; \
	fi

# make lint
lint:
	@echo "[INFO] Running golint on the current directory"
	@go get -u github.com/golang/lint/golint
	@golint -set_exit_status $(TEST_PACKAGES)

# make test
test: fmt vet lint
	@echo "[INFO] Testing $(TF_OPENAPI_PROVIDER_PLUGIN_NAME)"
	@go test -v -cover $(TEST_PACKAGES) ; if [ $$? -eq 1 ]; then \
		echo "[ERROR] Test returned with failures. Please go through the different scenarios and fix the tests that are failing"; \
		exit 1; \
	fi

# make integration-test
integration-test: local-env-down local-env
	@echo "[INFO] Testing $(TF_OPENAPI_PROVIDER_PLUGIN_NAME)"
	@TF_ACC=true go test -v -cover $(INT_TEST_PACKAGES) ; if [ $$? -eq 1 ]; then \
		echo "[ERROR] Test returned with failures. Please go through the different scenarios and fix the tests that are failing"; \
		exit 1; \
	fi

pre-requirements:
	@echo "[INFO] Creating $(TF_INSTALLED_PLUGINS_PATH) if it does not exist"
	@[ -d $(TF_INSTALLED_PLUGINS_PATH) ] || mkdir -p $(TF_INSTALLED_PLUGINS_PATH)

# make install
install: build pre-requirements
	$(call install_plugin,$(PROVIDER_NAME))

# make local-env-down
local-env-down: fmt
	@echo "[INFO] Tearing down local environment (clean up task)"
	@docker-compose -f ./build/docker-compose.yml down

# make local-env
local-env: fmt
	@echo "[INFO] Bringing up local environment"
	@docker-compose -f ./build/docker-compose.yml up -d --build --force-recreate

# [TF_CMD=apply] make run-terraform-example-swaggercodegen
run-terraform-example-swaggercodegen: build pre-requirements
	$(call run_terraform_example,"https://localhost:8443/swagger.yaml",swaggercodegen)

# [TF_CMD=apply] make run-terraform-example-goa
run-terraform-example-goa: build pre-requirements
	$(call run_terraform_example,"http://localhost:9090/swagger/swagger.yaml",goa)

# make latest-tag
latest-tag:
	@echo "[INFO] Latest tag released..."
	@git for-each-ref --sort=-taggerdate --count=1 --format '%(tag)' 'v*' refs/tags

# RELEASE_TAG=v.0.1.5 make delete-tag
delete-tag:
	@echo "[INFO] Deleting tag specified $(RELEASE_TAG) (local and remote)..."
	@git tag -d $(RELEASE_TAG) | echo
	@git push origin :refs/tags/$(RELEASE_TAG) | echo

# RELEASE_TAG="v0.1.1" RELEASE_MESSAGE="v0.1.1" GITHUB_TOKEN="PERSONAL_TOKEN" make release-version
release-version:
	@echo "[INFO] Creating release tag $(RELEASE_TAG)"
	@git tag -a $(RELEASE_TAG) -m $(RELEASE_MESSAGE)
	@echo "[INFO] Pushing release tag"
	@git push origin $(RELEASE_TAG)
	@echo "[INFO] Performing release"
	@GITHUB_TOKEN=$(GITHUB_TOKEN) goreleaser --rm-dist

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
	@if [ "$(SWAGGER_HTTP_STATUS)" = 200 ]; then\
        echo "[INFO] Terraform Configuration file located at $(TF_EXAMPLE_FOLDER)";\
        echo "[INFO] Executing TF command: OTF_INSECURE_SKIP_VERIFY=true OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init && terraform ${TF_CMD}";\
        cd $(TF_EXAMPLE_FOLDER) && export OTF_INSECURE_SKIP_VERIFY=true OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init && terraform ${TF_CMD};\
    else\
        echo "[ERROR] Sanity check against swagger endpoint[$(OTF_VAR_SWAGGER_URL)] failed...Please make sure the service provider API is up and running and exposes swagger APIs on '$(OTF_VAR_SWAGGER_URL)'";\
    fi
endef

.PHONY: all build fmt vet lint test run_terraform
