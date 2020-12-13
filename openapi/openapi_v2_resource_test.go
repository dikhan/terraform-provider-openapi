package openapi

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-openapi/jsonreference"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
)

func TestNewSpecV2Resource(t *testing.T) {
	Convey("Given a root path /users, a root path item, schema definitions", t, func() {
		inputPath := "/users"
		inputRootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		inputSchemaDefinitions := map[string]spec.Schema{}
		Convey("When the newSpecV2Resource method is called", func() {
			r, err := newSpecV2Resource(inputPath, spec.Schema{}, inputRootPathItem, spec.PathItem{}, inputSchemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the resource returned should have name `users` and there should be no error", func() {
				So(r.GetResourceName(), ShouldEqual, "users")
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestNewSpecV2ResourceWithConfig(t *testing.T) {
	Convey("Given a root path /users/ containing a trailing slash and a root path item item", t, func() {
		path := "/users/"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(r.GetResourceName(), ShouldEqual, "users")
			})
		})
	})
	Convey("Given a root path /users with no trailing slash and a root path item", t, func() {
		path := "/users"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(r.GetResourceName(), ShouldEqual, "users")
			})
		})
	})
	Convey("Given a root path /users, a root path item and a region", t, func() {
		path := "/users"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		region := "rst1"
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig(region, path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the value returned should be the resource name plus the region appended at the end
				So(r.GetResourceName(), ShouldEqual, "users_rst1")
			})
		})
	})
	Convey("Given a root path that is versioned such as '/v1/users/' and a root path item item", t, func() {
		path := "/v1/users/"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
				So(r.GetResourceName(), ShouldEqual, "users_v1")
			})
		})
	})
	Convey("Given a root path that is versioned with number higher than 9 such as '/v12/users/' and a root path item item", t, func() {
		path := "/v12/users/"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(r.GetResourceName(), ShouldEqual, "users_v12")
			})
		})
	})
	Convey("Given a root path such as '/v1/something/users' and a root path item", t, func() {
		path := "/v1/something/users"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(r.GetResourceName(), ShouldEqual, "users")
			})
		})
	})
	Convey("Given a root path which has path parameters '/api/v1/nodes/{name}/proxy' and a root path item", t, func() {
		path := "/api/v1/nodes/{name}/proxy"
		paths := map[string]spec.PathItem{
			"/api/v1/nodes": {
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{},
				},
			},
		}
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, paths)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(r.GetResourceName(), ShouldEqual, "nodes_v1_proxy")
			})
		})
	})
	Convey("Given a root path '/users' and the create operation has the extension 'x-terraform-resource-name' and a root path item item", t, func() {
		path := "/v1/users"
		expectedResourceName := "user"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfResourceName: expectedResourceName,
						},
					},
				},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", path, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			resourceName := r.GetResourceName()
			expectedTerraformName := fmt.Sprintf("%s_v1", expectedResourceName)
			Convey(fmt.Sprintf("And the value returned should still be '%s'", expectedTerraformName), func() {
				So(err, ShouldBeNil)
				So(resourceName, ShouldEqual, expectedTerraformName)
			})
		})
	})
	Convey("Given an invalid root path", t, func() {
		invalidRootPath := "/api/v1/nodes/{name}"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			_, err := newSpecV2ResourceWithConfig("", invalidRootPath, spec.Schema{}, rootPathItem, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("And the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
	Convey("Given an empty path", t, func() {
		path := ""
		Convey("When newSpecV2Resource method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2Resource(path, spec.Schema{}, spec.PathItem{}, spec.PathItem{}, schemaDefinitions, map[string]spec.PathItem{})
			Convey("And the err returned output should match the expectation", func() {
				So(err.Error(), ShouldEqual, "path must not be empty")
				So(r, ShouldBeNil)
			})
		})
	})
	Convey("Given paths is nil", t, func() {
		var paths map[string]spec.PathItem
		Convey("When newSpecV2ResourceWithConfig method is called", func() {
			schemaDefinitions := map[string]spec.Schema{}
			r, err := newSpecV2ResourceWithConfig("", "/v1/users", spec.Schema{}, spec.PathItem{}, spec.PathItem{}, schemaDefinitions, paths)
			Convey("And the err returned output should match the expectation", func() {
				So(err.Error(), ShouldEqual, "paths must not be nil")
				So(r, ShouldBeNil)
			})
		})
	})
}

func TestNewSpecV2ResourceWithRegion(t *testing.T) {
	Convey("Given a path, schemaDefinition, rootPathItem, instancePathItem, paths, schemaDefinitions AND a region that is empty", t, func() {
		path := "/v1/users"
		schemaDefinition := spec.Schema{}
		rootPathItem := spec.PathItem{}
		instancePathItem := spec.PathItem{}
		paths := map[string]spec.PathItem{}
		schemaDefinitions := map[string]spec.Schema{}
		region := ""
		Convey("When newSpecV2ResourceWithRegion method is called", func() {
			r, err := newSpecV2ResourceWithRegion(region, path, schemaDefinition, rootPathItem, instancePathItem, schemaDefinitions, paths)
			Convey("Then the result returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "region must not be empty")
				So(r, ShouldBeNil)
			})
		})
	})
	Convey("Given a path, schemaDefinition, rootPathItem, instancePathItem, paths, schemaDefinitions AND a region that is NOT empty", t, func() {
		path := "/users"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		schemaDefinition := spec.Schema{}
		instancePathItem := spec.PathItem{}
		paths := map[string]spec.PathItem{}
		schemaDefinitions := map[string]spec.Schema{}
		region := "rst1"
		Convey("When newSpecV2ResourceWithRegion method is called", func() {
			r, err := newSpecV2ResourceWithRegion(region, path, schemaDefinition, rootPathItem, instancePathItem, schemaDefinitions, paths)
			Convey("Then the value returned should be the resource name plus the region appended at the end", func() {
				So(err, ShouldBeNil)
				So(r.GetResourceName(), ShouldEqual, "users_rst1")
			})
		})
	})
}

func TestShouldIgnoreResource(t *testing.T) {
	Convey("Given a SpecV2Resource configured with a root path item that does not contain the post operation defined", t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: nil,
				},
			},
		}
		Convey("When ShouldIgnoreResource is called", func() {
			shouldIgnoreResource := r.ShouldIgnoreResource()
			Convey("Then the result should be false", func() {
				So(shouldIgnoreResource, ShouldBeFalse)
			})
		})
	})
	Convey(fmt.Sprintf("Given a SpecV2Resource configured with a root path item that does not contain the %s extension", extTfExcludeResource), t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{},
				},
			},
		}
		Convey("When ShouldIgnoreResource is called", func() {
			shouldIgnoreResource := r.ShouldIgnoreResource()
			Convey("Then the result should be false", func() {
				So(shouldIgnoreResource, ShouldBeFalse)
			})
		})
	})
	Convey(fmt.Sprintf("Given a SpecV2Resource configured with a root path item that DOES contain the %s extension with value equal true", extTfExcludeResource), t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfExcludeResource: true,
							},
						},
					},
				},
			},
		}
		Convey("When ShouldIgnoreResource is called", func() {
			shouldIgnoreResource := r.ShouldIgnoreResource()
			Convey("Then the result should be true", func() {
				So(shouldIgnoreResource, ShouldBeTrue)
			})
		})
	})
	Convey(fmt.Sprintf("Given a SpecV2Resource configured with a root path item that DOES contain the %s extension with value equal false", extTfExcludeResource), t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfExcludeResource: false,
							},
						},
					},
				},
			},
		}
		Convey("When ShouldIgnoreResource is called", func() {
			shouldIgnoreResource := r.ShouldIgnoreResource()
			Convey("Then the result should be false", func() {
				So(shouldIgnoreResource, ShouldBeFalse)
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a root path item where the extensions are nil", t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: nil,
						},
					},
				},
			},
		}
		Convey("When ShouldIgnoreResource is called", func() {
			shouldIgnoreResource := r.ShouldIgnoreResource()
			Convey("Then the result should be false", func() {
				So(shouldIgnoreResource, ShouldBeFalse)
			})
		})
	})
}

func TestBuildResourceName(t *testing.T) {

	testCases := []struct {
		name                 string
		path                 string
		paths                map[string]spec.PathItem
		rootPathItem         spec.PathItem
		expectedResourceName string
		expectedError        error
	}{
		{
			name:  "resource name built from path itself",
			path:  "/v1/cdns",
			paths: nil,
			rootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{},
				},
			},
			expectedResourceName: "cdns_v1",
			expectedError:        nil,
		},
		{
			name:  "preferred resource name in root level post operation",
			path:  "/v1/cdns",
			paths: nil,
			rootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceName: "cdn",
							},
						},
					},
				},
			},
			expectedResourceName: "cdn_v1",
			expectedError:        nil,
		},
		{
			name:  "preferred resource name in root level path",
			path:  "/v1/cdns",
			paths: nil,
			rootPathItem: spec.PathItem{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfResourceName: "cdn",
					},
				},
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{},
				},
			},
			expectedResourceName: "cdn_v1",
			expectedError:        nil,
		},
		{
			name: "first level sub-resource with no preferred parent names",
			path: "/cdns/{id}/firewalls",
			paths: map[string]spec.PathItem{
				"/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
			expectedResourceName: "cdns_firewalls",
			expectedError:        nil,
		},
		{
			name: "first level sub-resource with preferred parent names",
			path: "/cdns/{id}/firewalls",
			paths: map[string]spec.PathItem{
				"/cdns": {
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfResourceName: "cdn",
						},
					},
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
			expectedResourceName: "cdn_firewalls",
			expectedError:        nil,
		},
		{
			name: "two level sub-resource with version and only one parent using preferred name",
			path: "/v1/cdns/{id}/v2/firewalls/{id}/v3/rules",
			paths: map[string]spec.PathItem{
				"/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/v1/cdns/{id}/v2/firewalls": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfResourceName: "firewall",
								},
							},
						},
					},
				},
			},
			expectedResourceName: "cdns_v1_firewall_v2_rules_v3",
			expectedError:        nil,
		},
		{
			name: "two level sub-resource with preferred parent names",
			path: "/v1/cdns/{id}/v2/firewalls/{id}/v3/rules",
			paths: map[string]spec.PathItem{
				"/v1/cdns": {
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfResourceName: "cdn",
						},
					},
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/v1/cdns/{id}/v2/firewalls": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfResourceName: "firewall",
								},
							},
						},
					},
				},
			},
			expectedResourceName: "cdn_v1_firewall_v2_rules_v3",
			expectedError:        nil,
		},
		{
			name:  "",
			path:  "?",
			paths: nil,
			rootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{},
				},
			},
			expectedResourceName: "",
			expectedError:        errors.New("could not find a valid name for resource instance path '?'"),
		},
	}

	for _, tc := range testCases {
		r := SpecV2Resource{
			Path:         tc.path,
			Paths:        tc.paths,
			RootPathItem: tc.rootPathItem,
		}
		resourceName, err := r.buildResourceName()
		if tc.expectedError != nil {
			assert.Error(t, err, tc.expectedError.Error(), tc.name)
		} else {
			assert.NoError(t, err, tc.name)
			assert.Equal(t, tc.expectedResourceName, resourceName, tc.name)
		}
	}
}

