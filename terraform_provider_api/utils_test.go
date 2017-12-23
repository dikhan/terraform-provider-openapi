package main

import (
	"testing"

	"net/http"

	. "github.com/smartystreets/goconvey/convey"
)

func TestResponseContainsExpectedStatus(t *testing.T) {
	Convey("Given we have a list of expected response status codes and a response code", t, func() {
		expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
		responseCode := http.StatusCreated
		Convey("When responseContainsExpectedStatus is called with a response code that exists in 'expectedResponseStatusCodes'", func() {
			exists := responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the value returned should be true", func() {
				So(exists, ShouldBeTrue)
			})
		})
	})

	Convey("Given we have a list of expected response status codes and a response code", t, func() {
		expectedResponseStatusCodes := []int{http.StatusCreated, http.StatusAccepted}
		responseCode := http.StatusUnauthorized
		Convey("When responseContainsExpectedStatus is called with a response code that DOES NOT exists in 'expectedResponseStatusCodes'", func() {
			exists := responseContainsExpectedStatus(expectedResponseStatusCodes, responseCode)
			Convey("Then the value returned should be false", func() {
				So(exists, ShouldBeFalse)
			})
		})
	})
}
