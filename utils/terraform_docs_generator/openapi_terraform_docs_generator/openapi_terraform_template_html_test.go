package openapi_terraform_docs_generator

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"html/template"
	"io"
	"testing"
)

func TestArgumentReferenceTmpl_Required_String(t *testing.T) {
	var output bytes.Buffer
	property := Property{
		Name:        "string_property",
		Type:        "string",
		Required:    true,
		Description: "some description",
	}

	tmpl := fmt.Sprintf(`%s
{{- template "resource_argument_reference" .}}`, ArgumentReferenceTmpl)

	renderTest(t, &output, "ArgumentReference", tmpl, property)
	expectedOutput := "<li> string_property [string]  - (Required) some description</li>\n\t"
	assert.Equal(t, expectedOutput, output.String())
}

func renderTest(t *testing.T, w io.Writer, templateName string, templateContent string, data interface{}) {
	tmpl, err := template.New(templateName).Parse(templateContent)
	assert.Nil(t, err)
	err = tmpl.Execute(w, data)
	assert.Nil(t, err)
}
