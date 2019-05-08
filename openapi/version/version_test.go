package version

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)


func TestBuildUserAgent(t *testing.T) {
	Convey("Given a version and a commit hash", t, func() {
		Version = "someVersion"
		Commit = "someCommit"
		Convey("When BuildUserAgent method is called with some runtime", func() {
			runtime := "linux"
			arch := "amd64"
			value := BuildUserAgent(runtime, arch)
			Convey("Then the value of the header should be the expected one", func() {
				So(value, ShouldEqual, "OpenAPI Terraform Provider/someVersion-someCommit (linux/amd64)")
			})
		})
	})
}
