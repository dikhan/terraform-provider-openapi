package openapi

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func Test_getParentPropertiesNames(t *testing.T) {
	Convey("Given an empty subResourceInfo", t, func() {
		s := &subResourceInfo{}
		Convey("When the method getParentPropertiesNames is called", func() {
			p := s.getParentPropertiesNames()
			Convey("Then array returned should be empty", func() {
				So(p, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a subResourceInfo with empty parentResourceNames", t, func() {
		s := &subResourceInfo{
			parentResourceNames: []string{},
		}
		Convey("When the method getParentPropertiesNames is called", func() {
			p := s.getParentPropertiesNames()
			Convey("Then array returned should be empty", func() {
				So(p, ShouldBeEmpty)
			})
		})
	})

	Convey("Given a subResourceInfo with some parentResourceNames", t, func() {
		s := &subResourceInfo{
			parentResourceNames: []string{"cdn_v1", "cdn_v1_firewalls_v2"},
		}
		Convey("When the method getParentPropertiesNames is called", func() {
			p := s.getParentPropertiesNames()
			Convey("And the array returned should contain the expected parent names including the id postfix", func() {
				So(len(p), ShouldEqual, 2)
				So(p[0], ShouldEqual, "cdn_v1_id")
				So(p[1], ShouldEqual, "cdn_v1_firewalls_v2_id")
			})
		})
	})
}