func TestBuildResourceNameFromPath(t *testing.T) {

	testCases := []struct {
		name                 string
		path                 string
		expectedResourceName string
		preferredName        string
		expectedError        error
	}{
		{
			name:                 "basic resource - no version, no preferred name",
			path:                 "/cdns",
			preferredName:        "",
			expectedResourceName: "cdns",
			expectedError:        nil,
		},
		{
			name:                 "resource with hyphen",
			path:                 "/cdns-test",
			preferredName:        "",
			expectedResourceName: "cdns_test",
			expectedError:        nil,
		},
		{
			name:                 "resource with version",
			path:                 "/v1/cdns",
			preferredName:        "",
			expectedResourceName: "cdns_v1",
			expectedError:        nil,
		},
		{
			name:                 "resource with version and preferred name",
			path:                 "/v1/cdns",
			preferredName:        "cdn",
			expectedResourceName: "cdn_v1",
			expectedError:        nil,
		},
		{
			name:                 "resource with double digit version number",
			path:                 "/v11/cdns",
			preferredName:        "",
			expectedResourceName: "cdns_v11",
			expectedError:        nil,
		},
		{
			name:                 "resource using semver",
			path:                 "/v1.1.1/cdns",
			preferredName:        "",
			expectedResourceName: "cdns", // semver in paths is not supported at the moment, this documents the resource output for such use case
			expectedError:        nil,
		},
		{
			name:                 "resource with number and letter in version",
			path:                 "/v1a/cdns",
			preferredName:        "",
			expectedResourceName: "cdns",
			expectedError:        nil,
		},
		{
			name:                 "resource with version and no preferred name",
			path:                 "/v1/cdns/",
			preferredName:        "",
			expectedResourceName: "cdns_v1",
			expectedError:        nil,
		},
		{
			name:                 "basic resource with parent - no version or preferred name",
			path:                 "/cdns/{id}/firewalls",
			preferredName:        "",
			expectedResourceName: "firewalls",
			expectedError:        nil,
		},
		{
			name:                 "resource with parent and version on parent",
			path:                 "/v1/cdns/{id}/firewalls",
			preferredName:        "",
			expectedResourceName: "firewalls",
			expectedError:        nil,
		},
		{
			name:                 "resource with parent and version on child",
			path:                 "/cdns/{id}/v1/firewalls",
			preferredName:        "",
			expectedResourceName: "firewalls_v1",
			expectedError:        nil,
		},
		{
			name:                 "resource with parent and versions on both parent and child",
			path:                 "/v1/cdns/{id}/v2/firewalls",
			preferredName:        "",
			expectedResourceName: "firewalls_v2",
			expectedError:        nil,
		},
		{
			name:                 "resource with two parents - version on both parents and child",
			path:                 "/v1/cdns/{id}/v2/firewalls/{id}/v3/rules",
			preferredName:        "",
			expectedResourceName: "rules_v3",
			expectedError:        nil,
		},
		{
			name:                 "resource with two parents - version on both parents but not on child",
			path:                 "/v1/cdns/{id}/v2/firewalls/{id}/rules",
			preferredName:        "",
			expectedResourceName: "rules",
			expectedError:        nil,
		},
		{ // This is considered a wrongly structured path not following resful best practises for building subresource paths, however the plugin still supports it to not be so opinionated
			name:                 "resource with two parents - version on one parent and child",
			path:                 "/v1/cdns/{id}/firewalls/v3/rules",
			preferredName:        "",
			expectedResourceName: "rules_v3",
			expectedError:        nil,
		},
		{
			name:                 "path with no resource name",
			path:                 "/",
			preferredName:        "",
			expectedResourceName: "",
			expectedError:        nil,
		},
		{
			name:          "empty path",
			path:          "",
			preferredName: "",
			expectedError: errors.New("could not find a valid name for resource instance path ''"),
		},
		{
			name:          "badly formed path",
			path:          "&^",
			preferredName: "",
			expectedError: errors.New("could not find a valid name for resource instance path '&^'"),
		},
		{
			name:                 "path ending in backslash and starting with 'api'",
			path:                 "/api/v1/group/",
			preferredName:        "iamgroup",
			expectedResourceName: "iamgroup_v1",
			expectedError:        nil,
		},
	}

	for _, tc := range testCases {
		Convey("Given a SpecV2Resource", t, func() {
			r := SpecV2Resource{}
			Convey(fmt.Sprintf("When buildResourceName is called with the given path and preferred name: %s", tc.name), func() {
				resourceName, err := r.buildResourceNameFromPath(tc.path, tc.preferredName)
				if tc.expectedError != nil {
					Convey("Then the error returned should be the expected one", func() {
						So(err.Error(), ShouldEqual, tc.expectedError.Error())
						So(resourceName, ShouldBeEmpty)
					})
				} else {
					Convey("Then the resource name should be the expected one", func() {
						So(err, ShouldBeNil)
						So(resourceName, ShouldEqual, tc.expectedResourceName)
					})
				}
			})
		})
	}
}

func TestParentResourceInfo(t *testing.T) {
	Convey("Given a SpecV2Resource configured with a root path", t, func() {
		r := SpecV2Resource{
			Path:  "/cdns",
			Paths: map[string]spec.PathItem{},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the ParentResourceInfo struct returned should be nil", func() {
				So(parentResourceInfo, ShouldBeNil)
			})
		})
	})
	Convey("Given a SpecV2Resource configured with a root path using versioning", t, func() {
		r := SpecV2Resource{
			Path:  "/v1/cdns",
			Paths: map[string]spec.PathItem{},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the ParentResourceInfo struct returned should be nil", func() {
				So(parentResourceInfo, ShouldBeNil)
			})
		})
	})
	Convey("Given a SpecV2Resource configured with a path that is indeed a sub-resource (with parent using versioning)", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/firewalls",
			Paths: map[string]spec.PathItem{
				"/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				So(parentResourceInfo, ShouldPointTo, r.parentResourceInfoCached) // checking cache is populated
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 1)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns_v1")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns_v1")
				// the parentURIs contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/v1/cdns")
				// the parentInstanceURIs contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/v1/cdns/{id}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a path that is a sub-resource and the paths configured having trailing forward slashes and having a preferred name", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/firewalls",
			Paths: map[string]spec.PathItem{
				"/v1/cdns/": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfResourceName: "cdn",
								},
							},
						},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 1)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdn_v1")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdn_v1")
				// the parentURIs contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/v1/cdns")
				// the parentInstanceURIs contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/v1/cdns/{id}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a base path that is indeed a sub-resource", t, func() {
		r := SpecV2Resource{
			Path: "/api/v1/nodes/{name}/proxy",
			Paths: map[string]spec.PathItem{
				"/api/v1/nodes": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 1)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "nodes_v1")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "nodes_v1")
				// the parentURIs contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/api/v1/nodes")
				// the parentInstanceURIs contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/api/v1/nodes/{name}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a base path that is indeed a sub-resource with multiple levels", t, func() {
		r := SpecV2Resource{
			Path: "/api/v1/cdns/{id}/v2/firewalls/{id}/v3/rules",
			Paths: map[string]spec.PathItem{
				"/api/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/api/v1/cdns/{id}/v2/firewalls": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 2)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns_v1")
				So(parentResourceInfo.parentResourceNames[1], ShouldEqual, "firewalls_v2")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns_v1_firewalls_v2")
				// the parentURIs contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/api/v1/cdns")
				So(parentResourceInfo.parentURIs[1], ShouldEqual, "/api/v1/cdns/{id}/v2/firewalls")
				// the parentInstanceURIs contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/api/v1/cdns/{id}")
				So(parentResourceInfo.parentInstanceURIs[1], ShouldEqual, "/api/v1/cdns/{id}/v2/firewalls/{id}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a base path and the 2 level parent starts with some base path too and it's not versioned", t, func() {
		r := SpecV2Resource{
			Path: "/api/v1/cdns/{id}/something/firewalls/{id}/v3/rules",
			Paths: map[string]spec.PathItem{
				"/api/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/api/v1/cdns/{id}/something/firewalls": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 2)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns_v1")
				So(parentResourceInfo.parentResourceNames[1], ShouldEqual, "firewalls")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns_v1_firewalls")
				// the parentURIs contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/api/v1/cdns")
				So(parentResourceInfo.parentURIs[1], ShouldEqual, "/api/v1/cdns/{id}/something/firewalls")
				// the parentInstanceURIs contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/api/v1/cdns/{id}")
				So(parentResourceInfo.parentInstanceURIs[1], ShouldEqual, "/api/v1/cdns/{id}/something/firewalls/{id}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a path that is indeed a sub-resource (no versioning)", t, func() {
		r := SpecV2Resource{
			Path: "/cdns/{id}/firewalls",
			Paths: map[string]spec.PathItem{
				"/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 1)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns")
				// the parentURIs contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/cdns")
				// the parentInstanceURIs contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/cdns/{id}")
			})
		})
	})
	Convey("Given a SpecV2Resource configured with a path that is indeed a sub-resource (both using versioning)", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/v2/firewalls",
			Paths: map[string]spec.PathItem{
				"/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 1)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns_v1")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns_v1")
				// the parentURIs contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/v1/cdns")
				// the parentInstanceURIs contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/v1/cdns/{id}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a path that is indeed a multiple level sub-resource", t, func() {
		r := SpecV2Resource{
			Path: "/cdns/{id}/firewalls/{id}/rules",
			Paths: map[string]spec.PathItem{
				"/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/cdns/{id}/firewalls": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				// the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 2)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns")
				So(parentResourceInfo.parentResourceNames[1], ShouldEqual, "firewalls")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns_firewalls")
				// the parentURIs should contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/cdns")
				So(parentResourceInfo.parentURIs[1], ShouldEqual, "/cdns/{id}/firewalls")
				// the parentInstanceURIs should contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/cdns/{id}")
				So(parentResourceInfo.parentInstanceURIs[1], ShouldEqual, "/cdns/{id}/firewalls/{id}")
			})
		})
	})
	Convey("Given a SpecV2Resource configured with a path that is indeed a multiple level sub-resource with versioning", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/v2/firewalls/{id}/v3/rules",
			Paths: map[string]spec.PathItem{
				"/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
				"/v1/cdns/{id}/v2/firewalls": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the result returned should be the expected one", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				//the parentResourceNames should not be empty and contain the right items
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 2)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns_v1")
				So(parentResourceInfo.parentResourceNames[1], ShouldEqual, "firewalls_v2")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns_v1_firewalls_v2")
				// the parentURIs should contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/v1/cdns")
				So(parentResourceInfo.parentURIs[1], ShouldEqual, "/v1/cdns/{id}/v2/firewalls")
				// the parentInstanceURIs should contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/v1/cdns/{id}")
				So(parentResourceInfo.parentInstanceURIs[1], ShouldEqual, "/v1/cdns/{id}/v2/firewalls/{id}")
			})
		})
	})
	Convey("Given a SpecV2Resource configured with a path that is a subresource but the path is wrongly structured not following best restful practises for building subresource paths (the 'firewalls' parent in the path is missing the id path param)", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/v2/firewalls/v3/rules",
			Paths: map[string]spec.PathItem{
				"/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the resource should be considered a subresource and the output should match the expected output values", func() {
				So(parentResourceInfo, ShouldNotBeNil)
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 1)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdns_v1")
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdns_v1")
				// the parentURIs should contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/v1/cdns")
				// the parentInstanceURIs should contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 1)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/v1/cdns/{id}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a path that is a subresource and the parents have preferred names on the post operation", t, func() {
		expectedCDNResourceName := "cdn"
		expectedFirewallResourceName := "firewall"
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/v2/firewalls/{id}/v3/rules",
			Paths: map[string]spec.PathItem{
				"/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfResourceName: expectedCDNResourceName,
								},
							},
						},
					},
				},
				"/v1/cdns/{id}/v2/firewalls": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfResourceName: expectedFirewallResourceName,
								},
							},
						},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the parentResourceNames should not be empty and contain the right items", func() {
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 2)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdn_v1")
				So(parentResourceInfo.parentResourceNames[1], ShouldEqual, "firewall_v2")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdn_v1_firewall_v2")
				// the parentURIs should contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/v1/cdns")
				So(parentResourceInfo.parentURIs[1], ShouldEqual, "/v1/cdns/{id}/v2/firewalls")
				// the parentInstanceURIs should contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/v1/cdns/{id}")
				So(parentResourceInfo.parentInstanceURIs[1], ShouldEqual, "/v1/cdns/{id}/v2/firewalls/{id}")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a path that is a subresource and the parents have preferred names on the root level path", t, func() {
		expectedCDNResourceName := "cdn"
		expectedFirewallResourceName := "firewall"
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/v2/firewalls/{id}/v3/rules",
			Paths: map[string]spec.PathItem{
				"/v1/cdns": {
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfResourceName: expectedCDNResourceName,
						},
					},
				},
				"/v1/cdns/{id}/v2/firewalls": {
					VendorExtensible: spec.VendorExtensible{
						Extensions: spec.Extensions{
							extTfResourceName: expectedFirewallResourceName,
						},
					},
				},
			},
		}
		Convey("When ParentResourceInfo is called", func() {
			parentResourceInfo := r.GetParentResourceInfo()
			Convey("Then the parentResourceNames should not be empty and contain the right items", func() {
				So(len(parentResourceInfo.parentResourceNames), ShouldEqual, 2)
				So(parentResourceInfo.parentResourceNames[0], ShouldEqual, "cdn_v1")
				So(parentResourceInfo.parentResourceNames[1], ShouldEqual, "firewall_v2")
				// the fullParentResourceName should match the expected name
				So(parentResourceInfo.fullParentResourceName, ShouldEqual, "cdn_v1_firewall_v2")
				// the parentURIs should contain the expected parent URIs
				So(len(parentResourceInfo.parentURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentURIs[0], ShouldEqual, "/v1/cdns")
				So(parentResourceInfo.parentURIs[1], ShouldEqual, "/v1/cdns/{id}/v2/firewalls")
				// the parentInstanceURIs should contain the expected instances URIs
				So(len(parentResourceInfo.parentInstanceURIs), ShouldEqual, 2)
				So(parentResourceInfo.parentInstanceURIs[0], ShouldEqual, "/v1/cdns/{id}")
				So(parentResourceInfo.parentInstanceURIs[1], ShouldEqual, "/v1/cdns/{id}/v2/firewalls/{id}")
			})
		})
	})
	Convey("Given a SpecV2Resource with parentResourceInfoCached populated", t, func() {
		r := SpecV2Resource{parentResourceInfoCached: &ParentResourceInfo{}}
		Convey("When GetParentResourceInfo is called", func() {
			p := r.GetParentResourceInfo()
			Convey("Then the returned ParentResourceInfo should point to parentResourceInfoCached", func() {
				So(p, ShouldPointTo, r.parentResourceInfoCached)
			})
		})
	})
}

