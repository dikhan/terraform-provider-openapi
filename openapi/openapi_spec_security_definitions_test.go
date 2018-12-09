package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFindSecurityDefinitionFor(t *testing.T) {
	Convey("Given a SpecSecurityDefinitions", t, func() {
		expectedSecDefName := "secDefName"
		s := SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition(expectedSecDefName, "Authorization"),
		}
		Convey("When findSecurityDefinitionFor method is called with an existing sec def name", func() {
			secDef := s.findSecurityDefinitionFor(expectedSecDefName)
			Convey("Then the secDef result should not be nil", func() {
				So(secDef, ShouldNotBeNil)
			})
			Convey("And the secDef should match the expected one", func() {
				So(secDef.getName(), ShouldEqual, expectedSecDefName)
			})
		})
		Convey("When findSecurityDefinitionFor method is called with a NON existing sec def name", func() {
			secDef := s.findSecurityDefinitionFor("nonExistingSecDefName")
			Convey("Then the secDef result should be nil", func() {
				So(secDef, ShouldBeNil)
			})
		})
	})

}
