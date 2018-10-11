package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNewAPIKey(t *testing.T) {
	Convey("Given a name and an apiKeyIn", t, func() {
		name := "api_header_auth"
		apiKeyIn := inHeader
		Convey("When newAPIKey method is called", func() {
			specAPIKey := newAPIKey(name, apiKeyIn)
			Convey("Then the specAPIKey name should match", func() {
				So(specAPIKey.Name, ShouldEqual, name)
			})
			Convey("And the specAPIKey type should be apiKey", func() {
				So(specAPIKey.In, ShouldEqual, apiKeyIn)
			})
		})
	})
}

func TestNewAPIKeyQuery(t *testing.T) {
	Convey("Given a name", t, func() {
		name := "api_header_auth"
		Convey("When newAPIKeyQuery method is called", func() {
			specAPIKey := newAPIKeyQuery(name)
			Convey("Then the specAPIKey name should match", func() {
				So(specAPIKey.Name, ShouldEqual, name)
			})
			Convey("And the specAPIKey IN value should be inQuery", func() {
				So(specAPIKey.In, ShouldEqual, inQuery)
			})
		})
	})
}

func TestNewAPIKeyHeader(t *testing.T) {
	Convey("Given a name", t, func() {
		name := "api_header_auth"
		Convey("When newAPIKeyHeader method is called", func() {
			specAPIKey := newAPIKeyHeader(name)
			Convey("Then the specAPIKey name should match", func() {
				So(specAPIKey.Name, ShouldEqual, name)
			})
			Convey("And the specAPIKey IN value should be inQuery", func() {
				So(specAPIKey.In, ShouldEqual, inHeader)
			})
		})
	})
}