func assertSchemaProperty(actualSpecSchemaDefinition *SpecSchemaDefinition, expectedName string, expectedType schemaDefinitionPropertyType, expectedRequired, expectedReadOnly, expectedComputed bool) {
	prop, err := actualSpecSchemaDefinition.getProperty(expectedName)
	So(err, ShouldBeNil)
	fmt.Printf(">>> Validating '%s' property settings\n", prop.Name)
	So(prop.Type, ShouldEqual, expectedType)
	So(prop.Required, ShouldEqual, expectedRequired)
	So(prop.ReadOnly, ShouldEqual, expectedReadOnly)
	So(prop.Computed, ShouldEqual, expectedComputed)
}

func assertSchemaParentProperty(actualSpecSchemaDefinition *SpecSchemaDefinition, expectedName string) {
	assertSchemaProperty(actualSpecSchemaDefinition, expectedName, TypeString, true, false, false)
}

func TestGetResourceSchema(t *testing.T) {
	Convey("Given a SpecV2Resource containing a root path and various properties in the schema", t, func() {
		r := &SpecV2Resource{
			Path: "/cdns",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
						"int_optional_computed_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"integer"},
							},
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfComputed: true,
								},
							},
						},
						"number_required_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"number"},
							},
						},
						"bool_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"boolean"},
							},
						},
					},
				},
			},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("And the SpecSchemaDefinition returned should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(len(specSchemaDefinition.Properties), ShouldEqual, len(r.SchemaDefinition.SchemaProps.Properties))
				So(specSchemaDefinition, ShouldPointTo, r.specSchemaDefinitionCached) // checking cache is populated
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				assertSchemaProperty(specSchemaDefinition, "int_optional_computed_prop", TypeInt, false, false, true)
				assertSchemaProperty(specSchemaDefinition, "number_required_prop", TypeFloat, true, false, false)
				assertSchemaProperty(specSchemaDefinition, "bool_prop", TypeBool, false, false, false)
			})
		})
	})

	Convey("Given a SpecV2Resource containing a sub-resource path (one level) that has a weird sub-resource path (firewalls is missing firewalls/{id}) and a schema definition with a property", t, func() {
		r := &SpecV2Resource{
			Path: "/v1/cdns/{id}/firewalls/v1/rules",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should be configured with the expected number of properties including the parent id one
				So(len(specSchemaDefinition.Properties), ShouldEqual, 2)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				// the SpecSchemaDefinition returned should be configured with the parent id property marked as IsParentProperty, with the right name, type and being required
				assertSchemaParentProperty(specSchemaDefinition, "cdns_v1_id")
			})
		})
	})

	Convey("Given a SpecV2Resource containing a sub-resource path (one level and parent versioned) and a schema containing one property", t, func() {
		r := &SpecV2Resource{
			Path: "/v1/cdns/{id}/firewalls",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should be configured with the expected number of properties including the parent id one
				So(len(specSchemaDefinition.Properties), ShouldEqual, 2)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				// the SpecSchemaDefinition returned should be configured with the parent id property with the expected configuration
				assertSchemaParentProperty(specSchemaDefinition, "cdns_v1_id")
			})
		})
	})

	Convey("Given a SpecV2Resource that is a sub-resource (one level parent with no version)", t, func() {
		r := &SpecV2Resource{
			Path: "/parent/{parent_id}/child",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When getSchemaDefinition is called with a schema containing various properties", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should contain the expected number of properties (including the parent one)
				So(len(specSchemaDefinition.Properties), ShouldEqual, 2)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				// the SpecSchemaDefinition returned should also include the parent property
				assertSchemaParentProperty(specSchemaDefinition, "parent_id")
			})
		})
	})

	Convey("Given a SpecV2Resource that is a sub-resource (two level parent) and a schema for child", t, func() {
		r := &SpecV2Resource{
			Path: "/parent/{parent_id}/subparent/{subparent_id}/child",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should contain the expected number of properties (including the parent ones)
				So(len(specSchemaDefinition.Properties), ShouldEqual, 3)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				// the SpecSchemaDefinition returned should also include the parents properties
				assertSchemaParentProperty(specSchemaDefinition, "parent_id")
				assertSchemaParentProperty(specSchemaDefinition, "subparent_id")
			})
		})
	})

	Convey("Given a SpecV2Resource containing a sub-resource path (one level) and the parent resource using a preferred resource name", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{id}/firewalls",
			Paths: map[string]spec.PathItem{
				"/v1/cdns": {
					PathItemProps: spec.PathItemProps{
						Post: &spec.Operation{
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfResourceName: "cdn",
								},
							},
						},
					},
				},
			},
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should be configured with the expected number of properties including the parent id one
				So(len(specSchemaDefinition.Properties), ShouldEqual, 2)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				// the SpecSchemaDefinition returned should be configured with the parent id property named using the preferred parent name configured in the parent resource
				assertSchemaParentProperty(specSchemaDefinition, "cdn_v1_id")
			})
		})
	})

	Convey("Given a SpecV2Resource containing a subresource path (two level and parents versioned)", t, func() {
		r := &SpecV2Resource{
			Path:  "/v1/cdns/{cdn_id}/v2/firewalls/{fw_id}/rules",
			Paths: nil,
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
						"int_optional_computed_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"integer"},
							},
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfComputed: true,
								},
							},
						},
						"number_required_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"number"},
							},
						},
						"bool_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"boolean"},
							},
						},
					},
				},
			},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should be configured with the expected number of properties including the parent id ones
				So(len(specSchemaDefinition.Properties), ShouldEqual, 6)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				assertSchemaProperty(specSchemaDefinition, "int_optional_computed_prop", TypeInt, false, false, true)
				assertSchemaProperty(specSchemaDefinition, "number_required_prop", TypeFloat, true, false, false)
				assertSchemaProperty(specSchemaDefinition, "bool_prop", TypeBool, false, false, false)
				// the SpecSchemaDefinition returned should be configured with the parent id property too
				assertSchemaParentProperty(specSchemaDefinition, "cdns_v1_id")
				assertSchemaParentProperty(specSchemaDefinition, "firewalls_v2_id")
			})
		})
	})

	Convey("Given a SpecV2Resource configured with a miss configured schema (eg: schema contains a property that is missing the type", t, func() {
		r := &SpecV2Resource{
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"bad_property": {
							SchemaProps: spec.SchemaProps{
								// Type: Missing the type
							},
						},
					},
				},
			},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the schema definition returned should nil and the error should be the expected one", func() {
				So(err.Error(), ShouldEqual, "failed to process property 'bad_property': non supported '[]' type")
				So(specSchemaDefinition, ShouldBeNil)
			})
		})
	})

	Convey("Given a SpecV2Resource containing a cached specSchemaDefinitionCached", t, func() {
		r := &SpecV2Resource{
			specSchemaDefinitionCached: &SpecSchemaDefinition{},
		}
		Convey("When GetResourceSchema is called", func() {
			specSchemaDefinition, err := r.GetResourceSchema()
			Convey("Then the schema definition returned should be the cached one and the error should be nil", func() {
				So(err, ShouldBeNil)
				So(specSchemaDefinition, ShouldPointTo, r.specSchemaDefinitionCached)
			})
		})
	})
}

func TestGetSchemaDefinition(t *testing.T) {
	Convey("Given a SpecV2Resource containing a root path", t, func() {
		r := &SpecV2Resource{
			Path: "/cdns",
		}
		Convey("When getSchemaDefinition is called with a schema containing various properties", func() {
			s := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
						"int_optional_computed_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"integer"},
							},
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{
									extTfComputed: true,
								},
							},
						},
						"number_required_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"number"},
							},
						},
						"bool_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"boolean"},
							},
						},
					},
				},
			}
			specSchemaDefinition, err := r.getSchemaDefinition(s)
			Convey("Then the SpecSchemaDefinition returned should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(len(specSchemaDefinition.Properties), ShouldEqual, len(s.SchemaProps.Properties))
				assertSchemaProperty(specSchemaDefinition, "string_readonly_prop", TypeString, false, true, true)
				assertSchemaProperty(specSchemaDefinition, "int_optional_computed_prop", TypeInt, false, false, true)
				assertSchemaProperty(specSchemaDefinition, "number_required_prop", TypeFloat, true, false, false)
				assertSchemaProperty(specSchemaDefinition, "bool_prop", TypeBool, false, false, false)
			})
		})
	})

	Convey("Given a SpecV2Resource containing a sub-resource", t, func() {
		r := &SpecV2Resource{
			Path: "/zone/{zone_id}/recordset",
		}
		Convey("When getSchemaDefinition is called with a schema containing various properties", func() {
			s := &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"id": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
					},
				},
			}
			specSchemaDefinition, err := r.getSchemaDefinition(s)
			Convey("And the SpecSchemaDefinition returned should be 1 and no parent id should be in the schema. Only GetResourceSchema() allows parent ids to be injects", func() {
				So(err, ShouldBeNil)
				So(len(specSchemaDefinition.Properties), ShouldEqual, 1)
				assertSchemaProperty(specSchemaDefinition, "id", TypeString, false, true, true)
			})
		})
	})
}

