package openapiterraformdocsgenerator

import (
	"html/template"
	"io"
)

func render(w io.Writer, templateName string, templateContent string, data interface{}) error {
	tmpl, err := template.New(templateName).Parse(templateContent)
	if err != nil {
		return err
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		return err
	}
	return nil
}
