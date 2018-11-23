package openapi

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"testing"
)

func TestServiceSchemaConfigurationV1(t *testing.T) {
	Convey("Given a SchemaPropertyName, DefaultValue and some ExternalConfiguration ", t, func() {
		schemaPropertyName := "schemaPropertyName"
		defaultValue := "defaultValue"
		externalConfiguration := ServiceSchemaPropertyExternalConfigurationV1{
			KeyName:     "someKeyName",
			ContentType: "raw",
			File:        "/path/to/credentials",
		}
		Convey("When a new ServiceSchemaPropertyConfigurationV1 is created", func() {
			serviceSchemaConfigurationV1 := ServiceSchemaPropertyConfigurationV1{
				SchemaPropertyName:    schemaPropertyName,
				DefaultValue:          defaultValue,
				ExternalConfiguration: externalConfiguration,
			}
			Convey("And the serviceSchemaConfigurationV1 created should implement ServiceSchemaPropertyConfiguration interface", func() {
				var _ ServiceSchemaPropertyConfiguration = serviceSchemaConfigurationV1
			})
		})
	})
}

func TestServiceSchemaConfigurationV1GetDefaultValue(t *testing.T) {
	Convey("Given a ServiceSchemaPropertyConfigurationV1 with just a default value and no external config", t, func() {
		serviceSchemaConfigurationV1 := ServiceSchemaPropertyConfigurationV1{
			SchemaPropertyName: "schemaPropertyName",
			DefaultValue:       "defaultValue",
		}
		Convey("When GetDefaultValue method is called", func() {
			value, err := serviceSchemaConfigurationV1.GetDefaultValue()
			Convey("And the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should matched the default one", func() {
				So(value, ShouldEqual, serviceSchemaConfigurationV1.DefaultValue)
			})
		})
	})

	Convey("Given an external config file and a ServiceSchemaPropertyConfigurationV1 with a default value and external 'raw' config", t, func() {
		expectedValue := "some content"
		tmpFile, err := ioutil.TempFile("", "")
		So(err, ShouldBeNil)
		tmpFile.Write([]byte(expectedValue))
		serviceSchemaConfigurationV1 := ServiceSchemaPropertyConfigurationV1{
			SchemaPropertyName: "schemaPropertyName",
			DefaultValue:       "defaultValue",
			ExternalConfiguration: ServiceSchemaPropertyExternalConfigurationV1{
				//KeyName:     "someKeyName",
				ContentType: "raw",
				File:        tmpFile.Name(),
			},
		}
		Convey("When GetDefaultValue method is called", func() {
			value, err := serviceSchemaConfigurationV1.GetDefaultValue()
			Convey("And the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should matched the default one", func() {
				So(value, ShouldEqual, expectedValue)
			})
		})
	})

	Convey("Given an external config file and a ServiceSchemaPropertyConfigurationV1 with a default value and external 'json' config", t, func() {
		expectedValue := "someName"
		tmpFile, err := ioutil.TempFile("", "")
		defer os.Remove(tmpFile.Name())
		So(err, ShouldBeNil)
		tmpFile.Write([]byte(fmt.Sprintf(`{"firstName":"%s"}`, expectedValue)))
		serviceSchemaConfigurationV1 := ServiceSchemaPropertyConfigurationV1{
			SchemaPropertyName: "schemaPropertyName",
			DefaultValue:       "defaultValue",
			ExternalConfiguration: ServiceSchemaPropertyExternalConfigurationV1{
				KeyName:     "$.firstName",
				ContentType: "json",
				File:        tmpFile.Name(),
			},
		}
		Convey("When GetDefaultValue method is called", func() {
			value, err := serviceSchemaConfigurationV1.GetDefaultValue()
			Convey("And the err returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should matched the default one", func() {
				So(value, ShouldEqual, expectedValue)
			})
		})
	})

	Convey("Given an external config file and a ServiceSchemaPropertyConfigurationV1 with an external configuration's non supported content type", t, func() {
		expectedValue := "some content"
		tmpFile, err := ioutil.TempFile("", "")
		So(err, ShouldBeNil)
		tmpFile.Write([]byte(expectedValue))
		serviceSchemaConfigurationV1 := ServiceSchemaPropertyConfigurationV1{
			SchemaPropertyName: "schemaPropertyName",
			DefaultValue:       "defaultValue",
			ExternalConfiguration: ServiceSchemaPropertyExternalConfigurationV1{
				ContentType: "nonSupported",
				File:        tmpFile.Name(),
			},
		}
		Convey("When GetDefaultValue method is called", func() {
			_, err := serviceSchemaConfigurationV1.GetDefaultValue()
			Convey("And the err returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the err message should be", func() {
				So(err.Error(), ShouldContainSubstring, "'schemaPropertyName': 'nonSupported' content type not supported")
			})
		})
	})
}

