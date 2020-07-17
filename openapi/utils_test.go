package openapi

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestExpandPath(t *testing.T) {
	// Getting home dir to make the test OS-agnostic
	homeDir, _ := homedir.Dir()

	testCases := []struct {
		name           string
		inputPath      string
		expectedResult string
		expectedError  error
	}{
		{name: "file with absolute path", inputPath: "/Users/username/.terraform/plugins", expectedResult: "/Users/username/.terraform/plugins", expectedError: nil},
		{name: "file starting with ~", inputPath: "~/some_folder", expectedResult: fmt.Sprintf("%s/%s", homeDir, "some_folder"), expectedError: nil},
	}

	for _, tc := range testCases {
		Convey(fmt.Sprintf("When expandPath method is called: %s", tc.name), t, func() {
			returnedPath, err := expandPath(tc.inputPath)
			Convey("Then the result returned should be the expected one", func() {
				So(err, ShouldResemble, tc.expectedError)
				So(returnedPath, ShouldEqual, tc.expectedResult)
			})
		})
	}
}

func TestGetFileContent(t *testing.T) {
	Convey("Given a file", t, func() {
		expectedContent := "some content"
		f, err := ioutil.TempFile("", "")
		So(err, ShouldBeNil)
		_, err = f.Write([]byte(expectedContent))
		So(err, ShouldBeNil)
		defer func() {
			err := os.Remove(f.Name())
			So(err, ShouldBeNil)
		}()
		So(err, ShouldBeNil)
		Convey("When getFileContent is called", func() {
			content, err := getFileContent(f.Name())
			Convey("Then the content returned should match the expected one and there should be no error", func() {
				So(err, ShouldBeNil)
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
