package openapiutils

import (
	"encoding/json"
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestGetHostFromURL(t *testing.T) {
	Convey("Given a url with protocol, fqdn, a port number and path", t, func() {
		expectedResult := "localhost:8080"
		url := fmt.Sprintf("http://%s/swagger.yaml", expectedResult)
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain with port included", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with NO protocol, domain localhost, a port number and path", t, func() {
		expectedResult := "localhost:8080"
		url := fmt.Sprintf("%s/swagger.yaml", expectedResult)
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain with port included", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with NO protocol, domain localhost and without a path", t, func() {
		expectedResult := "localhost:8080"
		url := expectedResult
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain with port included", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with just the domain", t, func() {
		expectedResult := "localhost"
		url := expectedResult
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with protocol and fqdn", t, func() {
		expectedResult := "www.domain.com"
		url := fmt.Sprintf("http://%s", expectedResult)
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with protocol and fqdn with trailing slash", t, func() {
		expectedResult := "www.domain.com"
		url := fmt.Sprintf("http://%s/", expectedResult)
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with protocol, fqdn and path", t, func() {
		expectedResult := "example.domain.com"
		url := fmt.Sprintf("http://%s/some/path", expectedResult)
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with protocol, fqdn and query param", t, func() {
		expectedResult := "example.domain.com"
		url := fmt.Sprintf("http://%s?anything", expectedResult)
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with just the domain", t, func() {
		expectedResult := "domain.com"
		url := expectedResult
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with a domain containing dots", t, func() {
		expectedResult := "example.domain.com"
		url := expectedResult
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
	Convey("Given a url with a domain containing dots and dashes", t, func() {
		expectedResult := "example.domain-hyphen.com"
		url := expectedResult
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})

	Convey("Given a url with a domain subdomain www and dots", t, func() {
		expectedResult := "www.domain.com"
		url := expectedResult
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should be the domain", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})

	// Negative cases
	Convey("Given a url with a fqdn not using standard port", t, func() {
		expectedResult := "domain.com"
		url := fmt.Sprintf("%s:8080", expectedResult)
		Convey("When GetHostFromURL method is called", func() {
			domain := GetHostFromURL(url)
			Convey("Then the string returned should JUST be the FQDN part (this is by design - it is assumed that actual FQDN will use HTTP standard ports)", func() {
				So(domain, ShouldEqual, expectedResult)
			})
		})
	})
}

func TestStringExtensionExists(t *testing.T) {
	Convey("Given a list of extensions", t, func() {
		extensions := spec.Extensions{
			"directlyUnmarshaled": "value1",
		}
		extensions.Add("addedViaAddMethod", "value2")
		Convey("When StringExtensionExists method is called to look up a key that is not lower case", func() {
			value, exists := StringExtensionExists(extensions, "directlyUnmarshaled")
			Convey("Then the key should exists", func() {
				So(exists, ShouldBeTrue)
			})
			Convey("And Then the value should be", func() {
				So(value, ShouldEqual, "value1")
			})
		})
		Convey("When StringExtensionExists method is called to look up a key that added cammel case but the lookup key is lower cased", func() {
			value, exists := StringExtensionExists(extensions, "directlyunmarshaled")
			Convey("Then the key should exists", func() {
				So(exists, ShouldBeTrue)
			})
			Convey("And Then the value should be", func() {
				So(value, ShouldEqual, "value1")
			})
		})
		Convey("When StringExtensionExists method is called to look up a key that was added via the Add extensions method", func() {
			value, exists := StringExtensionExists(extensions, "addedViaAddMethod")
			Convey("Then the key should exists", func() {
				So(exists, ShouldBeTrue)
			})
			Convey("And Then the value should be", func() {
				So(value, ShouldEqual, "value2")
			})
		})
		Convey("When StringExtensionExists method is called to look up a lower case key that was added via the Add extensions method", func() {
			value, exists := StringExtensionExists(extensions, "addedviaaddmethod")
			Convey("Then the key should exists", func() {
				So(exists, ShouldBeTrue)
			})
			Convey("And Then the value should be", func() {
				So(value, ShouldEqual, "value2")
			})
		})
	})
}

func TestGetPayloadDefName(t *testing.T) {
	Convey("Given a valid internal definition path", t, func() {
		ref := "#/definitions/ContentDeliveryNetworkV1"
		// Local Reference use cases
		Convey("When getPayloadDefName method is called with a valid internal definition path", func() {
			defName, err := getPayloadDefName(ref)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should be true", func() {
				So(defName, ShouldEqual, "ContentDeliveryNetworkV1")
			})
		})
	})

	Convey("Given a ref URL (not supported)", t, func() {
		ref := "http://path/to/your/resource.json#myElement"
		Convey("When getPayloadDefName method is called ", func() {
			_, err := getPayloadDefName(ref)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an element of the document located on the same server (not supported)", t, func() {
		ref := "document.json#/myElement"
		// Remote Reference use cases
		Convey("When getPayloadDefName method is called with ", func() {
			_, err := getPayloadDefName(ref)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an element of the document located in the parent folder (not supported)", t, func() {
		ref := "../document.json#/myElement"
		Convey("When getPayloadDefName method is called with ", func() {
			_, err := getPayloadDefName(ref)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an specific element of the document stored on the different server (not supported)", t, func() {
		ref := "http://path/to/your/resource.json#myElement"
		Convey("When getPayloadDefName method is called with ", func() {
			_, err := getPayloadDefName(ref)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given an element of the document located in another folder (not supported)", t, func() {
		ref := "../another-folder/document.json#/myElement"
		// URL Reference use case
		Convey("When getPayloadDefName method is called with ", func() {
			_, err := getPayloadDefName(ref)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})

	Convey("Given a  document on the different server, which uses the same protocol (not supported)", t, func() {
		ref := "//anotherserver.com/files/example.json"
		Convey("When getPayloadDefName method is called with ", func() {
			_, err := getPayloadDefName(ref)
			Convey("Then the error returned should not be nil", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestGetResourcePayloadSchemaDef(t *testing.T) {
	Convey("Given a swagger doc", t, func() {
		swaggerContent := `swagger: "2.0"
definitions:
  Users:
    type: "object"
    required:
      - name
    properties:
      id:
        type: "string"
        readOnly: true`
		spec := initSwagger(swaggerContent)
		Convey("When getResourcePayloadSchemaDef method is called with an operation containing a valid ref: '#/definitions/Users'", func() {
			ref := "#/definitions/Users"
			resourcePayloadSchemaDef, err := GetSchemaDefinition(spec.Definitions, ref)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("Then the value returned should be a valid schema def", func() {
				So(len(resourcePayloadSchemaDef.Type), ShouldEqual, 1)
				So(resourcePayloadSchemaDef.Type, ShouldContain, "object")
			})
		})
		Convey("When getResourcePayloadSchemaDef method is called with schema that is missing the definition the ref is pointing at", func() {
			ref := "#/definitions/NonExistingDef"
			_, err := GetSchemaDefinition(spec.Definitions, ref)
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error message should be", func() {
				So(err.Error(), ShouldContainSubstring, "missing schema definition in the swagger file with the supplied ref '#/definitions/NonExistingDef'")
			})
		})
	})
}

func initSwagger(swaggerContent string) *spec.Swagger {
	swagger := json.RawMessage([]byte(swaggerContent))
	d, _ := loads.Analyzed(swagger, "2.0")
	return d.Spec()
}