func TestServiceExternalConfigurationV1GetFileParser(t *testing.T) {
	Convey("Given a ServiceSchemaPropertyExternalConfigurationV1 configured with 'raw' content", t, func() {
		expectedValue := "some content"
		tmpFile, err := ioutil.TempFile("", "")
		defer os.Remove(tmpFile.Name())
		So(err, ShouldBeNil)
		tmpFile.Write([]byte(expectedValue))
		serviceExternalConfigurationV1 := ServiceSchemaPropertyExternalConfigurationV1{
			ContentType: "raw",
			File:        tmpFile.Name(),
		}
		Convey("When getFileParser method is called with some content", func() {
			parser, err := serviceExternalConfigurationV1.getFileParser()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the parser returned should be of type parserRaw", func() {
				So(parser, ShouldHaveSameTypeAs, parserRaw{})
			})
			Convey("And the parser get value should return the right value", func() {
				value, err := parser.getValue()
				So(err, ShouldBeNil)
				So(value, ShouldEqual, expectedValue)
			})
		})
	})

	Convey("Given a ServiceSchemaPropertyExternalConfigurationV1 configured with 'json' content", t, func() {
		expectedValue := "someName"
		tmpFile, err := ioutil.TempFile("", "")
		tmpFile.Write([]byte(fmt.Sprintf(`{"firstName":"%s"}`, expectedValue)))
		defer os.Remove(tmpFile.Name())
		So(err, ShouldBeNil)
		serviceExternalConfigurationV1 := ServiceSchemaPropertyExternalConfigurationV1{
			ContentType: "json",
			File:        tmpFile.Name(),
			KeyName:     "$.firstName",
		}
		Convey("When getFileParser method is called with some content", func() {
			parser, err := serviceExternalConfigurationV1.getFileParser()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the parser returned should be of type parserRaw", func() {
				So(parser, ShouldHaveSameTypeAs, parserJSON{})
			})
			Convey("And the parser get value should return the right value", func() {
				value, err := parser.getValue()
				So(err, ShouldBeNil)
				So(value, ShouldEqual, expectedValue)
			})
		})
	})

	Convey("Given a ServiceSchemaPropertyExternalConfigurationV1 configured with a non supported content", t, func() {
		expectedValue := "some content"
		tmpFile, err := ioutil.TempFile("", "")
		defer os.Remove(tmpFile.Name())
		So(err, ShouldBeNil)
		tmpFile.Write([]byte(expectedValue))
		serviceExternalConfigurationV1 := ServiceSchemaPropertyExternalConfigurationV1{
			ContentType: "nonSupported",
			File:        tmpFile.Name(),
		}
		Convey("When getFileParser method is called with some content", func() {
			_, err := serviceExternalConfigurationV1.getFileParser()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the err message should be", func() {
				So(err.Error(), ShouldContainSubstring, "'nonSupported' content type not supported")
			})
		})
	})
}

func TestSchemaFileParserRaw(t *testing.T) {
	Convey("Given a content", t, func() {
		content := "someContent"
		Convey("When a new parserRaw is created", func() {
			parserRaw := parserRaw{
				content: content,
			}
			Convey("And the parserRaw created should implement schemaFileParser interface", func() {
				var _ schemaFileParser = parserRaw
			})
		})
	})
}

func TestSchemaFileParserRawGetValue(t *testing.T) {
	Convey("Given a content", t, func() {
		expectedContent := "some content"
		parserRaw := parserRaw{
			content: expectedContent,
		}
		Convey("When getValue method is called", func() {
			value, err := parserRaw.getValue()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should match the content", func() {
				So(value, ShouldEqual, expectedContent)
			})
		})
	})
}

func TestSchemaFileParserJSON(t *testing.T) {
	Convey("Given a content", t, func() {
		jsonContent := "{}"
		keyName := "$.firstName"
		Convey("When a new parserRaw is created", func() {
			parserJSON := parserJSON{
				jsonContent: jsonContent,
				keyName:     keyName,
			}
			Convey("And the parserJSON created should implement schemaFileParser interface", func() {
				var _ schemaFileParser = parserJSON
			})
		})
	})
}

func TestSchemaFileParserJsonGetValue(t *testing.T) {
	Convey("Given a jsonContent and an existing key name", t, func() {
		expectedValue := "someName"
		jsonContent := fmt.Sprintf(`{"firstName":"%s"}`, expectedValue)
		keyName := "$.firstName"
		parserJSON := parserJSON{
			jsonContent: jsonContent,
			keyName:     keyName,
		}
		Convey("When getValue method is called", func() {
			value, err := parserJSON.getValue()
			Convey("Then the error returned should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("And the value returned should match the content", func() {
				So(value, ShouldEqual, expectedValue)
			})
		})
	})

	Convey("Given a content and NON existing key name", t, func() {
		expectedValue := "someName"
		jsonContent := fmt.Sprintf(`{"firstName":"%s"}`, expectedValue)
		keyName := "$.someNonExistingKey"
		parserJSON := parserJSON{
			jsonContent: jsonContent,
			keyName:     keyName,
		}
		Convey("When getValue method is called", func() {
			_, err := parserJSON.getValue()
			Convey("Then the error returned should NOT be nil", func() {
				So(err, ShouldNotBeNil)
			})
			Convey("And the error returned should be", func() {
				So(err.Error(), ShouldEqual, "key error: someNonExistingKey not found in object")
			})
		})
	})
}
