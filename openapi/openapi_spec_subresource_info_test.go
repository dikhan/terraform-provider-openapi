package openapi

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_getParentPropertiesNames(t *testing.T) {
	Convey("Given an empty ParentResourceInfo", t, func() {
		s := &ParentResourceInfo{}
		Convey("When the method GetParentPropertiesNames is called", func() {
			p := s.GetParentPropertiesNames()
			Convey("Then array returned should be empty", func() {
				So(p, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a ParentResourceInfo with empty parentResourceNames", t, func() {
		s := &ParentResourceInfo{
			parentResourceNames: []string{},
		}
		Convey("When the method GetParentPropertiesNames is called", func() {
			p := s.GetParentPropertiesNames()
			Convey("Then array returned should be empty", func() {
				So(p, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a ParentResourceInfo with some parentResourceNames", t, func() {
		s := &ParentResourceInfo{
			parentResourceNames: []string{"cdn_v1", "cdn_v1_firewalls_v2"},
		}
		Convey("When the method GetParentPropertiesNames is called", func() {
			p := s.GetParentPropertiesNames()
			Convey("And the array returned should contain the expected parent names including the id postfix", func() {
				So(len(p), ShouldEqual, 2)
				So(p[0], ShouldEqual, "cdn_v1_id")
				So(p[1], ShouldEqual, "cdn_v1_firewalls_v2_id")
			})
		})
	})
}
