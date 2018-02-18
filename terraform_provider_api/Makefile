PROVIDER_NAME?="sp"
OTF_VAR_SWAGGER_URL?="https://localhost:8443/swagger.yaml"
TF_CMD?="plan"

TEST_PACKAGES?=$$(go list ./... | grep -v vendor/)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v 'vendor')

default: build

all: test build

build: fmt
	@go build -o terraform-provider
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

run_terraform: build
	@echo "[INFO] Performing sanity check against the service provider's swagger endpoint '$(OTF_VAR_SWAGGER_URL)'"
	@$(eval SWAGGER_HTTP_STATUS := $(shell curl -s -o /dev/null -w '%{http_code}' $(OTF_VAR_SWAGGER_URL) -k))

	@if [ "$(SWAGGER_HTTP_STATUS)" = 200 ]; then\
        echo "[INFO] Service provider swagger end point '$(OTF_VAR_SWAGGER_URL)' is reachable";\
        echo "[INFO] Executing TF command: OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init && terraform ${TF_CMD}";\
        export OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && terraform init  && terraform ${TF_CMD};\
    else\
        echo "[ERROR] Sanity check against swagger endpoint[$(OTF_VAR_SWAGGER_URL)] failed...Please make sure the service provider API is up and running and exposes swagger APIs on '$(OTF_VAR_SWAGGER_URL)'";\
    fi

.PHONY: all build fmt vet lint test run_terraform
