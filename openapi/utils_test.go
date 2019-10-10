package openapi

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestExpandPath(t *testing.T) {
	Convey("Given a file with absolute path", t, func() {
		expectedPath := "/Users/username/.terraform/plugins"
		Convey("When expandPath is called", func() {
			path, err := expandPath(expectedPath)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the path should be the same as the input", func() {
				So(path, ShouldEqual, path)
			})
		})
	})
	Convey("Given a file starting with ~", t, func() {
		homePath := "~/some_folder"
		Convey("When expandPath is called", func() {
			path, err := expandPath(homePath)
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the path returned should be expanded with the home dir", func() {
				// Getting home dir to make the test OS-agnostic
				homeDir, _ := homedir.Dir()
				So(path, ShouldEqual, fmt.Sprintf("%s/%s", homeDir, "some_folder"))
			})
		})
	})
}

func TestGetFileContent(t *testing.T) {
	Convey("Given a file", t, func() {
		expectedContent := "some content"
		f, err := createTmpFile(expectedContent)
		defer os.Remove(f.Name())
		So(err, ShouldBeNil)
		Convey("When getFileContent is called", func() {
			content, err := getFileContent(f.Name())
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the content should be the expected one", func() {
				So(content, ShouldEqual, expectedContent)
			})
		})
	})
}

func TestIsUrl(t *testing.T) {

	testCases := []struct {
		name           string
		input          string
		expectedResult bool
	}{
		{
			name:           "url well formed",
			input:          "http://something.com",
			expectedResult: true,
		},
		{
			name:           "url with path",
			input:          "http://something.com/something",
			expectedResult: true,
		},
		{
			name:           "url with no protocol",
			input:          "something.com",
			expectedResult: false,
		},
	}

	for _, tc := range testCases {
		isURL := isURL(tc.input)
		assert.Equal(t, tc.expectedResult, isURL, tc.name)
	}
}
