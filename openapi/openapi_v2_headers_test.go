package openapi

import "github.com/go-openapi/spec"

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetHeaderConfigurations(t *testing.T) {
	Convey("Given a list of parameters containing one header parameter with the 'x-terraform-header' extension", t, func() {
		parameters := parameterGroups{
			[]spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Request-ID",
						In:       "header",
						Required: true,
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							"x-terraform-header": "x_request_id",
						},
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getHeaderConfigurationsForParameterGroups(parameters)
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id")
			})
		})
	})
	Convey("Given a list of parameters containing one header parameter with the 'x-terraform-header' extension but value is not terraform field compliant", t, func() {
		parameters := parameterGroups{
			[]spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Request-ID",
						In:       "header",
						Required: true,
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							"x-terraform-header": "x-request-id",
						},
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getHeaderConfigurationsForParameterGroups(parameters)
			Convey("Then the header configs returned should contain 'x_request_id' as a terraform field name translation (converting dashes to underscores) should have been performed", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id")
			})
		})
	})
	Convey("Given a list of parameters containing multiple header parameter", t, func() {
		parameters := parameterGroups{
			[]spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Request-ID",
						In:       "header",
						Required: true,
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							"x-terraform-header": "x_request_id",
						},
					},
				},
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Some-Other-Header",
						In:       "header",
						Required: true,
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							"x-terraform-header": "x_some_other_header",
						},
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getHeaderConfigurationsForParameterGroups(parameters)
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id")
			})
			Convey("And the header configs returned should also contain 'x_some_other_header'", func() {
				So(headerConfigProps, ShouldContainKey, "x_some_other_header")
			})
		})
	})
	Convey("Given a multiple list of parameters containing one parameter", t, func() {
		parameterGroups := parameterGroups{
			[]spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Request-ID",
						In:       "header",
						Required: true,
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							"x-terraform-header": "x_request_id",
						},
					},
				},
			},
			[]spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Request-ID2",
						In:       "header",
						Required: true,
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							"x-terraform-header": "x_request_id2",
						},
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getHeaderConfigurationsForParameterGroups(parameterGroups)
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id")
			})
			Convey("Then the header configs returned should contain 'x_request_id_2'", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id_2")
			})
		})
	})
	Convey("Given a list of parameters containing one parameter that does not contain the extension 'x-terraform-header'", t, func() {
		parameters := parameterGroups{
			[]spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Request-ID",
						In:       "header",
						Required: true,
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getHeaderConfigurationsForParameterGroups(parameters)
			Convey("Then the header configs returned should contain 'x_request_id' due to the automatic conversion from header name 'X-Request-ID' to a terraform compliant field name", func() {
				// This prevent terraform from throwing the following error: * X-Request-ID: Field name may only contain lowercase alphanumeric characters & underscores.
				So(headerConfigProps, ShouldContainKey, "x_request_id")
			})
		})
	})
}

func TestGetAllHeaderParameters(t *testing.T) {
	Convey("Given a swagger doc containing a path that contains one POST operation with a header parameter", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/v1/cdns": {
							PathItemProps: spec.PathItemProps{
								Post: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID",
													In:       "header",
													Required: true,
												},
												VendorExtensible: spec.VendorExtensible{
													Extensions: spec.Extensions{
														"x-terraform-header": "x_request_id",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getPathHeaderParams(spec.Paths.Paths["/v1/cdns"])
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "x_request_id"})
			})
		})
	})

	Convey("Given a swagger doc containing a path that contains all CRUD operations with a header parameter", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/v1/cdns": {
							PathItemProps: spec.PathItemProps{
								Post: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID-post",
													In:       "header",
													Required: true,
												},
											},
										},
									},
								},
								Get: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID-get",
													In:       "header",
													Required: true,
												},
											},
										},
									},
								},
								Put: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID-put",
													In:       "header",
													Required: true,
												},
											},
										},
									},
								},
								Delete: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID-delete",
													In:       "header",
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getPathHeaderParams(spec.Paths.Paths["/v1/cdns"])
			Convey("Then the header configs returned should contain all header parameters specified in the path operation", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-post"})
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-get"})
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-put"})
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-delete"})
			})
		})
	})
}

func TestAppendOperationParametersIfPresent(t *testing.T) {
	Convey("Given a swagger doc containing a path that contains one POST operation with a header parameter", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/v1/cdns": {
							PathItemProps: spec.PathItemProps{
								Post: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID",
													In:       "header",
													Required: true,
												},
												VendorExtensible: spec.VendorExtensible{
													Extensions: spec.Extensions{
														"x-terraform-header": "x_request_id",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		Convey("When appendOperationParametersIfPresent method is called", func() {
			parametersGroup := parameterGroups{}
			headerConfigProps := appendOperationParametersIfPresent(parametersGroup, spec.Paths.Paths["/v1/cdns"].Post)
			Convey("Then the header config groups should not be empty", func() {
				So(headerConfigProps, ShouldNotBeEmpty)
			})
			Convey("And the header group should contain", func() {
				So(headerConfigProps[0], ShouldContain, spec.Paths.Paths["/v1/cdns"].Post.Parameters[0])
			})
		})
	})

	Convey("Given a swagger doc containing a path that non of the operations have parameters", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Paths: &spec.Paths{
					Paths: map[string]spec.PathItem{
						"/v1/cdns": {
							PathItemProps: spec.PathItemProps{
								Post: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{},
									},
								},
							},
						},
					},
				},
			},
		}
		Convey("When appendOperationParametersIfPresent method is called", func() {
			parametersGroup := parameterGroups{}
			headerConfigProps := appendOperationParametersIfPresent(parametersGroup, spec.Paths.Paths["/v1/cdns"].Post)
			Convey("And the header group should contain", func() {
				So(headerConfigProps[0], ShouldBeEmpty)
			})
		})
	})
}
