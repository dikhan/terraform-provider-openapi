package openapiutils

import (
	"github.com/go-openapi/spec"
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
			headerConfigProps := GetHeaderConfigurationsForParameterGroups(parameters)
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
			headerConfigProps := GetHeaderConfigurationsForParameterGroups(parameters)
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
			headerConfigProps := GetHeaderConfigurationsForParameterGroups(parameters)
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
			headerConfigProps := GetHeaderConfigurationsForParameterGroups(parameterGroups)
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id")
			})
			Convey("Then the header configs returned should contain 'x_request_id2'", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id2")
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
			headerConfigProps := GetHeaderConfigurationsForParameterGroups(parameters)
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
			headerConfigProps := GetAllHeaderParameters(spec)
			Convey("Then the header configs returned should contain 'x_request_id'", func() {
				So(headerConfigProps, ShouldContainKey, "x_request_id")
			})
		})
	})
}
