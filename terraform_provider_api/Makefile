PROVIDER_NAME?="sp"
OTF_VAR_SWAGGER_URL?="https://localhost:8443/swagger.yaml"
TF_CMD?="plan"

TEST_PACKAGES?=$$(go list ./... | grep -v vendor/)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v 'vendor')

default: build

all: test build

build: fmt
	@go build -o terraform-provider-$(PROVIDER_NAME)

fmt:
	@echo "Running gofmt on the current directory"
	gofmt -w $(GOFMT_FILES)

vet:
	@echo "Running go vet on the current directory"
	@go vet $(TEST_PACKAGES) ; if [ $$? -eq 1 ]; then \
		echo "Vet found suspicious constructs. Please fix the reported constructs before submitting code for review"; \
		exit 1; \
	fi

lint:
	@echo "Running golint on the current directory"
	@go get -u github.com/golang/lint/golint
	@golint -set_exit_status $(TEST_PACKAGES)

test: fmt vet lint
	@echo "Testing terraform-provider-openapi"
	go test -v -cover $(TEST_PACKAGES) || exit 1

run_terraform: build
	@echo "Executing terraform ${TF_CMD}"
	@export OTF_VAR_$(PROVIDER_NAME)_SWAGGER_URL=$(OTF_VAR_SWAGGER_URL) && \
	terraform init  && \
	terraform ${TF_CMD}

.PHONY: all build fmt vet lint test run_terraform