func TestGetSchemaDefinitionWithOptions(t *testing.T) {
	Convey("Given a blank SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When getSchemaDefinitionWithOptions is called with a nil arg", func() {
			_, err := r.getSchemaDefinitionWithOptions(nil, true)
			Convey("Then the error returned matches the expected one", func() {
				So(err.Error(), ShouldEqual, "schema argument must not be nil")
			})
		})
		Convey("When getSchemaDefinitionWithOptions is called passing a blank schema", func() {
			d, e := r.getSchemaDefinitionWithOptions(&spec.Schema{}, true)
			Convey("Then the schema definition contains empty Properties", func() {
				So(e, ShouldBeNil)
				So(d, ShouldNotBeNil)
				So(d.Properties, ShouldBeEmpty)
			})
		})
		Convey("When getSchemaDefinitionWithOptions is called passing a schema with a weird property type", func() {
			schema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Properties: map[string]spec.Schema{
						"string_readonly_prop": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"something weird"},
							},
						},
					},
				},
			}
			d, e := r.getSchemaDefinitionWithOptions(&schema, true)
			Convey("And the schema definition returned is nil", func() {
				So(e, ShouldNotBeNil)
				So(d, ShouldBeNil)
			})
		})
	})

	Convey("Given a SpecV2Resource containing a sub-resource path (one level) with a schema containing a property that matches the parent property id", t, func() {
		r := &SpecV2Resource{
			Path: "/zone/{zone_id}/recordset",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"id": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
						"zone_id": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
						},
					},
				},
			},
		}
		Convey("When getSchemaDefinitionWithOptions is called with the addParentProps set to true", func() {
			specSchemaDefinition, err := r.getSchemaDefinitionWithOptions(&r.SchemaDefinition, true)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should be configured with the id and the zone_id properties (and no extra parent property 'zone_id' will be added since it's already there)
				So(len(specSchemaDefinition.Properties), ShouldEqual, 2)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "id", TypeString, false, true, true)
				// the SpecSchemaDefinition returned should be configured with the parent id property with the expected configuration
				// Note due to the model already containing a parent id property (zone_id) it will be reconfigured to be required. This ensures the resource tf configuration requires the parent id property to be populated.
				assertSchemaParentProperty(specSchemaDefinition, "zone_id")
			})
		})
	})

	Convey("Given a SpecV2Resource containing a sub-resource path with a schema containing an array of objects", t, func() {
		r := &SpecV2Resource{
			Path: "/zone/{zone_id}/recordset",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"number_required_prop"},
					Properties: map[string]spec.Schema{
						"id": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
							SwaggerSchemaProps: spec.SwaggerSchemaProps{
								ReadOnly: true,
							},
						},
						"record": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"array"},
								Items: &spec.SchemaOrArray{
									Schema: &spec.Schema{
										SchemaProps: spec.SchemaProps{
											Type: spec.StringOrArray{"object"},
											Properties: map[string]spec.Schema{
												"content": {
													SchemaProps: spec.SchemaProps{
														Type: spec.StringOrArray{"string"},
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
		Convey("When getSchemaDefinitionWithOptions is called with addParentProps set to true", func() {
			specSchemaDefinition, err := r.getSchemaDefinitionWithOptions(&r.SchemaDefinition, true)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				// the SpecSchemaDefinition returned should be configured with the expected number of properties including the parent id one
				So(len(specSchemaDefinition.Properties), ShouldEqual, 3)
				// the SpecSchemaDefinition returned should be configured as expected
				assertSchemaProperty(specSchemaDefinition, "id", TypeString, false, true, true)
				assertSchemaProperty(specSchemaDefinition, "record", TypeList, false, false, false)
				// the SpecSchemaDefinition for the array property should not contain any parent id
				recordProp, _ := specSchemaDefinition.getProperty("record")
				So(len(recordProp.SpecSchemaDefinition.Properties), ShouldEqual, 1)
				So(recordProp.SpecSchemaDefinition.Properties[0].Name, ShouldEqual, "content")
				// the SpecSchemaDefinition returned should be configured with the parent id property with the expected configuration
				assertSchemaParentProperty(specSchemaDefinition, "zone_id")
			})
		})
	})
}

func TestGetResourcePath(t *testing.T) {
	Convey("Given a SpecV2Resource with path resource that is not parameterised (root resource)", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns",
		}
		Convey("When getResourcePath is called with an empty list of IDs", func() {
			resourcePath, err := r.getResourcePath([]string{})
			Convey("Then the returned resource path should match the expected one", func() {
				So(err, ShouldBeNil)
				So(resourcePath, ShouldEqual, "/v1/cdns")
				So(r.resolvedPathCached, ShouldEqual, "/v1/cdns")
			})
		})
		Convey("When getResourcePath is called with a nil list of IDs", func() {
			resourcePath, err := r.getResourcePath(nil)
			Convey("Then the returned resource path should match the expected one", func() {
				So(err, ShouldBeNil)
				So(resourcePath, ShouldEqual, "/v1/cdns")
				So(r.resolvedPathCached, ShouldEqual, "/v1/cdns")
			})
		})
	})

	Convey("Given a SpecV2Resource with path resource that is parameterised (one level sub-resource)", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{cdn_id}/v1/firewalls",
		}
		Convey("When getResourcePath is called with a list of IDs", func() {
			ids := []string{"parentID"}
			resourcePath, err := r.getResourcePath(ids)
			Convey("Then the returned resource path should match the expected one", func() {
				So(err, ShouldBeNil)
				So(resourcePath, ShouldEqual, "/v1/cdns/parentID/v1/firewalls")
				So(r.resolvedPathCached, ShouldEqual, "/v1/cdns/parentID/v1/firewalls")
			})
		})
		Convey("When getResourcePath is called with an empty list of IDs", func() {
			resourcePath, err := r.getResourcePath([]string{})
			Convey("Then the error returned should not be nil", func() {
				So(resourcePath, ShouldBeEmpty)
				So(r.resolvedPathCached, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, "could not resolve sub-resource path correctly '/v1/cdns/{cdn_id}/v1/firewalls' with the given ids - missing ids to resolve the path params properly: []")
			})
		})
		Convey("When getResourcePath is called with an nil list of IDs", func() {
			resourcePath, err := r.getResourcePath(nil)
			Convey("Then the error returned should not be nil", func() {
				So(resourcePath, ShouldBeEmpty)
				So(r.resolvedPathCached, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, "could not resolve sub-resource path correctly '/v1/cdns/{cdn_id}/v1/firewalls' with the given ids - missing ids to resolve the path params properly: []")
			})
		})
		Convey("When getResourcePath is called with a list of IDs that is bigger than the parameterised params in the path", func() {
			resourcePath, err := r.getResourcePath([]string{"cdnID", "somethingThatDoesNotBelongHere"})
			Convey("Then the error returned should not be nil", func() {
				So(resourcePath, ShouldBeEmpty)
				So(r.resolvedPathCached, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, "could not resolve sub-resource path correctly '/v1/cdns/{cdn_id}/v1/firewalls' with the given ids - more ids than path params: [cdnID somethingThatDoesNotBelongHere]")
			})
		})
		Convey("When getResourcePath is called with a list of IDs twhere some IDs contain forward slashes", func() {
			resourcePath, err := r.getResourcePath([]string{"cdnID/somethingElse"})
			Convey("Then the error returned should not be nil", func() {
				So(resourcePath, ShouldBeEmpty)
				So(r.resolvedPathCached, ShouldBeEmpty)
				So(err.Error(), ShouldEqual, "could not resolve sub-resource path correctly '/v1/cdns/{cdn_id}/v1/firewalls' due to parent IDs ([cdnID/somethingElse]) containing not supported characters (forward slashes)")
			})
		})
	})

	Convey("Given a SpecV2Resource with path resource that is parameterised (few levels sub-resource)", t, func() {
		r := SpecV2Resource{
			Path: "/v1/cdns/{cdn_id}/v1/firewalls/{fw_id}/rules",
		}
		Convey("When getResourcePath is called with a list of IDs", func() {
			ids := []string{"cdnID", "fwID"}
			resourcePath, err := r.getResourcePath(ids)
			Convey("And the returned resource path should match the expected one", func() {
				So(err, ShouldBeNil)
				So(resourcePath, ShouldEqual, "/v1/cdns/cdnID/v1/firewalls/fwID/rules")
				So(r.resolvedPathCached, ShouldEqual, "/v1/cdns/cdnID/v1/firewalls/fwID/rules")
			})
		})
	})

	Convey("Given a SpecV2Resource with resolvedPathCached populated", t, func() {
		r := SpecV2Resource{
			resolvedPathCached: "/v1/cdns",
		}
		Convey("When getResourcePath is called with a nil list of IDs", func() {
			resourcePath, err := r.getResourcePath(nil)
			Convey("Then the returned resource path should match the expected one", func() {
				So(err, ShouldBeNil)
				So(resourcePath, ShouldEqual, "/v1/cdns")
			})
		})
	})
}

