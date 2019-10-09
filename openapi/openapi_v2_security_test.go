package openapi

import (
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetAPIKeySecurityDefinitions(t *testing.T) {
	Convey("Given a specV2Security loaded with a security definition of type header bearer auth", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{
				"apikey_auth": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "header",
						Type: "apiKey",
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfAuthenticationSchemeBearer: true,
						},
					},
				},
			},
		}
		Convey("When GetAPIKeySecurityDefinitions method is called", func() {
			securityDefinitions, err := specV2Security.GetAPIKeySecurityDefinitions()
			secDefs := *securityDefinitions
			Convey("Then the the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the security schemes match the expectations", func() {
				So(secDefs, ShouldNotBeEmpty)
			})
			Convey("And the security schemes should be of type header bearer", func() {
				So(secDefs[0], ShouldHaveSameTypeAs, specAPIKeyHeaderBearerSecurityDefinition{})
				So(secDefs[0].getAPIKey().Name, ShouldEqual, authorizationHeader)
				So(secDefs[0].buildValue("jwtToken"), ShouldEqual, "Bearer jwtToken")
			})
		})
	})
	Convey("Given a specV2Security loaded with a security definition of type header", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{
				"apikey_auth": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "header",
						Type: "apiKey",
						Name: "headerName",
					},
				},
			},
		}
		Convey("When GetAPIKeySecurityDefinitions method is called", func() {
			securityDefinitions, err := specV2Security.GetAPIKeySecurityDefinitions()
			secDefs := *securityDefinitions
			Convey("Then the the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the security schemes match the expectations", func() {
				So(secDefs, ShouldNotBeEmpty)
			})
			Convey("And the security schemes should be of type header bearer", func() {
				So(secDefs[0], ShouldHaveSameTypeAs, specAPIKeyHeaderSecurityDefinition{})
				So(secDefs[0].getAPIKey().Name, ShouldEqual, "headerName")
				So(secDefs[0].buildValue("someToken"), ShouldEqual, "someToken")
			})
		})
	})
	Convey("Given a specV2Security loaded with a security definition of type header bearer", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{
				"apikey_auth": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "header",
						Type: "apiKey",
					},
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfAuthenticationSchemeBearer: true,
						},
					},
				},
			},
		}
		Convey("When GetAPIKeySecurityDefinitions method is called", func() {
			securityDefinitions, err := specV2Security.GetAPIKeySecurityDefinitions()
			secDefs := *securityDefinitions
			Convey("Then the the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the security schemes match the expectations", func() {
				So(secDefs, ShouldNotBeEmpty)
			})
			Convey("And the security schemes should be of type header bearer", func() {
				So(secDefs[0], ShouldHaveSameTypeAs, specAPIKeyHeaderBearerSecurityDefinition{})
				So(secDefs[0].getAPIKey().Name, ShouldEqual, "Authorization")
				So(secDefs[0].buildValue("jwtToken"), ShouldEqual, "Bearer jwtToken")
			})
		})
	})
	Convey("Given a specV2Security loaded with a security definition of type query bearer", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{
				"apikey_auth": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "query",
						Type: "apiKey",
						Name: "queryParamName",
					},
				},
			},
		}
		Convey("When GetAPIKeySecurityDefinitions method is called", func() {
			securityDefinitions, err := specV2Security.GetAPIKeySecurityDefinitions()
			secDefs := *securityDefinitions
			Convey("Then the the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the security schemes match the expectations", func() {
				So(secDefs, ShouldNotBeEmpty)
			})
			Convey("And the security schemes should be of type query bearer", func() {
				So(secDefs[0], ShouldHaveSameTypeAs, specAPIKeyQuerySecurityDefinition{})
				So(secDefs[0].getAPIKey().Name, ShouldEqual, "queryParamName")
				So(secDefs[0].buildValue("someToken"), ShouldEqual, "someToken")
			})
		})
	})

	Convey("Given a specV2Security loaded with a header security definition that is missing the apiKeyName", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{
				"apikey_auth": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "header",
						Type: "apiKey",
					},
				},
			},
		}
		Convey("When GetAPIKeySecurityDefinitions method is called", func() {
			_, err := specV2Security.GetAPIKeySecurityDefinitions()
			Convey("Then the the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected one", func() {
				So(err.Error(), ShouldEqual, "specAPIKeyHeaderSecurityDefinition missing mandatory apiKey name")
			})
		})
	})

	Convey("Given a specV2Security loaded with a query security definition that is missing the apiKeyName", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{
				"apikey_auth": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "query",
						Type: "apiKey",
					},
				},
			},
		}
		Convey("When GetAPIKeySecurityDefinitions method is called", func() {
			_, err := specV2Security.GetAPIKeySecurityDefinitions()
			Convey("Then the the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected one", func() {
				So(err.Error(), ShouldEqual, "specAPIKeyQuerySecurityDefinition missing mandatory apiKey name")
			})
		})
	})
}

