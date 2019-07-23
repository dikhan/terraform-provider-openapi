package openapi

import (
	"fmt"
	"testing"

	"github.com/go-openapi/spec"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewOpenAPIBackendConfigurationV2(t *testing.T) {
	Convey("Given a swagger spec 2.0 and an openAPIDocumentURL", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
		}
		openAPIDocumentURL := "www.domain.com"
		Convey("When newOpenAPIBackendConfigurationV2 method is called", func() {
			specV2BackendConfiguration, err := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
			Convey("Then the error returned should be  nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the providerClient should comply with SpecBackendConfiguration interface", func() {
				var _ SpecBackendConfiguration = specV2BackendConfiguration
			})
		})
	})

	Convey("Given a swagger spec that is not supported 3.0 and an openAPIDocumentURL", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "3.0",
			},
		}
		openAPIDocumentURL := "www.domain.com"
		Convey("When newOpenAPIBackendConfigurationV2 method is called", func() {
			_, err := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
			Convey("Then the error returned should be NOT nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldEqual, "swagger version '3.0' not supported, specV2BackendConfiguration only supports 2.0")
			})
		})
	})

	Convey("Given a swagger spec 2.0 and an empty openAPIDocumentURL", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
		}
		openAPIDocumentURL := ""
		Convey("When newOpenAPIBackendConfigurationV2 method is called", func() {
			_, err := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
			Convey("Then the error returned should be NOT nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldEqual, "missing mandatory parameter openAPIDocumentURL")
			})
		})
	})
}

func TestGetHost(t *testing.T) {
	Convey("Given a specV2BackendConfiguration with the host configured", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
				Host:    "www.some-backend.com",
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getHost method is called", func() {
			host, err := specV2BackendConfiguration.getHost()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the host should be correct", func() {
				So(host, ShouldEqual, "www.some-backend.com")
			})
		})
	})

	Convey("Given a specV2BackendConfiguration with the host not configured", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
				Host:    "",
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getHost method is called", func() {
			host, err := specV2BackendConfiguration.getHost()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the host should be the one where the swagger file is being served", func() {
				So(host, ShouldEqual, openAPIDocumentURL)
			})
		})
	})

	Convey("Given a specV2BackendConfiguration with the host not configured", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
				Host:    "",
			},
		}
		specV2BackendConfiguration := specV2BackendConfiguration{spec: spec, openAPIDocumentURL: ""}
		Convey("When getHost method is called", func() {
			_, err := specV2BackendConfiguration.getHost()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldEqual, "could not find valid host from URL provided: ''")
			})
		})
	})
}

func TestGetHostByRegion(t *testing.T) {
	Convey("Given a specV2BackendConfiguration with a multi-region configuration", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					extTfProviderRegions:         "rst1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getHostByRegion method is called with an existing region", func() {
			host, err := specV2BackendConfiguration.getHostByRegion("rst1")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the host should be correct", func() {
				So(host, ShouldEqual, "www.rst1.some-backend.com")
			})
		})
		Convey("When getHostByRegion method is called with a NON existing region", func() {
			_, err := specV2BackendConfiguration.getHostByRegion("nonExisting")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "region nonExisting not matching allowed ones [rst1]")
			})
		})
		Convey("When getHostByRegion method is called with an empty region", func() {
			_, err := specV2BackendConfiguration.getHostByRegion("")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "can't get host by region, missing region value")
			})
		})
	})

	Convey("Given a specV2BackendConfiguration with a NON multi-region configuration", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
				Host:    "www.some-backend.com",
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getHostByRegion method is called with a NON existing region", func() {
			_, err := specV2BackendConfiguration.getHostByRegion("nonExisting")
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "missing 'x-terraform-provider-multiregion-fqdn' extension or value provided not matching multiregion host format")
			})
		})
	})
}

func TestValidateRegion(t *testing.T) {
	Convey("Given a region and a list of allowed regions", t, func() {
		region := "allowed"
		allowedRegions := []string{region}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(&spec.Swagger{SwaggerProps: spec.SwaggerProps{Swagger: "2.0"}}, openAPIDocumentURL)
		Convey("When validateRegion method is called with a region that is allowed", func() {
			err := specV2BackendConfiguration.validateRegion(region, allowedRegions)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("When validateRegion method is called with a non allowed region", func() {
			err := specV2BackendConfiguration.validateRegion("nonAllowed", allowedRegions)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "region nonAllowed not matching allowed ones [allowed]")
			})
		})
	})
}

