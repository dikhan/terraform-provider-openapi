package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCreateAPIKeyAuthenticator(t *testing.T) {
	Convey("Given a secDef of header type and a auth value ", t, func() {
		secDef := newAPIKeyHeaderSecurityDefinition("header_auth", authorizationHeader)
		value := "value"
		Convey("When createAPIKeyAuthenticator method is constructed", func() {
			apiKeyAuthenticator := createAPIKeyAuthenticator(secDef, value)
			Convey("And the the specAPIKeyAuthenticator returned Should Have Same Type As apiKeyHeaderAuthenticator", func() {
				So(apiKeyAuthenticator, ShouldHaveSameTypeAs, apiKeyHeaderAuthenticator{})
			})
			Convey("And the the specAPIKeyAuthenticator returned should be of type authTypeAPIKeyHeader", func() {
				So(apiKeyAuthenticator.getType(), ShouldEqual, authTypeAPIKeyHeader)
			})
		})
	})

	Convey("Given a secDef of query type and a auth value ", t, func() {
		secDef := newAPIKeyQuerySecurityDefinition("query_auth", authorizationHeader)
		value := "value"
		Convey("When createAPIKeyAuthenticator method is constructed", func() {
			apiKeyAuthenticator := createAPIKeyAuthenticator(secDef, value)
			Convey("And the the specAPIKeyAuthenticator returned Should Have Same Type As apiKeyQueryAuthenticator", func() {
				So(apiKeyAuthenticator, ShouldHaveSameTypeAs, apiKeyQueryAuthenticator{})
			})
			Convey("And the the specAPIKeyAuthenticator returned should be of type authTypeAPIQuery", func() {
				So(apiKeyAuthenticator.getType(), ShouldEqual, authTypeAPIQuery)
			})
		})
	})
}