func TestCreateSchemaDefinitionProperty(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}

		//////////////////
		// Type checks
		//////////////////

		Convey("When createSchemaDefinitionProperty is called with a propertyName and a propertySchema of type string that is not required", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeString)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})
		Convey("When createSchemaDefinitionProperty is called with a propertyName, propertySchema of type integer that is not required", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"integer"},
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeInt)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})
		Convey("When createSchemaDefinitionProperty is called with a propertyName, propertySchema of type number that is not required", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"number"},
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeFloat)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})
		Convey("When createSchemaDefinitionProperty is called with a propertyName, propertySchema of type boolean that is not required", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"boolean"},
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeBool)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a propertyName, propertySchema of unknown type", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					//Type: spec.StringOrArray{"boolean"}, NO TYPE ASSIGNED
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be the expected one and the schemaDefinitionProperty should be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to process property 'propertyName': non supported '[]' type")
				So(schemaDefinitionProperty, ShouldBeNil)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a propertyName, propertySchema of type object with nested properties that is not required", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					Properties: map[string]spec.Schema{
						"objectProperty": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
						},
					},
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeObject)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a propertyName, propertySchema of type object with NO nested properties nor a REF", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					// Missing object schema information
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be the expected one and the schemaDefinitionProperty should be nil", func() {
				So(err.Error(), ShouldEqual, "failed to process property 'propertyName': object is missing the nested schema definition or the ref is pointing to a non existing schema definition")
				So(schemaDefinitionProperty, ShouldBeNil)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a propertyName and non required propertySchema of type array with items of type string", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type:        spec.StringOrArray{"string"},
								Description: "items description",
							},
						},
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeList)
				So(schemaDefinitionProperty.ArrayItemsType, ShouldEqual, TypeString)
				So(schemaDefinitionProperty.SpecSchemaDefinition, ShouldBeNil)
				So(schemaDefinitionProperty.Description, ShouldEqual, "items description")
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a propertyName and propertySchema non required of type array with items of type object (nested)", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Properties: map[string]spec.Schema{
									"prop1": {
										SchemaProps: spec.SchemaProps{
											Type: spec.StringOrArray{"string"},
										},
									},
									"prop2": {
										SchemaProps: spec.SchemaProps{
											Type: spec.StringOrArray{"integer"},
										},
									},
								},
							},
						},
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeList)
				So(schemaDefinitionProperty.ArrayItemsType, ShouldEqual, TypeObject)
				So(schemaDefinitionProperty.SpecSchemaDefinition, ShouldNotBeNil)
				exists, _ := assertPropertyExists(schemaDefinitionProperty.SpecSchemaDefinition.Properties, "prop1")
				So(exists, ShouldBeTrue)
				exists, _ = assertPropertyExists(schemaDefinitionProperty.SpecSchemaDefinition.Properties, "prop2")
				So(exists, ShouldBeTrue)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a propertyName and propertySchema non required of type array with items of type object (external ref definition)", func() {
			r := SpecV2Resource{
				SchemaDefinitions: map[string]spec.Schema{
					"Listeners": {
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"object"},
							Properties: map[string]spec.Schema{
								"protocol": {
									SchemaProps: spec.SchemaProps{
										Type: spec.StringOrArray{"string"},
									},
								},
							},
						},
					},
				},
			}
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/Listeners")},
							},
						},
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Name, ShouldEqual, propertyName)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeList)
				So(schemaDefinitionProperty.ArrayItemsType, ShouldEqual, TypeObject)
				So(schemaDefinitionProperty.SpecSchemaDefinition, ShouldNotBeNil)
				So(schemaDefinitionProperty.SpecSchemaDefinition.Properties, ShouldNotBeEmpty)
				So(schemaDefinitionProperty.SpecSchemaDefinition.Properties[0].Name, ShouldEqual, "protocol")
				So(schemaDefinitionProperty.SpecSchemaDefinition.Properties[0].Type, ShouldEqual, TypeString)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that is required", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			requiredProperties := []string{"propertyName"}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Required, ShouldBeTrue)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that is required AND the property is also set as readOnly (nonesense)", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			requiredProperties := []string{"propertyName"}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be the expected one and the schemaDefinitionProperty should be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to process property 'propertyName': a required property cannot be readOnly too")
				So(schemaDefinitionProperty, ShouldBeNil)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with an optional property schema that is computed (readOnly)", func() {
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeTrue)
				So(schemaDefinitionProperty.isComputed(), ShouldBeTrue)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with an optional property schema that is computed (readOnly) AND has a default value (meaning the computed value is known at runtime)", func() {
			expectedDefaultValue := "someDefaultValue"
			propertyName := "propertyName"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:    spec.StringOrArray{"string"},
					Default: expectedDefaultValue,
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
			}
			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeTrue)
				So(schemaDefinitionProperty.isComputed(), ShouldBeTrue)
				So(schemaDefinitionProperty.Default, ShouldEqual, expectedDefaultValue)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with an optional property schema", func() {
			propertyName := "propertyWithNestedObj"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					Properties: map[string]spec.Schema{
						"nested_obj": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Properties: map[string]spec.Schema{
									"nested_prop": {
										SchemaProps: spec.SchemaProps{
											Type: spec.StringOrArray{"string"},
										},
									},
								},
							},
						},
					},
				}}

			requiredProperties := []string{}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty(propertyName, propertySchema, requiredProperties)
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Type, ShouldEqual, TypeObject)
				So(len(schemaDefinitionProperty.SpecSchemaDefinition.Properties), ShouldEqual, 1)
				So(schemaDefinitionProperty.SpecSchemaDefinition.Properties[0].Type, ShouldEqual, TypeObject)
				nestedSpecSchema := *(schemaDefinitionProperty.SpecSchemaDefinition.Properties)[0]
				So(nestedSpecSchema.SpecSchemaDefinition.Properties[0].Type, ShouldEqual, TypeString)

			})
		})

		Convey("When createSchemaDefinitionProperty is called with an optional property schema that has a default value (this means the property is optional-computed, since the API is expected to honour the default value (known at runtime) if input is not provided by the client)", func() {
			expectedDefaultValue := "someDefaultValue"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:    spec.StringOrArray{"string"},
					Default: expectedDefaultValue,
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
				So(schemaDefinitionProperty.Default, ShouldEqual, expectedDefaultValue)
			})
		})

		/////////////////////
		// Extension checks
		/////////////////////

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-terraform-field-name' extension", func() {
			expectedTerraformName := "property_terraform_name"
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfFieldName: expectedTerraformName,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.PreferredName, ShouldEqual, expectedTerraformName)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-terraform-force-new' extension", func() {
			expectedForceNewValue := true
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfForceNew: expectedForceNewValue,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.ForceNew, ShouldEqual, expectedForceNewValue)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-terraform-sensitive' extension", func() {
			expectedSensitiveValue := true
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfSensitive: expectedSensitiveValue,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Sensitive, ShouldEqual, expectedSensitiveValue)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)

			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-terraform-id' extension", func() {
			expectedIsIdentifierValue := true
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfID: expectedIsIdentifierValue,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IsIdentifier, ShouldEqual, expectedIsIdentifierValue)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-terraform-immutable' extension", func() {
			expectedIsImmutableValue := true
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfImmutable: expectedIsImmutableValue,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.Immutable, ShouldEqual, expectedIsImmutableValue)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-terraform-ignore-order' extension", func() {
			expectedIgnoreOrder := true
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
						},
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfIgnoreOrder: expectedIgnoreOrder,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IgnoreItemsOrder, ShouldEqual, expectedIgnoreOrder)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-ignore-order' extension", func() {
			expectedIgnoreOrder := true
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
						},
					},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extIgnoreOrder: expectedIgnoreOrder,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IgnoreItemsOrder, ShouldEqual, expectedIgnoreOrder)
			})
		})

		Convey("When createSchemaDefinitionProperty is called with a property schema that has the 'x-terraform-field-status' extension", func() {
			expectedIsStatusFieldValue := true
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfFieldStatus: expectedIsStatusFieldValue,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IsStatusIdentifier, ShouldEqual, expectedIsStatusFieldValue)
				So(schemaDefinitionProperty.isComputed(), ShouldBeFalse)
			})
		})

		Convey(fmt.Sprintf("When createSchemaDefinitionProperty is called with an optional property schema that has the %s extension (this means the property is optional-computed, and the value computed is not known at runtime)", extTfComputed), func() {
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be nil and the schemaDefinitionProperty should be configured as expected", func() {
				So(err, ShouldBeNil)
				So(schemaDefinitionProperty.IsRequired(), ShouldBeFalse)
				So(schemaDefinitionProperty.isReadOnly(), ShouldBeFalse)
				So(schemaDefinitionProperty.isComputed(), ShouldBeTrue)
				So(schemaDefinitionProperty.Default, ShouldBeNil)
			})
		})

		Convey(fmt.Sprintf("When createSchemaDefinitionProperty is called with an optional property schema that violates one optional-computed validation (properties with default attributes cannot have the %s extension too, that does not make any sense)", extTfComputed), func() {
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type:    spec.StringOrArray{"string"},
					Default: "someDefaultValue",
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be the expected one and the schemaDefinitionProperty should be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "optional computed property validation failed for property 'propertyName': optional computed properties with default attributes should not have 'x-terraform-computed' extension too")
				So(schemaDefinitionProperty, ShouldBeNil)
			})
		})

		Convey(fmt.Sprintf("When createSchemaDefinitionProperty is called with an optional property schema that violates one optional-computed validation (properties with %s extension, should not be readOnly)", extTfComputed), func() {
			propertySchema := spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			}
			schemaDefinitionProperty, err := r.createSchemaDefinitionProperty("propertyName", propertySchema, []string{})
			Convey("Then the error returned should be the expected one and the schemaDefinitionProperty should be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "optional computed property validation failed for property 'propertyName': optional computed properties marked with 'x-terraform-computed' can not be readOnly")
				So(schemaDefinitionProperty, ShouldBeNil)
			})
		})
	})
}

func TestIsBoolExtensionEnabled(t *testing.T) {
	testCases := []struct {
		name            string
		inputExtensions spec.Extensions
		inputExtension  string
		expectedResult  bool
	}{
		{name: "nil extensions", inputExtensions: nil, inputExtension: "", expectedResult: false},
		{name: "empty extensions", inputExtensions: spec.Extensions{}, inputExtension: "", expectedResult: false},
		{name: "populated extensions but empty extension", inputExtensions: spec.Extensions{"some-extension": true}, inputExtension: "", expectedResult: false},
		{name: "populated extensions and matching bool extension with value true", inputExtensions: spec.Extensions{"some-extension": true}, inputExtension: "some-extension", expectedResult: true},
		{name: "populated extensions and matching bool extension with value false", inputExtensions: spec.Extensions{"some-extension": false}, inputExtension: "some-extension", expectedResult: false},
		{name: "populated extensions and matching non bool extension", inputExtensions: spec.Extensions{"some-extension": "some value"}, inputExtension: "some-extension", expectedResult: false},
	}
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		for _, tc := range testCases {
			Convey(fmt.Sprintf("When isBoolExtensionEnabled method is called: %s", tc.name), func() {
				isEnabled := r.isBoolExtensionEnabled(tc.inputExtensions, tc.inputExtension)
				Convey("Then the result returned should be the expected one", func() {
					So(isEnabled, ShouldEqual, tc.expectedResult)
				})
			})
		}
	})
}

func TestIsOptionalComputedProperty(t *testing.T) {
	testCases := []struct {
		name                    string
		inputPropertyName       string
		inputProperty           spec.Schema
		inputRequiredProperties []string
		expectedResult          bool
		expectedError           error
	}{
		{
			name:              "property is required",
			inputPropertyName: "some_required_property_name",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				SchemaProps: spec.SchemaProps{
					Default: "some_defaul_value",
				},
			},
			inputRequiredProperties: []string{"some_required_property_name"},
			expectedResult:          false,
			expectedError:           nil,
		},
		{
			name:              "property is optional (it's not required) and matches the OptionalComputedWithDefault requirements (it's not readOnly/computed and has a default value)",
			inputPropertyName: "some_optional_property_name",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				SchemaProps: spec.SchemaProps{
					Default: "some_defaul_value",
				},
			},
			inputRequiredProperties: []string{"some_required_property_name"},
			expectedResult:          true,
			expectedError:           nil,
		},
		{
			name:              fmt.Sprintf("property is optional, and matches the IsOptionalComputed requirements (no computed and has the %s extension)", extTfComputed),
			inputPropertyName: "some_optional_property_name",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			},
			inputRequiredProperties: []string{"some_required_property_name"},
			expectedResult:          true,
			expectedError:           nil,
		},
		{
			name:              fmt.Sprintf("property is optional, and DOES NOT pass the validation as far as IsOptionalComputed requirements is concerned (properties with %s extension cannot be readOnly)", extTfComputed),
			inputPropertyName: "some_optional_property_name",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			},
			inputRequiredProperties: []string{"some_required_property_name"},
			expectedResult:          false,
			expectedError:           errors.New("optional computed property validation failed for property 'some_optional_property_name': optional computed properties marked with 'x-terraform-computed' can not be readOnly"),
		},
		{
			name:              "property that not optional computed (e,g: property is just computed)",
			inputPropertyName: "some_optional_property_name",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
			},
			inputRequiredProperties: []string{"some_required_property_name"},
			expectedResult:          false,
			expectedError:           nil,
		},
	}
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		for _, tc := range testCases {
			Convey(fmt.Sprintf("When isOptionalComputedProperty method is called: %s", tc.name), func() {
				isOptionalComputedProperty, err := r.isOptionalComputedProperty(tc.inputPropertyName, tc.inputProperty, tc.inputRequiredProperties)
				Convey("Then the result returned should be the expected one", func() {
					So(err, ShouldResemble, tc.expectedError)
					So(isOptionalComputedProperty, ShouldEqual, tc.expectedResult)
				})
			})
		}
	})
}