func TestGetDefaultRegion(t *testing.T) {
	Convey("Given a specV2BackendConfiguration", t, func() {
		spec := &spec.Swagger{SwaggerProps: spec.SwaggerProps{Swagger: "2.0"}}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getDefaultRegion() method is called with an array of regions", func() {
			regions := []string{"rst1", "dub1"}
			region, err := specV2BackendConfiguration.getDefaultRegion(regions)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then region should match the default one, which is the first one found in the 'x-terraform-provider-regions' extension value", func() {
				So(region, ShouldEqual, "rst1")
			})
		})
		Convey("When getDefaultRegion() method is called with an empty array", func() {
			regions := []string{}
			_, err := specV2BackendConfiguration.getDefaultRegion(regions)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "empty regions provided")
			})
		})
		Convey("When getDefaultRegion() method is called with a nil array", func() {
			_, err := specV2BackendConfiguration.getDefaultRegion(nil)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "empty regions provided")
			})
		})
	})
}

func TestIsMultiRegion(t *testing.T) {
	Convey("Given a specV2BackendConfiguration that is multi-region (contains x-terraform-provider-multiregion-fqdn with parameterised host) and x-terraform-provider-regions extension with regions", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					extTfProviderRegions:         "rst1, dub1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isMultiRegion() method is called", func() {
			isMultiRegion, host, regions, err := specV2BackendConfiguration.isMultiRegion()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then it should be multi region", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("Then host should be the parameterised host", func() {
				So(host, ShouldEqual, "www.${region}.some-backend.com")
			})
			Convey("Then regions should contain the right regions", func() {
				So(regions, ShouldContain, "rst1")
				So(regions, ShouldContain, "dub1")
			})
		})
	})
	Convey("Given a specV2BackendConfiguration that is NOT multi-region (missing multi-region configuration)", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			// Missing required configuration that makes provider multi-region
			//VendorExtensible: spec.VendorExtensible{
			//	Extensions: spec.Extensions{
			//		extTfProviderMultiRegionFQDN: "www.some-backend.com",
			//		extTfProviderRegions:         "rst1",
			//	},
			//},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isMultiRegion() method is called", func() {
			isMultiRegion, _, _, err := specV2BackendConfiguration.isMultiRegion()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then it should NOT be multi region", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
		})
	})
	Convey("Given a specV2BackendConfiguration that is multi-region but the x-terraform-provider-multiregion-fqdn does not have a parameterised value", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.some-backend.com",
					extTfProviderRegions:         "rst1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isMultiRegion() method is called", func() {
			isMultiRegion, _, _, err := specV2BackendConfiguration.isMultiRegion()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "'x-terraform-provider-multiregion-fqdn' extension value provided not matching multiregion host format")
			})
			Convey("Then it should NOT be multi region", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
		})
	})

	Convey("Given a specV2BackendConfiguration that has a multiregion host but the x-terraform-provider-regions does not have any regions", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					extTfProviderRegions:         "",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isMultiRegion() method is called", func() {
			isMultiRegion, _, _, err := specV2BackendConfiguration.isMultiRegion()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "mandatory multiregion 'x-terraform-provider-regions' extension empty value provided")
			})
			Convey("Then it should NOT be multi region", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
		})
	})
}

func TestGetProviderRegions(t *testing.T) {
	Convey("Given a specV2BackendConfiguration that has the x-terraform-provider-regions populated with comma separated string values", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					extTfProviderRegions:         "rst1, dub1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getProviderRegions() method is called", func() {
			regions, err := specV2BackendConfiguration.getProviderRegions()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then regions should contain the right regions", func() {
				So(regions, ShouldContain, "rst1")
				So(regions, ShouldContain, "dub1")
			})
		})
	})
	Convey("Given a specV2BackendConfiguration that has the x-terraform-provider-regions but empty values", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					extTfProviderRegions:         "",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getProviderRegions() method is called", func() {
			_, err := specV2BackendConfiguration.getProviderRegions()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "mandatory multiregion 'x-terraform-provider-regions' extension empty value provided")
			})
		})
	})
	Convey("Given a specV2BackendConfiguration that has the x-terraform-provider-regions does not exists", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					//extTfProviderRegions:         "rst1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getProviderRegions() method is called", func() {
			_, err := specV2BackendConfiguration.getProviderRegions()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "mandatory multiregion 'x-terraform-provider-regions' extension missing")
			})
		})
	})
}

