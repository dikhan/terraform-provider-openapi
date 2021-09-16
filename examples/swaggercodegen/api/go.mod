module github.com/dikhan/terraform-provider-openapi/examples/swaggercodegen/api

go 1.14

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/dikhan/terraform-provider-openapi v0.31.1
	github.com/gorilla/mux v1.6.2
	github.com/pborman/uuid v0.0.0-20170612153648-e790cca94e6c
)

replace (
	github.com/dikhan/terraform-provider-openapi => ./
)