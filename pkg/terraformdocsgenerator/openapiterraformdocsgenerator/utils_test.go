package openapiterraformdocsgenerator

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRender(t *testing.T) {

	testCases := []struct {
		name          string
		template      string
		data          interface{}
		expectedError string
	}{
		{
			name:          "template renders just fine",
			template:      `Some template {{printf "%d" 23}}`,
			expectedError: "",
		},
		{
			name:          "forcing error when executing template",
			template:      `{{template "not_existing" .}}`,
			expectedError: "no such template",
		},
		{
			name:          "forcing error when parsing template",
			template:      `{{with $v, $u := 3}}{{end}}`,
			expectedError: "too many declarations in with",
		},
	}

	for _, tc := range testCases {
		var output bytes.Buffer
		err := render(&output, "templateName", tc.template, tc.data)
		if tc.expectedError == "" {
			assert.Nil(t, err)
		} else {
			assert.Contains(t, err.Error(), tc.expectedError)
		}
	}

}
