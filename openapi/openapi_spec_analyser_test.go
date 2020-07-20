package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"testing"
)

func TestCreateSpecAnalyser(t *testing.T) {
	Convey("Given a specAnalyserVersion and a openAPIDocumentURL", t, func() {
		specAnalyserVersion := specAnalyserV2

		file := initAPISpecFile(`swagger: "2.0"`)
		defer os.Remove(file.Name())

		openAPIDocumentURL := file.Name()
		Convey("When CreateSpecAnalyser method is called", func() {
			specAnalyser, err := CreateSpecAnalyser(specAnalyserVersion, openAPIDocumentURL)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldBeNil)
				So(specAnalyser, ShouldHaveSameTypeAs, &specV2Analyser{})
			})
		})

		Convey("When CreateSpecAnalyser method is called with a non valid openAPIDocumentURL", func() {
			_, err := CreateSpecAnalyser(specAnalyserVersion, "some non valid spec file")
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "failed to retrieve the OpenAPI document from 'some non valid spec file' - error = open some non valid spec file: no such file or directory")
			})
		})

		Convey("When CreateSpecAnalyser method is called with a non supported version", func() {
			_, err := CreateSpecAnalyser("nonSupportedVersion", openAPIDocumentURL)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "open api spec analyser version 'nonSupportedVersion' not supported, please choose a valid SpecAnalyser implementation [v2]")
			})
		})
	})
}
