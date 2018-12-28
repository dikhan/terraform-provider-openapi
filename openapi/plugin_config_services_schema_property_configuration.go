package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/oliveagle/jsonpath"
	"log"
	"os/exec"
	"time"
)

// ServiceSchemaPropertyConfiguration defines the behaviour expected for the service schema property configuration
type ServiceSchemaPropertyConfiguration interface {
	GetDefaultValue() (string, error)
	ExecuteCommand() error
}

const cmdTimeout = 10

// ServiceSchemaPropertyConfigurationV1 implements the ServiceSchemaPropertyConfiguration and defines the different
// fields supported that enable the configuration of the provider's properties via the terraform-provider-openapi.yaml plugin
// config file
type ServiceSchemaPropertyConfigurationV1 struct {
	SchemaPropertyName    string                                       `yaml:"schema_property_name"`
	DefaultValue          string                                       `yaml:"default_value"`
	Command               []string                                     `yaml:"cmd"`
	CommandTimeout        int                                          `yaml:"cmd_timeout"`
	ExternalConfiguration ServiceSchemaPropertyExternalConfigurationV1 `yaml:"schema_property_external_configuration"`
}

// ServiceSchemaPropertyExternalConfigurationV1 defines the external configuration for a provider property.
type ServiceSchemaPropertyExternalConfigurationV1 struct {
	// File defines the file containing the value of the schema property
	File string `yaml:"file"`
	// KeyName defines the specific key to look for within the File (only when json content type)
	KeyName string `yaml:"key_name"`
	// ContentType defines the type of content the File has
	ContentType string `yaml:"content_type"` // Currently supported types: raw, json
}

// GetDefaultValue returns the default value for the schema property configuration. The following logic defines the preference
// when deciding what should be the default value of the property:
// - if the property does not have external configuration ('schema_property_external_configuration') and it does have a 'default_value' is set, then value used will be the one specified in the 'default_value' field
// - if the property has both the external configuration ('schema_property_external_configuration') and the 'default_value' fields set:
//    - If 'file' field is populated then:
//      - If the 'content_type' is raw the contents of the 'file' will be used as default value
//      - If the 'content_type' is json then the content of the 'file' must be json structure and the default value used will be the one defined in the 'key_name'
//    - An error is thrown otherwise
func (s ServiceSchemaPropertyConfigurationV1) GetDefaultValue() (string, error) {
	if &s.ExternalConfiguration != nil {
		if s.ExternalConfiguration.File != "" {
			log.Printf("[DEBUG] provider schema property '%s' configured to use as default value [ContentType=%s; File=%s, KeyName=%s]", s.SchemaPropertyName, s.ExternalConfiguration.ContentType, s.ExternalConfiguration.File, s.ExternalConfiguration.KeyName)
			schemaFileParser, err := s.ExternalConfiguration.getFileParser()
			if err != nil {
				return "", fmt.Errorf("failed to read external configuration file '%s' for schema property '%s': %s", s.ExternalConfiguration.File, s.SchemaPropertyName, err)
			}
			defaultValue, err := schemaFileParser.getValue()
			if err != nil {
				return "", err
			}
			return defaultValue, nil
		}
	}
	return s.DefaultValue, nil
}

// ExecuteCommand run the 'Command' configured in the ServiceSchemaPropertyConfigurationV1 struct if applicable.
// - If the command fails to execute the appropriate error will be returned including the error returned by exec
// - If the command execution does not finish within the expected time (either before CommandTimeout or before the default timeout 10s)
// a timeout error will be returned
// - Otherwise, a nil error will be returned should the command executes successfully with a clean exit code
func (s ServiceSchemaPropertyConfigurationV1) ExecuteCommand() error {
	doneChan := make(chan error)
	// execute the command in a routine and wait for completion
	go s.exec(doneChan)
	err := <-doneChan
	if err != nil {
		return err
	}
	return nil
}

func (s ServiceSchemaPropertyConfigurationV1) exec(doneChan chan error) {
	if len(s.Command) > 0 {
		start := time.Now()
		log.Printf("[INFO] executing '%s' command '%s'", s.SchemaPropertyName, s.Command)

		timeout := cmdTimeout
		if s.CommandTimeout > 0 {
			timeout = s.CommandTimeout
		}

		// Create a new context and add a timeout to it
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
		defer cancel() // The cancel should be deferred so resources are cleaned up

		// Create the command with our context
		cmd := exec.CommandContext(ctx, s.Command[0], s.Command[1:]...)

		// Capture stdout and stderr
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		// We want to check the context error to see if the timeout was executed. The error returned by cmd.Output() will be OS specific based on what
		// happens when a process is killed.
		if ctx.Err() == context.DeadlineExceeded {
			doneChan <- fmt.Errorf("command '%s' did not finish executing within the expected time %ds (%s)", s.Command, timeout, err)
			return
		}

		// If there's no context error, we know the command completed (or errored).
		if err != nil {
			doneChan <- fmt.Errorf("failed to execute '%s' command '%s': %s(%s)", s.SchemaPropertyName, s.Command, stderr.String(), err)
			return
		}
		log.Printf("[INFO] provider schema property '%s' command '%s' executed successfully (time:%s): %s", s.SchemaPropertyName, s.Command, time.Since(start), stdout.String())
	}
	doneChan <- nil
}

func (c ServiceSchemaPropertyExternalConfigurationV1) getFileParser() (schemaFileParser, error) {
	schemaFileContent, err := getFileContent(c.File)
	if err != nil {
		return nil, err
	}
	switch c.ContentType {
	case "raw":
		if c.KeyName != "" {
			log.Printf("[WARN] service external configuration of type 'raw' configured with key value '%s'", c.KeyName)
		}
		return parserRaw{content: schemaFileContent}, nil
	case "json":
		return parserJSON{jsonContent: schemaFileContent, keyName: c.KeyName}, nil
	default:
		return nil, fmt.Errorf("'%s' content type not supported", c.ContentType)
	}
}

type schemaFileParser interface {
	getValue() (string, error)
}

type parserRaw struct {
	content string
}

func (p parserRaw) getValue() (string, error) {
	return p.content, nil
}

type parserJSON struct {
	jsonContent string
	keyName     string
}

func (p parserJSON) getValue() (string, error) {
	var jsonData interface{}
	json.Unmarshal([]byte(p.jsonContent), &jsonData)
	res, err := jsonpath.JsonPathLookup(jsonData, p.keyName)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}
