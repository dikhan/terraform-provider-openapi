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
			Convey("And the security schemes should be of type query bearer", func() {
				So(secDefs[0], ShouldHaveSameTypeAs, specAPIKeyQueryBearerSecurityDefinition{})
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