func TestGetGlobalSecuritySchemes(t *testing.T) {
	Convey("Given a specV2Security loaded with a global security scheme which is defined in the security definitions", t, func() {
		expectedSecuritySchemeName := "apikey_auth"
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{
				{
					expectedSecuritySchemeName: []string{},
				},
			},
			SecurityDefinitions: spec.SecurityDefinitions{
				expectedSecuritySchemeName: &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "header",
						Type: "apiKey",
						Name: authorizationHeader,
					},
				},
			},
		}
		Convey("When GetGlobalSecuritySchemes method is called", func() {
			specSecuritySchemes, err := specV2Security.GetGlobalSecuritySchemes()
			Convey("Then the the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the security schemes should not be empty", func() {
				So(specSecuritySchemes, ShouldNotBeEmpty)
			})
			Convey("And the security schemes should have the right security scheme name", func() {
				So(specSecuritySchemes[0].Name, ShouldEqual, expectedSecuritySchemeName)
			})
		})
	})
	Convey("Given a specV2Security loaded with a NON defined global security scheme", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity: []map[string][]string{
				{
					"nonExistingScheme": []string{},
				},
			},
			SecurityDefinitions: spec.SecurityDefinitions{
				"apikey_auth": &spec.SecurityScheme{
					SecuritySchemeProps: spec.SecuritySchemeProps{
						In:   "header",
						Type: "apiKey",
						Name: authorizationHeader,
					},
				},
			},
		}
		Convey("When GetGlobalSecuritySchemes method is called", func() {
			_, err := specV2Security.GetGlobalSecuritySchemes()
			Convey("Then the the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the security schemes should not be empty", func() {
				So(err.Error(), ShouldEqual, "global security scheme 'nonExistingScheme' not found or not matching supported 'apiKey' type")
			})
		})
	})
}

func TestIsBearerScheme(t *testing.T) {
	Convey("Given a specV2Security", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity:      []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{},
		}
		Convey("When isBearerScheme with a SecurityScheme that has the bearer extension with value true", func() {
			secDef := &spec.SecurityScheme{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfAuthenticationSchemeBearer: true,
					},
				},
			}
			isBearerAuth := specV2Security.isBearerScheme(secDef)
			Convey("The the value returned should be true", func() {
				So(isBearerAuth, ShouldBeTrue)
			})
		})
		Convey("When isBearerScheme with a SecurityScheme that has the bearer extension with value false", func() {
			secDef := &spec.SecurityScheme{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfAuthenticationSchemeBearer: false,
					},
				},
			}
			isBearerAuth := specV2Security.isBearerScheme(secDef)
			Convey("The the value returned should be false", func() {
				So(isBearerAuth, ShouldBeFalse)
			})
		})
		Convey("When isBearerScheme with a SecurityScheme that DOES not have the bearer extension", func() {
			secDef := &spec.SecurityScheme{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{},
				},
			}
			isBearerAuth := specV2Security.isBearerScheme(secDef)
			Convey("The the value returned should be false", func() {
				So(isBearerAuth, ShouldBeFalse)
			})
		})
	})
}

func TestIsRefreshTokenAuth(t *testing.T) {
	Convey("Given a specV2Security", t, func() {
		specV2Security := specV2Security{
			GlobalSecurity:      []map[string][]string{},
			SecurityDefinitions: spec.SecurityDefinitions{},
		}
		Convey("When isRefreshTokenAuth is called with a SecurityScheme that has the refresh token extension and a value specified", func() {
			secDef := &spec.SecurityScheme{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfAuthenticationRefreshToken: "refresh token server URL",
					},
				},
			}
			isRefreshAuth := specV2Security.isRefreshTokenAuth(secDef)
			Convey("Then the value returned should be as specified in SecuritySchema", func() {
				So(isRefreshAuth, ShouldEqual, "refresh token server URL")
			})
		})
		Convey("When isRefreshTokenAuth is called with a SecurityScheme that has the refresh token extension with an empty string value", func() {
			secDef := &spec.SecurityScheme{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfAuthenticationRefreshToken: "",
					},
				},
			}
			isRefreshAuth := specV2Security.isRefreshTokenAuth(secDef)
			Convey("Then the value returned should be an empty string", func() {
				So(isRefreshAuth, ShouldEqual, "")
			})
		})
		Convey("When isRefreshTokenAuth is called with a SecurityScheme that DOES not have the refresh token extension", func() {
			secDef := &spec.SecurityScheme{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{},
				},
			}
			isRefreshAuth := specV2Security.isRefreshTokenAuth(secDef)
			Convey("Then the value returned should be an empty string", func() {
				So(isRefreshAuth, ShouldEqual, "")
			})
		})
	})
}
