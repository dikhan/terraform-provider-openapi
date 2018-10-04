package openapiutils

//
import (
	"fmt"
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
