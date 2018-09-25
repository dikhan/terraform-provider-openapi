package openapi

import (
	"fmt"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestNewSpecV2Resource(t *testing.T) {
	Convey("Given a root path /users/ containing a trailing slash and a root path item item", t, func() {
		path := "/users/"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When getResourceName method is called", func() {
			r, err := newSpecV2Resource(path, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("The the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be 'users'", func() {
				So(r.getResourceName(), ShouldEqual, "users")
			})
		})
	})
	Convey("Given a root path /users with no trailing slash and a root path item item", t, func() {
		path := "/users"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When getResourceName method is called", func() {
			r, err := newSpecV2Resource(path, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("The the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be 'users'", func() {
				So(r.getResourceName(), ShouldEqual, "users")
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
		Convey("When getResourceName method is called", func() {
			r, err := newSpecV2Resource(path, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("The the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be 'users_v1'", func() {
				So(r.getResourceName(), ShouldEqual, "users_v1")
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
		Convey("When getResourceName method is called", func() {
			r, err := newSpecV2Resource(path, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("The the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be 'users_v1'", func() {
				So(r.getResourceName(), ShouldEqual, "users_v12")
			})
		})
	})

	Convey("Given a root path that is versioned such as '/v1/something/users' and a root path item item", t, func() {
		path := "/v1/something/users"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When getResourceName method is called", func() {
			r, err := newSpecV2Resource(path, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("The the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should still be 'users_v1'", func() {
				So(r.getResourceName(), ShouldEqual, "users_v1")
			})
		})
	})

	Convey("Given a root path which has path parameters '/api/v1/nodes/{name}/proxy' and a root path item item", t, func() {
		path := "/api/v1/nodes/{name}/proxy"
		rootPathItem := spec.PathItem{
			PathItemProps: spec.PathItemProps{
				Post: &spec.Operation{},
			},
		}
		Convey("When getResourceName method is called", func() {
			r, err := newSpecV2Resource(path, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("The the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should still be 'proxy_v1'", func() {
				So(r.getResourceName(), ShouldEqual, "proxy_v1")
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
		Convey("When getResourceName method is called", func() {
			r, err := newSpecV2Resource(path, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("The the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			resourceName := r.getResourceName()
			expectedTerraformName := fmt.Sprintf("%s_v1", expectedResourceName)
			Convey(fmt.Sprintf("And the value returned should still be '%s'", expectedTerraformName), func() {
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
		Convey("When newSpecV2Resource method is called", func() {
			_, err := newSpecV2Resource(invalidRootPath, spec.Schema{}, rootPathItem, spec.PathItem{})
			Convey("And the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})

	Convey("Given an empty path", t, func() {
		path := ""
		Convey("When newSpecV2Resource method is called", func() {
			_, err := newSpecV2Resource(path, spec.Schema{}, spec.PathItem{}, spec.PathItem{})
			Convey("And the err returned should not be nil", func() {
				So(err, ShouldNotBeNil)
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain the expected duration from the operation", func() {
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
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
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
				So(*duration, ShouldEqual, time.Duration(30*time.Second))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in minutes (using fractions)", extTfResourceTimeout), func() {
			expectedTimeout := "20.5m"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
				So(*duration, ShouldEqual, time.Duration((20*time.Minute)+(30*time.Second)))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that contains the extension passed in '%s' with value in hours", extTfResourceTimeout), func() {
			expectedTimeout := "1h"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should contain", func() {
				So(*duration, ShouldEqual, time.Duration(1*time.Hour))
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES NOT contain the extension passed in '%s'", extTfResourceTimeout), func() {
			expectedTimeout := "30s"
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, expectedTimeout)
			duration, err := r.getTimeDuration(extensions, "nonExistingExtension")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is an empty string", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "invalid duration value: ''. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is a negative duration", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "-1.5h")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "invalid duration value: '-1.5h'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
			})
		})
		Convey(fmt.Sprintf("When getTimeDuration method is called with a list of extensions that DOES contain the extension passed in '%s' BUT the value is NOT supported (distinct than s,m and h)", extTfResourceTimeout), func() {
			extensions := spec.Extensions{}
			extensions.Add(extTfResourceTimeout, "300ms")
			duration, err := r.getTimeDuration(extensions, extTfResourceTimeout)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the duration returned should be nil", func() {
				So(duration, ShouldBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "invalid duration value: '300ms'. The value must be a sequence of decimal numbers each with optional fraction and a unit suffix (negative durations are not allowed). The value must be formatted either in seconds (s), minutes (m) or hours (h)")
			})
		})
	})
}

func TestGetDuration(t *testing.T) {
	Convey("Given a SpecV2Resource", t, func() {
		r := SpecV2Resource{}
		Convey("When getDuration method is called a valid formatted time'", func() {
			duration, err := r.getDuration("30s")
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the statuses returned should contain", func() {
				fmt.Println(duration)
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