func TestIsOptionalComputedWithDefault(t *testing.T) {
	testCases := []struct {
		name           string
		inputProperty  spec.Schema
		expectedResult bool
		expectedError  error
	}{
		{
			name: "property that is NOT readOnly and has a default attribute",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				SchemaProps: spec.SchemaProps{
					Default: "some_defaul_value",
				},
			},
			expectedResult: true,
			expectedError:  nil,
		},
		{
			name: "property that is readOnly and has a default attribute",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
				SchemaProps: spec.SchemaProps{
					Default: "some_defaul_value",
				},
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name: "property that is NOT readOnly and has NO default attribute",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				SchemaProps: spec.SchemaProps{
					Default: nil,
				},
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name: "property that is just readOnly",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true,
				},
				SchemaProps: spec.SchemaProps{
					Default: nil,
				},
			},
			expectedResult: false,
			expectedError:  nil,
		},
		{
			name: "property that does not pass the validation phase since it has a default value AND the extension, this is wrong documentation",
			inputProperty: spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				SchemaProps: spec.SchemaProps{
					Default: "some_value",
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			},
			expectedResult: false,
			expectedError:  errors.New("optional computed property validation failed for property 'propertyName': optional computed properties with default attributes should not have 'x-terraform-computed' extension too"),
		},
	}
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		for _, tc := range testCases {
			Convey(fmt.Sprintf("When isOptionalComputedWithDefault method is called: %s", tc.name), func() {
				isOptionalComputedWithDefault, err := r.isOptionalComputedWithDefault("propertyName", tc.inputProperty)
				Convey("Then the result returned should be the expected one", func() {
					So(err, ShouldResemble, tc.expectedError)
					So(isOptionalComputedWithDefault, ShouldEqual, tc.expectedResult)
				})
			})
		}
	})
}

func TestIsOptionalComputed(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey(fmt.Sprintf("When IsOptionalComputed method is called with a property that is NOT computed (readOnly) and has the extension %s with value true", extTfComputed), func() {
			property := spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			}
			isOptionalComputed, err := r.isOptionalComputed("propertyName", property)
			Convey("Then the result returned should be true since the property matches the requirements to be an optional computed property", func() {
				So(err, ShouldBeNil)
				So(isOptionalComputed, ShouldBeTrue)
			})
		})
		Convey(fmt.Sprintf("When IsOptionalComputed method is called with a property that is NOT computed (readOnly) and has the extension %s with value false", extTfComputed), func() {
			property := spec.Schema{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: false,
					},
				},
			}
			isOptionalComputed, err := r.isOptionalComputed("propertyName", property)
			Convey("Then the result returned should be false since the property DOES NOT match the requirements to be an optional computed property", func() {
				So(err, ShouldBeNil)
				So(isOptionalComputed, ShouldBeFalse)
			})
		})
		Convey(fmt.Sprintf("When IsOptionalComputed method is called with a property that is computed (readOnly) and has the extension %s with value true", extTfComputed), func() {
			property := spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: true, // this specifies that the property is computed
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			}
			isOptionalComputed, err := r.isOptionalComputed("propertyName", property)
			Convey("Then the result returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "optional computed property validation failed for property 'propertyName': optional computed properties marked with 'x-terraform-computed' can not be readOnly")
				// the result returned should not be nil since properties with the x-terraform-computed extension cannot be computed
				So(err, ShouldNotBeNil)
				// the result returned should be false since the property DOES NOT match the requirements to be an optional computed property
				So(isOptionalComputed, ShouldBeFalse)
			})
		})
		Convey(fmt.Sprintf("When IsOptionalComputed method is called with a property that is optional, and DOES NOT pass the validation as far as IsOptionalComputed requirements is concerned (properties with %s extension cannot have default value populated)", extTfComputed), func() {
			property := spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{
					ReadOnly: false,
				},
				SchemaProps: spec.SchemaProps{
					Default: "some_defaul_value",
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfComputed: true,
					},
				},
			}
			isOptionalComputedProperty, err := r.isOptionalComputed("some_optional_property_name", property)
			Convey("Then the result returned should be false since the property is NOT optional computed ", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "optional computed property validation failed for property 'some_optional_property_name': optional computed properties marked with 'x-terraform-computed' can not have the default value as the value is not known at plan time. If the value is known, then this extension should not be used, and rather the 'default' attribute should be populated")
				So(isOptionalComputedProperty, ShouldBeFalse)
			})
		})
		Convey(fmt.Sprintf("When IsOptionalComputed method is called with a property that DOES NOT have the extension %s present", extTfComputed), func() {
			property := spec.Schema{
				SwaggerSchemaProps: spec.SwaggerSchemaProps{},
			}
			isOptionalComputed, err := r.isOptionalComputed("propertyName", property)
			Convey("Then the result returned should be false", func() {
				So(err, ShouldBeNil)
				So(isOptionalComputed, ShouldBeFalse)
			})
		})
	})
}

func TestIsArrayItemPrimitiveType(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When isArrayItemPrimitiveType method is called with a primitive type TypeString", func() {
			isPrimitive := r.isArrayItemPrimitiveType(TypeString)
			Convey("Then the result returned should be true", func() {
				So(isPrimitive, ShouldBeTrue)
			})
		})
		Convey("When isArrayItemPrimitiveType method is called with a primitive type TypeInt", func() {
			isPrimitive := r.isArrayItemPrimitiveType(TypeInt)
			Convey("Then the result returned should be true", func() {
				So(isPrimitive, ShouldBeTrue)
			})
		})
		Convey("When isArrayItemPrimitiveType method is called with a primitive type TypeFloat", func() {
			isPrimitive := r.isArrayItemPrimitiveType(TypeFloat)
			Convey("Then the result returned should be true", func() {
				So(isPrimitive, ShouldBeTrue)
			})
		})
		Convey("When isArrayItemPrimitiveType method is called with a primitive type TypeBool", func() {
			isPrimitive := r.isArrayItemPrimitiveType(TypeBool)
			Convey("Then the result returned should be true", func() {
				So(isPrimitive, ShouldBeTrue)
			})
		})
		Convey("When isArrayItemPrimitiveType method is called with a NON primitive type TypeList", func() {
			isPrimitive := r.isArrayItemPrimitiveType(TypeList)
			Convey("Then the result returned should be false", func() {
				So(isPrimitive, ShouldBeFalse)
			})
		})
		Convey("When isArrayItemPrimitiveType method is called with a NON primitive type TypeObject", func() {
			isPrimitive := r.isArrayItemPrimitiveType(TypeObject)
			Convey("Then the result returned should be false", func() {
				So(isPrimitive, ShouldBeFalse)
			})
		})
	})
}

func TestValidateArrayItems(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When validateArrayItems method is called with a property that does not have items", func() {
			property := spec.Schema{}
			_, err := r.validateArrayItems(property)
			Convey("Then the error message should be the expected", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "array property is missing items schema definition")
			})
		})
		Convey("When validateArrayItems method is called with a property that does have items but they lack the schema", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Items: &spec.SchemaOrArray{
						// no schema :(
					},
				},
			}
			_, err := r.validateArrayItems(property)
			Convey("Then the error message should be the expected", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "array property is missing items schema definition")
			})
		})
		Convey("When validateArrayItems method is called with a property that does have items and a schema BUT the items are of type array (this is not supported at the moment)", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"array"},
							},
						},
					},
				},
			}
			_, err := r.validateArrayItems(property)
			Convey("Then the error message should be the expected", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "array property can not have items of type 'array'")
			})
		})
		Convey("When validateArrayItems method is called with an array of unknown type items", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"unknown"},
							},
						},
					},
				},
			}
			_, err := r.validateArrayItems(property)
			Convey("Then the error message should be the expected", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "non supported '[unknown]' type")
			})
		})
		Convey("When validateArrayItems method is called with a valid array property that has items of type string", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
						},
					},
				},
			}
			itemsPropType, err := r.validateArrayItems(property)
			Convey("Then the type of the items should match the expected string", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeString)
			})
		})
		Convey("When validateArrayItems method is called with a valid array property that has items of type object", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Properties: map[string]spec.Schema{
									"prop1": {},
								},
							},
						},
					},
				},
			}
			itemsPropType, err := r.validateArrayItems(property)
			Convey("Then the type of the items should match the expected object", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeObject)
			})
		})
	})
}

func TestGetPropertyType(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When getPropertyType method is called with a property of type array", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
				},
			}
			itemsPropType, err := r.getPropertyType(property)
			Convey("Then the type of the items should match the expected array", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeList)
			})
		})

		Convey("When getPropertyType method is called with a property of type object", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					Properties: map[string]spec.Schema{
						"prop1": {},
					},
				},
			}
			itemsPropType, err := r.getPropertyType(property)
			Convey("Then the type of the items should match the expected object", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeObject)
			})
		})

		Convey("When getPropertyType method is called with a property of type string", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			itemsPropType, err := r.getPropertyType(property)
			Convey("Then the type of the items should match the expected string", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeString)
			})
		})

		Convey("When getPropertyType method is called with a property of type integer", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"integer"},
				},
			}
			itemsPropType, err := r.getPropertyType(property)
			Convey("Then the type of the items should match the expected integer", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeInt)
			})
		})

		Convey("When getPropertyType method is called with a property of type float", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"number"},
				},
			}
			itemsPropType, err := r.getPropertyType(property)
			Convey("Then the type of the items should match the expected float", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeFloat)
			})
		})

		Convey("When getPropertyType method is called with a property of type bool", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"boolean"},
				},
			}
			itemsPropType, err := r.getPropertyType(property)
			Convey("Then the type of the items should match the expected bool", func() {
				So(err, ShouldBeNil)
				So(itemsPropType, ShouldEqual, TypeBool)
			})
		})

		Convey("When getPropertyType method is called with a property of type non supported", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"non supported"},
				},
			}
			_, err := r.getPropertyType(property)
			Convey("Then the error returned should be as expected", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "non supported '[non supported]' type")
			})
		})
	})
}