func TestIsHostMultiRegion(t *testing.T) {
	Convey("Given a specV2BackendConfiguration that is multi-region (with 'x-terraform-provider-regions' extension values being comma separated with spaces)", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					extTfProviderRegions:         "rst1, dub1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isHostMultiRegion() method is called", func() {
			isMultiRegion, host, err := specV2BackendConfiguration.isHostMultiRegion()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then it should be multi region", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("Then host should be the parameterised host", func() {
				So(host, ShouldEqual, "www.${region}.some-backend.com")
			})
		})
	})
	Convey("Given a specV2BackendConfiguration that is multi-region (with 'x-terraform-provider-regions' extension values being comma separated with NO spaces)", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.${region}.some-backend.com",
					extTfProviderRegions:         "rst1,dub1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isHostMultiRegion() method is called", func() {
			isMultiRegion, host, err := specV2BackendConfiguration.isHostMultiRegion()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then it should be multi region", func() {
				So(isMultiRegion, ShouldBeTrue)
			})
			Convey("Then host should be the parameterised host", func() {
				So(host, ShouldEqual, "www.${region}.some-backend.com")
			})
		})
	})
	Convey("Given a specV2BackendConfiguration that is multi-region but the x-terraform-provider-multiregion-fqdn does not have a parameterised value", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
			VendorExtensible: spec.VendorExtensible{
				Extensions: spec.Extensions{
					extTfProviderMultiRegionFQDN: "www.some-backend.com",
					extTfProviderRegions:         "rst1",
				},
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isHostMultiRegion() method is called", func() {
			isMultiRegion, _, err := specV2BackendConfiguration.isHostMultiRegion()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error should match the expected message", func() {
				So(err.Error(), ShouldEqual, "'x-terraform-provider-multiregion-fqdn' extension value provided not matching multiregion host format")
			})
			Convey("Then it should be multi region", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
		})
	})
	Convey("Given a specV2BackendConfiguration that is not multiregion", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger: "2.0",
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When isHostMultiRegion() method is called", func() {
			isMultiRegion, _, err := specV2BackendConfiguration.isHostMultiRegion()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the result shoudl be false", func() {
				So(isMultiRegion, ShouldBeFalse)
			})
		})
	})
}

func TestGetBasePath(t *testing.T) {
	Convey("Given a specV2BackendConfiguration with the basePath configured", t, func() {
		spec := &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger:  "2.0",
				Host:     "www.some-backend.com",
				BasePath: "/api",
			},
		}
		openAPIDocumentURL := "www.domain.com"
		specV2BackendConfiguration, _ := newOpenAPIBackendConfigurationV2(spec, openAPIDocumentURL)
		Convey("When getBasePath method is called", func() {
			basePath := specV2BackendConfiguration.getBasePath()
			Convey("And the host should be correct", func() {
				So(basePath, ShouldEqual, "/api")
			})
		})
	})
}

func TestGetHTTPSchemes(t *testing.T) {
	testCases := []struct {
		name           string
		inputSchemes   []string
		expectedScheme string
		expectedError  string
	}{
		{name: "both http and https schemes are configured", inputSchemes: []string{"http", "https"}, expectedScheme: "https"},
		{name: "mix of schemes configured including supported ones without https", inputSchemes: []string{"http", "ws"}, expectedScheme: "http"},
		{name: "mix of schemes configured including supported ones with https", inputSchemes: []string{"http", "ws", "https"}, expectedScheme: "https"},
		{name: "none http or https schemes are configured", inputSchemes: []string{}, expectedError: "no schemes specified - must use http or https"},
		{name: "none of the schemes configured are supported", inputSchemes: []string{"ws"}, expectedError: "specified schemes [ws] are not supported - must use http or https"},
	}
	for _, tc := range testCases {
		Convey(fmt.Sprintf("Given a specV2BackendConfiguration with %s", tc.name), t, func() {
			spec := &spec.Swagger{
				SwaggerProps: spec.SwaggerProps{
					Swagger: "2.0",
					Schemes: tc.inputSchemes,
				},
			}
			specV2BackendConfiguration, err := newOpenAPIBackendConfigurationV2(spec, "www.domain.com")
			So(err, ShouldBeNil)
			Convey("When getHTTPSchemes method is called", func() {
				httpScheme, err := specV2BackendConfiguration.getHTTPScheme()
				Convey("Then the returned http scheme and  error should be as expected", func() {
					So(err == nil && tc.expectedError == "" || err.Error() == tc.expectedError, ShouldBeTrue)
					So(httpScheme, ShouldEqual, tc.expectedScheme)
				})
			})
		})

	}
}
