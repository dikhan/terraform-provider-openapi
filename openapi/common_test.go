package openapi

import (
	"errors"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func TestCheckHTTPStatusCode(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		openAPIResource := &specStubResource{name: "resourceName"}
		Convey("When checkHTTPStatusCode is called with a response containing a status codes that matches one of the expected response status codes", func() {
			response := &http.Response{
				StatusCode: http.StatusOK,
			}
			expectedStatusCodes := []int{http.StatusOK}
			err := checkHTTPStatusCode(openAPIResource, response, expectedStatusCodes)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
		})
		Convey("When checkHTTPStatusCode is called with a response that IS NOT expected", func() {
			response := &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("some backend error")),
				StatusCode: http.StatusInternalServerError,
			}
			expectedStatusCodes := []int{http.StatusOK}
			err := checkHTTPStatusCode(openAPIResource, response, expectedStatusCodes)
			Convey("Then the err returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the err messages should equal", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] HTTP Response Status Code 500 not matching expected one [200] (some backend error)")
			})
		})
		Convey("When checkHTTPStatusCode is called with a response known with code 401 Unauthorized", func() {
			response := &http.Response{
				Body:       ioutil.NopCloser(strings.NewReader("unauthorized")),
				StatusCode: http.StatusUnauthorized,
			}
			expectedStatusCodes := []int{http.StatusOK}
			err := checkHTTPStatusCode(openAPIResource, response, expectedStatusCodes)
			Convey("Then the err returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("Then the err messages should equal", func() {
				So(err.Error(), ShouldEqual, "[resource='resourceName'] HTTP Response Status Code 401 - Unauthorized: API access is denied due to invalid credentials (unauthorized)")
			})
		})
	})
}

func TestResponseContainsExpectedStatus(t *testing.T) {
	Convey("Given a resource factory", t, func() {
		Convey("When responseContainsExpectedStatus is called with a response code that exists in the given list of expected status codes", func() {
			expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
			responseCode := http.StatusCreated
			exists := responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the expectedValue returned should be true", func() {
				So(exists, ShouldBeTrue)
			})
		})
		Convey("When responseContainsExpectedStatus is called with a response code that DOES NOT exists in 'expectedResponseStatusCodes'", func() {
			expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
			responseCode := http.StatusUnauthorized
			exists := responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the expectedValue returned should be false", func() {
				So(exists, ShouldBeFalse)
			})
		})
	})
}

func TestGetParentIDsAndResourcePath(t *testing.T) {
	Convey("Given an nil openapi resource (internal getParentIDs call fails for some reason)", t, func() {
		Convey("When getParentIDsAndResourcePath is called", func() {
			parentIDs, resourcePath, err := getParentIDsAndResourcePath(nil, nil)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "can't get parent ids from a resourceFactory with no openAPIResource")
			})
			Convey("And the parentIDs should be empty", func() {
				So(parentIDs, ShouldBeEmpty)
			})
			Convey("And the resourcePath should be empty", func() {
				So(resourcePath, ShouldBeEmpty)
			})
		})
	})

	Convey("Given an empty openapi resource (internal getResourcePath() call fails for some reason)", t, func() {
		someFirewallProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")
		testSchema := newTestSchema(someFirewallProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		openAPIResource := &specStubResource{
			funcGetResourcePath: func(parentIDs []string) (s string, e error) {
				return "", errors.New("getResourcePath() failed")
			}}

		Convey("When getParentIDsAndResourcePath is called", func() {
			parentIDs, resourcePath, err := getParentIDsAndResourcePath(openAPIResource, resourceData)
			Convey("Then the error returned should be the expected one", func() {
				So(err.Error(), ShouldEqual, "getResourcePath() failed")
			})
			Convey("And the parentIDs should be empty", func() {
				So(parentIDs, ShouldBeEmpty)
			})
			Convey("And the resourcePath should be empty", func() {
				So(resourcePath, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resource configured with a subreousrce", t, func() {
		someFirewallProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")

		// Pretending the data has already been populated with the parent property
		testSchema := newTestSchema(someFirewallProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		openAPIResource := &SpecV2Resource{
			Path: "/v1/cdns/{id}/firewall",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"some_string_prop"},
					Properties: map[string]spec.Schema{
						"some_string_prop": spec.Schema{
							SchemaProps: spec.SchemaProps{
								Required: []string{},
							},
						},
					},
				},
			},
		}

		Convey("When getParentIDsAndResourcePath is called", func() {
			parentIDs, resourcePath, err := getParentIDsAndResourcePath(openAPIResource, resourceData)
			Convey("Then the error returned should be the expected one", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the parentIDs should match the expected", func() {
				So(len(parentIDs), ShouldEqual, 1)
				So(parentIDs[0], ShouldEqual, "parentPropertyID")
			})
			Convey("And the resourcePath be '/v1/cdns/parentPropertyID/firewall'", func() {
				So(resourcePath, ShouldEqual, "/v1/cdns/parentPropertyID/firewall")
			})
		})
	})
}

func Test_getParentIDs(t *testing.T) {

	Convey("Given a resourceFactory with no openAPIResource", t, func() {
		rf := resourceFactory{}
		Convey("When getParentIDs is called", func() {
			ss, e := rf.getParentIDs(nil)
			Convey("Then an error is raised", func() {
				So(e.Error(), ShouldEqual, "can't get parent ids from a resourceFactory with no openAPIResource")
			})
			Convey("And the slice of string returned is empty", func() {
				So(ss, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resourceFactory with a pointer to a blank SpecV2Resource", t, func() {
		rf := resourceFactory{openAPIResource: &SpecV2Resource{}}
		Convey("When getParentIDs is called with a nil arg", func() {
			ss, err := rf.getParentIDs(nil)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the slice of string returned is empty", func() {
				So(ss, ShouldBeEmpty)
			})
		})
		Convey("When getParentIDs is called with an empty ResourceData", func() {
			ss, err := rf.getParentIDs(&schema.ResourceData{})
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the slice of string returned is empty", func() {
				So(ss, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a resourceFactory with a some schema", t, func() {
		someFirewallProperty := newStringSchemaDefinitionPropertyWithDefaults("some_string_prop", "", true, false, "some value")
		parentProperty := newStringSchemaDefinitionPropertyWithDefaults("cdns_v1_id", "", true, false, "parentPropertyID")

		// Pretending the data has already been populated with the parent property
		testSchema := newTestSchema(someFirewallProperty, parentProperty)
		resourceData := testSchema.getResourceData(t)

		rf := newResourceFactory(&SpecV2Resource{
			Path: "/v1/cdns/{id}/firewall",
			SchemaDefinition: spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{"some_string_prop"},
					Properties: map[string]spec.Schema{
						"some_string_prop": spec.Schema{
							SchemaProps: spec.SchemaProps{
								Required: []string{},
							},
						},
					},
				},
			},
		})

		Convey("When getParentIDs is called with non-empty ResourceData", func() {
			parentIDs, err := rf.getParentIDs(resourceData)
			Convey("Then the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the parent IDs returned should be populated as expected", func() {
				So(parentIDs[0], ShouldEqual, "parentPropertyID")
			})
		})
	})
}