func TestResourceIsObjectProperty(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When isObjectProperty method is called with a property of type object that has nested properties", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					Properties: map[string]spec.Schema{
						"prop1": {},
					},
				},
			}
			isObject, objectSchema, err := r.isObjectProperty(property)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(isObject, ShouldBeTrue)
				So(objectSchema, ShouldNotBeNil)
			})
		})
		Convey("When isObjectProperty method is called with a property of type object that has a ref to an external schema but is missing the type", func() {
			r := SpecV2Resource{
				SchemaDefinitions: map[string]spec.Schema{
					"Listeners": {
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"object"},
							Properties: map[string]spec.Schema{
								"protocol": {
									SchemaProps: spec.SchemaProps{
										Type: spec.StringOrArray{"string"},
									},
								},
							},
						},
					},
				},
			}
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					//Type: spec.StringOrArray{"object"}, // Missing type info but still should be considered as object
					Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/Listeners")},
				},
			}
			isObject, objectSchema, err := r.isObjectProperty(propertySchema)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(isObject, ShouldBeTrue)
				So(objectSchema, ShouldNotBeNil)
			})
		})
		Convey("When isObjectProperty method is called with a property of type object that has a ref to an external schema and also has the type", func() {
			r := SpecV2Resource{
				SchemaDefinitions: map[string]spec.Schema{
					"Listeners": {
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"object"},
							Properties: map[string]spec.Schema{
								"protocol": {
									SchemaProps: spec.SchemaProps{
										Type: spec.StringOrArray{"string"},
									},
								},
							},
						},
					},
				},
			}
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					Ref:  spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/Listeners")},
				},
			}
			isObject, objectSchema, err := r.isObjectProperty(propertySchema)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(isObject, ShouldBeTrue)
				So(objectSchema, ShouldNotBeNil)
			})
		})

		Convey("When isObjectProperty method is called with a property of type object that has a ref to a non existing schema", func() {
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					Ref:  spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/nonExisting")},
				},
			}
			isObject, objectSchema, err := r.isObjectProperty(propertySchema)
			Convey("The error should NOT be nil", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "object ref is poitning to a non existing schema definition: missing schema definition in the swagger file with the supplied ref '#/definitions/nonExisting'")
				So(isObject, ShouldBeTrue)
				So(objectSchema, ShouldBeNil)
			})
		})

		Convey("When isObjectProperty method is called with a property that has nested schema with no properties", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{},
					},
				},
			}
			isObject, objectSchema, err := r.isObjectProperty(property)
			Convey("The error should NOT be nil", func() {
				So(err, ShouldBeNil)
				So(isObject, ShouldBeFalse)
				So(objectSchema, ShouldBeNil)
			})
		})

		Convey("When isObjectProperty method is called with a property of type string", func() {
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			isObject, objectSchema, err := r.isObjectProperty(propertySchema)
			Convey("The error should be nil", func() {
				So(err, ShouldBeNil)
				So(isObject, ShouldBeFalse)
				So(objectSchema, ShouldBeNil)
			})
		})
	})
}

func TestResourceIsArrayProperty(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When isArrayProperty is called with an array type property that has items of type object (nested)", func() {
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
								Properties: map[string]spec.Schema{
									"prop1": {
										SchemaProps: spec.SchemaProps{
											Type: spec.StringOrArray{"string"},
										},
									},
									"prop2": {
										SchemaProps: spec.SchemaProps{
											Type: spec.StringOrArray{"integer"},
										},
									},
								},
							},
						},
					},
				},
			}
			isArray, arrayItemType, objectItemSchema, err := r.isArrayProperty(propertySchema)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(isArray, ShouldBeTrue)
				So(arrayItemType, ShouldEqual, TypeObject)
				So(objectItemSchema, ShouldNotBeNil)
				exists, _ := assertPropertyExists(objectItemSchema.Properties, "prop1")
				So(exists, ShouldBeTrue)
				exists, _ = assertPropertyExists(objectItemSchema.Properties, "prop2")
				So(exists, ShouldBeTrue)
			})
		})
		Convey("When isArrayProperty is called with an array type property that has items of type primitive (e,g: string)", func() {
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"string"},
							},
						},
					},
				},
			}
			isArray, arrayItemType, objectItemSchema, err := r.isArrayProperty(propertySchema)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(isArray, ShouldBeTrue)
				So(arrayItemType, ShouldEqual, TypeString)
				So(objectItemSchema, ShouldBeNil)
			})
		})
		Convey("When isArrayProperty is called with an array type property that has items of type object (ref)", func() {
			r := SpecV2Resource{
				SchemaDefinitions: map[string]spec.Schema{
					"Listeners": {
						SchemaProps: spec.SchemaProps{
							Type: spec.StringOrArray{"object"},
							Properties: map[string]spec.Schema{
								"protocol": {
									SchemaProps: spec.SchemaProps{
										Type: spec.StringOrArray{"string"},
									},
								},
							},
						},
					},
				},
			}
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
					Items: &spec.SchemaOrArray{
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: spec.Ref{Ref: jsonreference.MustCreateRef("#/definitions/Listeners")},
							},
						},
					},
				},
			}
			isArray, arrayItemType, objectItemSchema, err := r.isArrayProperty(propertySchema)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(isArray, ShouldBeTrue)
				So(arrayItemType, ShouldEqual, TypeObject)
				So(objectItemSchema, ShouldNotBeNil)
				exists, _ := assertPropertyExists(objectItemSchema.Properties, "protocol")
				So(exists, ShouldBeTrue)
			})
		})
		Convey("When isArrayProperty is called with a NON array property type", func() {
			propertySchema := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			isArray, _, objectItemSchema, err := r.isArrayProperty(propertySchema)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(isArray, ShouldBeFalse)
				So(objectItemSchema, ShouldBeNil)
			})
		})
	})
}

func TestIsObjectTypeProperty(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When isObjectTypeProperty method is called a property of type object", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
				},
			}
			isArrayType := r.isObjectTypeProperty(property)
			Convey("Then the result returned should be true", func() {
				So(isArrayType, ShouldBeTrue)
			})
		})
		Convey("When isObjectTypeProperty method is called a property that IS NOT of type object", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
				},
			}
			isArrayType := r.isObjectTypeProperty(property)
			Convey("Then the result returned should be false", func() {
				So(isArrayType, ShouldBeFalse)
			})
		})
	})
}

func TestIsArrayTypeProperty(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When isArrayTypeProperty method is called a property of type array", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"array"},
				},
			}
			isArrayType := r.isArrayTypeProperty(property)
			Convey("Then the result returned should be true", func() {
				So(isArrayType, ShouldBeTrue)
			})
		})
		Convey("When isArrayTypeProperty method is called a property that IS NOT of type array", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
				},
			}
			isArrayType := r.isArrayTypeProperty(property)
			Convey("Then the result returned should be false", func() {
				So(isArrayType, ShouldBeFalse)
			})
		})
	})
}

func TestIsOfType(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When isOfType method is called a property of the expected type", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			isString := r.isOfType(property, "string")
			Convey("Then the result returned should be true", func() {
				So(isString, ShouldBeTrue)
			})
		})
		Convey("When isArrayTypeProperty method is called a property that IS NOT of the expected type", func() {
			property := spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"string"},
				},
			}
			isInteger := r.isOfType(property, "integer")
			Convey("Then the result returned should be false", func() {
				So(isInteger, ShouldBeFalse)
			})
		})
	})
}

func TestSwaggerPropIsRequired(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := &SpecV2Resource{}
		Convey("When IsRequired is called with a required prop", func() {
			requiredProp := "requiredProp"
			requiredProps := []string{requiredProp}
			isRequired := r.isRequired(requiredProp, requiredProps)
			Convey("Then the result returned should be true", func() {
				So(isRequired, ShouldBeTrue)
			})
		})
		Convey("When IsRequired is called with a NON required prop", func() {
			requiredProps := []string{"requiredProp"}
			isRequired := r.isRequired("nonRequired", requiredProps)
			Convey("Then the result returned should be true", func() {
				So(isRequired, ShouldBeFalse)
			})
		})
	})
}

func TestGetResourceTerraformName(t *testing.T) {
	Convey("Given a SpecV2Resource with a root path item containing a post operation with the extension 'x-terraform-resource-name'", t, func() {
		extensions := spec.Extensions{}
		expectedResourceName := "example"
		extensions.Add(extTfResourceName, expectedResourceName)
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: extensions,
						},
					},
				},
			},
		}
		Convey("When getResourceTerraformName method is called an existing extension", func() {
			value := r.getResourceTerraformName()
			Convey("Then the value returned should match the value in the extension", func() {
				So(value, ShouldEqual, expectedResourceName)
			})
		})
	})
	Convey("Given a SpecV2Resource with a root path item containing the extension 'x-terraform-resource-name'", t, func() {
		extensions := spec.Extensions{}
		expectedResourceName := "rootLevelPreferredName"
		extensions.Add(extTfResourceName, expectedResourceName)
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				VendorExtensible: spec.VendorExtensible{
					Extensions: extensions,
				},
			},
		}
		Convey("When getResourceTerraformName method is called an existing extension", func() {
			value := r.getResourceTerraformName()
			Convey("Then the value returned should match the value in the extension", func() {
				So(value, ShouldEqual, expectedResourceName)
			})
		})
	})
	Convey("Given a SpecV2Resource without a rootPathItem", t, func() {
		r := SpecV2Resource{}
		Convey("When getResourceTerraformName method is called", func() {
			value := r.getResourceTerraformName()
			Convey("Then the value returned should be empty since the resource does not have such extension defined", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
}

func TestGetPreferredName(t *testing.T) {
	testCases := []struct {
		name                 string
		inputPathItem        spec.PathItem
		expectedResourceName string
	}{
		{
			name: "path item with the extension 'x-terraform-resource-name' on the POST level",
			inputPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{extTfResourceName: "postLevelPreferredName"},
						},
					},
				},
			},
			expectedResourceName: "postLevelPreferredName",
		},
		{
			name: "path item with the extension 'x-terraform-resource-name' on the root level",
			inputPathItem: spec.PathItem{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{extTfResourceName: "rootLevelPreferredName"},
				},
			},
			expectedResourceName: "rootLevelPreferredName",
		},
		{
			name: "path item with the extension 'x-terraform-resource-name' on the POST and a different extension on the root level",
			inputPathItem: spec.PathItem{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						"x-something": "something ext value",
					},
				},
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{extTfResourceName: "postLevelPreferredName"},
						},
					},
				},
			},
			expectedResourceName: "postLevelPreferredName",
		},
		{
			name: "path item with the extension 'x-terraform-resource-name' on both the POST and root levels",
			inputPathItem: spec.PathItem{
				VendorExtensible: spec.VendorExtensible{
					Extensions: spec.Extensions{
						extTfResourceName: "rootLevelPreferredName",
					},
				},
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceName: "postPreferredName",
							},
						},
					},
				},
			},
			expectedResourceName: "rootLevelPreferredName",
		},
		{
			name:                 " an empty path item",
			inputPathItem:        spec.PathItem{},
			expectedResourceName: "",
		},
	}

	for _, tc := range testCases {
		specV2Resource := SpecV2Resource{}
		value := specV2Resource.getPreferredName(tc.inputPathItem)
		assert.Equal(t, tc.expectedResourceName, value, tc.name)
	}
}

func TestGetExtensionStringValue(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When getExtensionStringValue method is called an existing extension", func() {
			extensions := spec.Extensions{}
			expectedExtensionValue := "example"
			extensions.Add(extTfResourceName, expectedExtensionValue)
			value := r.getExtensionStringValue(extensions, extTfResourceName)
			Convey("Then the value returned should match the value in the extension", func() {
				So(value, ShouldEqual, expectedExtensionValue)
			})
		})
		Convey("When getExtensionStringValue method is called a NON existing extension", func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceName, "example")
			value := r.getExtensionStringValue(extensions, "somethingOtherExtensionName")
			Convey("Then the value returned should match the value in the extension", func() {
				So(value, ShouldEqual, "")
			})
		})
	})
}

func TestCreateResponses(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When createResponses method is called with an operation that has the 'x-terraform-resource-poll-enabled' extension set to true", func() {
			expectedTarget := "deployed"
			expectedStatus := "deploy_pending"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollEnabled, true)
			extensions.Add(extTfResourcePollTargetStatuses, expectedTarget)
			extensions.Add(extTfResourcePollPendingStatuses, expectedStatus)
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{
								http.StatusAccepted: {
									VendorExtensible: spec.VendorExtensible{
										Extensions: extensions,
									},
								},
							},
						},
					},
				},
			}
			specResponses := r.createResponses(operation)
			Convey("Then the result returned should be the expected one", func() {
				So(specResponses, ShouldNotBeEmpty)
				So(specResponses, ShouldContainKey, http.StatusAccepted)
				So(specResponses[http.StatusAccepted].isPollingEnabled, ShouldBeTrue)
				So(specResponses[http.StatusAccepted].pollTargetStatuses, ShouldContain, expectedTarget)
				So(specResponses[http.StatusAccepted].pollPendingStatuses, ShouldContain, expectedStatus)
			})
		})

		Convey("When createResponses method is called with an operation does not have any status responses", func() {
			operation := &spec.Operation{
				OperationProps: spec.OperationProps{
					Responses: &spec.Responses{
						ResponsesProps: spec.ResponsesProps{
							StatusCodeResponses: map[int]spec.Response{},
						},
					},
				},
			}
			specResponses := r.createResponses(operation)
			Convey("Then the spec responses map returned should not be empty", func() {
				So(specResponses, ShouldBeEmpty)
			})
		})
	})
}

