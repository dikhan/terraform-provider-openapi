package openapi

import "github.com/go-openapi/spec"

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetHeaderConfigurations(t *testing.T) {
	Convey("Given a list of parameters containing one required header parameter", t, func() {
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
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "", IsRequired: true})
			})
		})
	})
	Convey("Given a list of parameters containing one optional header parameter", t, func() {
		parameters := parameterGroups{
			[]spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "X-Request-ID",
						In:       "header",
						Required: false,
					},
				},
			},
		}
		Convey("When GetHeaderConfigurationsForParameterGroups method is called", func() {
			headerConfigProps := getHeaderConfigurationsForParameterGroups(parameters)
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "", IsRequired: false})
			})
		})
	})
	Convey("Given a list of parameters containing one required header parameter with the 'x-terraform-header' extension", t, func() {
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
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "x_request_id", IsRequired: true})
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
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "x-request-id", IsRequired: true})
			})
		})
	})
	Convey("Given a list of parameters containing multiple required header parameter", t, func() {
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
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "x_request_id", IsRequired: true})
			})
			Convey("And the header configs returned should also contain 'x_some_other_header'", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Some-Other-Header", TerraformName: "x_some_other_header", IsRequired: true})
			})
		})
	})
	Convey("Given a multiple list of parameters containing one required parameter", t, func() {
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
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "x_request_id", IsRequired: true})
			})
			Convey("Then the header configs returned should contain 'x_request_id_2'", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID2", TerraformName: "x_request_id2", IsRequired: true})
			})
		})
	})
}

func TestGetAllHeaderParameters(t *testing.T) {
	Convey("Given a swagger doc containing paths with header type parameters and different header names", t, func() {
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
													Name:     "X-Request-ID-cdn",
													In:       "header",
													Required: true,
												},
											},
										},
									},
								},
							},
						},
						"/v1/lbs": {
							PathItemProps: spec.PathItemProps{
								Post: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID-lb",
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
		Convey("When getPathHeaderParams method is called", func() {
			headerConfigProps := getAllHeaderParameters(spec.Paths.Paths)
			Convey("Then the header configs returned should contain all the different headers found", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-cdn", IsRequired: true})
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-lb", IsRequired: true})
			})
		})
	})
	Convey("Given a swagger doc containing paths with header type parameters and same header names", t, func() {
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
											},
										},
									},
								},
							},
						},
						"/v1/lbs": {
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
		Convey("When getPathHeaderParams method is called", func() {
			headerConfigProps := getAllHeaderParameters(spec.Paths.Paths)
			Convey("Then the headers shoud contain just one header since the other header names were the same", func() {
				So(len(headerConfigProps), ShouldEqual, 1)
			})
			Convey("Then the header configs returned ", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", IsRequired: true})
			})
		})
	})
}

func TestGetPathHeaderParams(t *testing.T) {
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
		Convey("When getPathHeaderParams method is called", func() {
			headerConfigProps := getPathHeaderParams(spec.Paths.Paths["/v1/cdns"])
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", TerraformName: "x_request_id", IsRequired: true})
			})
		})
	})

	Convey("Given a swagger doc containing a path that contains all CRUD operations with header type parameter and different name", t, func() {
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
		Convey("When getPathHeaderParams method is called", func() {
			headerConfigProps := getPathHeaderParams(spec.Paths.Paths["/v1/cdns"])
			Convey("Then the header configs returned should contain all header parameters specified in the path operation", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-post", IsRequired: true})
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-get", IsRequired: true})
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-put", IsRequired: true})
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID-delete", IsRequired: true})
			})
		})
	})

	Convey("Given a swagger doc containing a path that contains all CRUD operations with header type parameter and the same name", t, func() {
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
											},
										},
									},
								},
								Get: &spec.Operation{
									OperationProps: spec.OperationProps{
										Parameters: []spec.Parameter{
											{
												ParamProps: spec.ParamProps{
													Name:     "X-Request-ID",
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
													Name:     "X-Request-ID",
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
													Name:     "X-Request-ID",
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
		Convey("When getPathHeaderParams method is called", func() {
			headerConfigProps := getPathHeaderParams(spec.Paths.Paths["/v1/cdns"])
			Convey("Then the headers size should be 1", func() {
				So(len(headerConfigProps), ShouldEqual, 1)
			})
			Convey("Then the header configs returned should contain all header parameters specified in the path operation", func() {
				So(headerConfigProps, ShouldContain, SpecHeaderParam{Name: "X-Request-ID", IsRequired: true})
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