func TestIsResourcePollingEnabled(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When isResourcePollingEnabled method is called with a list of responses where one of the responses matches the response status received and has the 'x-terraform-resource-poll-enabled' extension set to true", func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollEnabled, true)
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			isResourcePollingEnabled := r.isResourcePollingEnabled(responses.StatusCodeResponses[http.StatusAccepted])
			Convey("Then the bool returned should be true", func() {
				So(isResourcePollingEnabled, ShouldBeTrue)
			})
		})
		Convey("When isResourcePollingEnabled method is called with a list of responses where one of the responses matches the response status received and has the 'x-terraform-resource-poll-enabled' extension set to false", func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollEnabled, false)
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			isResourcePollingEnabled := r.isResourcePollingEnabled(responses.StatusCodeResponses[http.StatusAccepted])
			Convey("Then the bool returned should be false", func() {
				So(isResourcePollingEnabled, ShouldBeFalse)
			})
		})
		Convey("When isResourcePollingEnabled method is called with list of responses where non of the codes match the given response http code", func() {
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusOK: {},
					},
				},
			}
			isResourcePollingEnabled := r.isResourcePollingEnabled(responses.StatusCodeResponses[http.StatusOK])
			Convey("Then bool returned should be false", func() {
				So(isResourcePollingEnabled, ShouldBeFalse)
			})
		})
	})
}

func TestGetResourcePollTargetStatuses(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When getResourcePollTargetStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses'", func() {
			expectedTarget := "deployed"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTarget)
			responses := &spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses := r.getResourcePollTargetStatuses(responses.StatusCodeResponses[http.StatusAccepted])
			Convey("Then the status returned should contain", func() {
				So(statuses, ShouldContain, expectedTarget)
			})
		})
	})
}

func TestGetResourcePollPendingStatuses(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When getResourcePollPendingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-pending-statuses'", func() {
			expectedStatus := "deploy_pending"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollPendingStatuses, expectedStatus)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses := r.getResourcePollPendingStatuses(responses.StatusCodeResponses[http.StatusAccepted])
			Convey("Then the status returned should contain", func() {
				So(statuses, ShouldContain, expectedStatus)
			})
		})
	})
}

func TestGetPollingStatuses(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When getPollingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses'", func() {
			expectedTarget := "deployed"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTarget)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the statuses returned should contain", func() {
				So(statuses, ShouldContain, expectedTarget)
			})
		})

		Convey("When getPollingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses' containing multiple targets (comma separated with spaces)", func() {
			expectedTargets := "deployed, completed, done"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTargets)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the statuses returned should contain expected targets", func() {
				// the expected Targets are a list of targets with no spaces whatsoever, hence why the removal of spaces
				for _, expectedTarget := range strings.Split(strings.Replace(expectedTargets, " ", "", -1), ",") {
					So(statuses, ShouldContain, expectedTarget)
				}
			})
		})

		Convey("When getPollingStatuses method is called with a response that has a given extension 'x-terraform-resource-poll-completed-statuses' containing multiple targets (comma separated with no spaces)", func() {
			expectedTargets := "deployed,completed,done"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourcePollTargetStatuses, expectedTargets)
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: extensions,
							},
						},
					},
				},
			}
			statuses := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the statuses returned should contain expected targets", func() {
				for _, expectedTarget := range strings.Split(expectedTargets, ",") {
					So(statuses, ShouldContain, expectedTarget)
				}
			})
		})

		Convey("When getPollingStatuses method is called with a response that has does not have a given extension 'x-terraform-resource-poll-completed-statuses'", func() {
			responses := spec.Responses{
				ResponsesProps: spec.ResponsesProps{
					StatusCodeResponses: map[int]spec.Response{
						http.StatusAccepted: {
							VendorExtensible: spec.VendorExtensible{
								Extensions: spec.Extensions{},
							},
						},
					},
				},
			}
			statuses := r.getPollingStatuses(responses.StatusCodeResponses[http.StatusAccepted], extTfResourcePollTargetStatuses)
			Convey("Then the status returned should be empty", func() {
				So(statuses, ShouldBeEmpty)
			})
		})
	})
}

func TestGetTimeouts(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		expectedTimeout := "30s"
		extensions := spec.Extensions{}
		extensions.Add(extTfResourceTimeout, expectedTimeout)
		op := &spec.Operation{
			VendorExtensible: spec.VendorExtensible{
				Extensions: extensions,
			},
		}
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: op,
				},
			},
			InstancePathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Put:    op,
					Get:    op,
					Delete: op,
				},
			},
		}
		Convey("When getTimeouts method is called ", func() {
			timeouts, err := r.getTimeouts()
			Convey("Then the duration returned should contain the expected duration from the operation", func() {
				So(err, ShouldBeNil)
				So(*timeouts.Post, ShouldEqual, time.Duration(30*time.Second))
				So(*timeouts.Get, ShouldEqual, time.Duration(30*time.Second))
				So(*timeouts.Put, ShouldEqual, time.Duration(30*time.Second))
				So(*timeouts.Delete, ShouldEqual, time.Duration(30*time.Second))
			})
		})
	})
}

func TestGetResourceTimeout(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey(fmt.Sprintf("When getResourceTimeout method is called with an operation that has the extension '%s'", extTfResourceTimeout), func() {
			expectedTimeout := "30s"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			post := &spec.Operation{
				VendorExtensible: spec.VendorExtensible{
					Extensions: extensions,
				},
			}
			duration, err := r.getResourceTimeout(post)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(*duration, ShouldEqual, time.Duration(30*time.Second))
			})
		})
	})
}

func TestGetTimeDuration(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in seconds", extTfResourceTimeout), func() {
			expectedTimeout := "30s"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(*duration, ShouldEqual, time.Duration(30*time.Second))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in minutes (using fractions)", extTfResourceTimeout), func() {
			expectedTimeout := "20.5m"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(*duration, ShouldEqual, time.Duration((20*time.Minute)+(30*time.Second)))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in hours", extTfResourceTimeout), func() {
			expectedTimeout := "1h"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(*duration, ShouldEqual, time.Duration(1*time.Hour))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES NOT contain the extension passed in '%s'", extTfResourceTimeout), func() {
			expectedTimeout := "30s"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, "nonExistingExtension")
			Convey("Then the duration returned should be nil", func() {
				So(err, ShouldBeNil)
				So(duration, ShouldBeNil)
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is an empty string", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid duration value: ''. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
				So(duration, ShouldBeNil)
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is a negative duration", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "-1.5h")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid duration value: '-1.5h'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
				So(duration, ShouldBeNil)
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is NOT supported (distinct than s,m and h)", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "300ms")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldContainSubstring, "invalid duration value: '300ms'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
				So(duration, ShouldBeNil)
			})
		})
	})
}

func TestGetDuration(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When getDuration method is called a valid formatted time'", func() {
			duration, err := r.getDuration("30s")
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(*duration, ShouldEqual, time.Duration(30*time.Second))
			})
		})
		Convey("When getDuration method is called a invalid formatted time'", func() {
			_, err := r.getDuration("some invalid formatted time")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestSpecV2ResourceGetHost(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: "www.some-host.com",
							},
						},
					},
				},
			},
		}
		Convey("When getHost is called", func() {
			host, err := r.getHost()
			Convey("Then the host returned should be the override host", func() {
				So(err, ShouldBeNil)
				So(host, ShouldEqual, "www.some-host.com")
			})
		})
	})
	Convey("Given a SpecV2Resource without an override host specified", t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{},
				},
			},
		}
		Convey("When getHost is called", func() {
			host, err := r.getHost()
			Convey("Then the host returned should be the override host", func() {
				So(err, ShouldBeNil)
				So(host, ShouldEqual, "")
			})
		})
	})
	Convey("Given a SpecV2Resource that doesn't have a POST operation specified", t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: nil,
				},
			},
		}
		Convey("When getHost is called", func() {
			host, err := r.getHost()
			Convey("Then the host returned should be an empty string", func() {
				So(err, ShouldBeNil)
				So(host, ShouldBeEmpty)
			})
		})
	})
	Convey("Given a SpecV2Resource that is multi region but region isn't specified", t, func() {
		r := SpecV2Resource{
			Region: "",
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: "www.${region}.some-host.com",
							},
						},
					},
				},
			},
		}
		Convey("When getHost is called", func() {
			host, err := r.getHost()
			Convey("Then the host returned should be an empty string", func() {
				So(err.Error(), ShouldEqual, "region can not be empty for multiregion resources")
				So(host, ShouldBeEmpty)
			})
		})
	})
	Convey("Given a SpecV2Resource that is multi region with region specified", t, func() {
		r := SpecV2Resource{
			Region: "rst1",
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: "www.${region}.some-host.com",
							},
						},
					},
				},
			},
		}
		Convey("When getHost is called", func() {
			host, err := r.getHost()
			Convey("Then the host returned should be the override host", func() {
				So(err, ShouldBeNil)
				So(host, ShouldEqual, "www.rst1.some-host.com")
			})
		})
	})
}

func TestGetResourceOverrideHost(t *testing.T) {
	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-resource-host with a non parametrized host containing the host to use", t, func() {
		expectedHost := "some.api.domain.com"
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: expectedHost,
							},
						},
					},
				},
			},
		}
		Convey("When getResourceOverrideHost method is called", func() {
			host := getResourceOverrideHost(r.RootPathItem.Post)
			Convey("Then the value returned should be the host value", func() {
				So(host, ShouldEqual, expectedHost)
			})
		})
	})

	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-resource-host with a parametrized host containing the host to use", t, func() {
		expectedHost := "some.api.${serviceProviderName}.domain.com"
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: expectedHost,
							},
						},
					},
				},
			},
		}
		Convey("When getResourceOverrideHost method is called", func() {
			host := getResourceOverrideHost(r.RootPathItem.Post)
			Convey("Then the value returned should be the host value", func() {
				So(host, ShouldEqual, expectedHost)
			})
		})
	})

	Convey("Given a terraform compliant resource that has a POST operation containing the x-terraform-resource-host with an empty string value", t, func() {
		expectedHost := ""
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: &spec.Operation{
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								extTfResourceURL: expectedHost,
							},
						},
					},
				},
			},
		}
		Convey("When getResourceOverrideHost method is called", func() {
			host := getResourceOverrideHost(r.RootPathItem.Post)
			Convey("Then the value returned should be the host value", func() {
				So(host, ShouldEqual, expectedHost)
			})
		})
	})

	Convey("Given a terraform resource that doesn't have a POST operation", t, func() {
		r := SpecV2Resource{
			RootPathItem: spec.PathItem{
				PathItemProps: spec.PathItemProps{
					Post: nil,
				},
			},
		}
		Convey("When getResourceOverrideHost method is called", func() {
			host := getResourceOverrideHost(r.RootPathItem.Post)
			Convey("Then the value returned should be an empty string", func() {
				So(host, ShouldEqual, "")
			})
		})
	})
}
