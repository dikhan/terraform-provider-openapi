package openapi

import (
	"fmt"

	"log"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
)

const extTfImmutable = "x-terraform-immutable"
const extTfForceNew = "x-terraform-force-new"
const extTfSensitive = "x-terraform-sensitive"
const extTfID = "x-terraform-id"

type resourcesInfo map[string]resourceInfo

// resourceInfo serves as translator between swagger definitions and terraform schemas
type resourceInfo struct {
	name     string
	basePath string
	// path contains relative path to the resource e,g: /v1/resource
	path             string
	host             string
	httpSchemes      []string
	schemaDefinition spec.Schema
	// createPathInfo contains info about /resource, including the POST operation
	createPathInfo spec.PathItem
	// pathInfo contains info about /resource/{id}, including GET, PUT and REMOVE operations if applicable
	pathInfo spec.PathItem
}

func (r resourceInfo) createTerraformResourceSchema() (map[string]*schema.Schema, error) {
	s := map[string]*schema.Schema{}
	for propertyName, property := range r.schemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
		tfSchema, err := r.createTerraformPropertySchema(propertyName, property)
		if err != nil {
			return nil, err
		}
		s[propertyName] = tfSchema
	}
	return s, nil
}

func (r resourceInfo) createTerraformPropertySchema(propertyName string, property spec.Schema) (*schema.Schema, error) {
	propertySchema, err := r.createTerraformPropertyBasicSchema(propertyName, property)
	if err != nil {
		return nil, err
	}
	// ValidateFunc is not yet supported on lists or sets
	if !r.isArrayProperty(property) {
		propertySchema.ValidateFunc = r.validateFunc(propertyName, property)
	}
	return propertySchema, nil
}

func (r resourceInfo) validateFunc(propertyName string, property spec.Schema) schema.SchemaValidateFunc {
	return func(v interface{}, k string) (ws []string, errors []error) {
		if property.Default != nil {
			if property.ReadOnly {
				err := fmt.Errorf(
					"'%s.%s' is configured as 'readOnly' and can not have a default value. The value is expected to be computed by the API. To fix the issue, pick one of the following options:\n"+
						"1. Remove the 'readOnly' attribute from %s in the swagger file so the default value '%v' can be applied. Default must be nil if computed\n"+
						"OR\n"+
						"2. Remove the 'default' attribute from %s in the swagger file, this means that the API will compute the value as specified by the 'readOnly' attribute\n", r.name, k, k, property.Default, k)
				errors = append(errors, err)
			}
		}
		return
	}
}

func (r resourceInfo) isRequired(propertyName string, requiredProps []string) bool {
	var required = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (r resourceInfo) createTerraformPropertyBasicSchema(propertyName string, property spec.Schema) (*schema.Schema, error) {
	var propertySchema *schema.Schema
	// Arrays only support 'string' items at the moment
	if r.isArrayProperty(property) {
		propertySchema = &schema.Schema{
			Type: schema.TypeList,
			Elem: &schema.Schema{Type: schema.TypeString},
		}
	} else if property.Type.Contains("string") {
		propertySchema = &schema.Schema{
			Type: schema.TypeString,
		}
	} else if property.Type.Contains("integer") {
		propertySchema = &schema.Schema{
			Type: schema.TypeInt,
		}
	} else if property.Type.Contains("number") {
		propertySchema = &schema.Schema{
			Type: schema.TypeFloat,
		}
	} else if property.Type.Contains("boolean") {
		propertySchema = &schema.Schema{
			Type: schema.TypeBool,
		}
	}

	// Set the property as required or optional
	required := r.isRequired(propertyName, r.schemaDefinition.Required)
	if required {
		propertySchema.Required = true
	} else {
		propertySchema.Optional = true
	}

	// If the value of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new value will be created
	if forceNew, ok := property.Extensions.GetBool(extTfForceNew); ok && forceNew {
		propertySchema.ForceNew = true
	}

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	if property.ReadOnly {
		propertySchema.Computed = true
	}

	// A sensitive property means that the value will not be disclosed in the state file, preventing secrets from
	// being leaked
	if sensitive, ok := property.Extensions.GetBool(extTfSensitive); ok && sensitive {
		propertySchema.Sensitive = true
	}

	if property.Default != nil {
		if property.ReadOnly {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error
			log.Printf("[WARN] '%s.%s' is readOnly and can not have a default value. The value is expected to be computed by the API. Terraform will fail on runtime when performing the property validation check", r.name, propertyName)
		} else {
			propertySchema.Default = property.Default
		}
	}
	return propertySchema, nil
}

func (r resourceInfo) isArrayProperty(property spec.Schema) bool {
	return property.Type.Contains("array")
}

func (r resourceInfo) getImmutableProperties() []string {
	var immutableProperties []string
	for propertyName, property := range r.schemaDefinition.Properties {
		if propertyName == "id" {
			continue
		}
		if immutable, ok := property.Extensions.GetBool(extTfImmutable); ok && immutable {
			immutableProperties = append(immutableProperties, propertyName)
		}
	}
	return immutableProperties
}

func (r resourceInfo) getResourceURL() (string, error) {
	if r.host == "" || r.path == "" {
		return "", fmt.Errorf("host and path are mandatory attributes to get the resource URL - host['%s'], path['%s']", r.host, r.path)
	}
	defaultScheme := "http"
	for _, scheme := range r.httpSchemes {
		if scheme == "https" {
			defaultScheme = "https"
		}
	}
	path := r.path
	if strings.Index(r.path, "/") != 0 {
		path = fmt.Sprintf("/%s", r.path)
	}
	if r.basePath != "" && r.basePath != "/" {
		if strings.Index(r.basePath, "/") == 0 {
			return fmt.Sprintf("%s://%s%s%s", defaultScheme, r.host, r.basePath, path), nil
		}
		return fmt.Sprintf("%s://%s/%s%s", defaultScheme, r.host, r.basePath, path), nil
	}
	return fmt.Sprintf("%s://%s%s", defaultScheme, r.host, path), nil
}

func (r resourceInfo) getResourceIDURL(id string) (string, error) {
	url, err := r.getResourceURL()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", url, id), nil
}

// getResourceIdentifier returns the property name that is supposed to be used as the identifier. The resource id
// is selected as follows:
// 1.If the given schema definition contains a property configured with metadata 'x-terraform-id' set to true, that property value
// will be used to set the state ID of the resource. Additionally, the value will be used when performing GET/PUT/DELETE requests to
// identify the resource in question.
// 2. If none of the properties of the given schema definition contain such metadata, it is expected that the payload
// will have a property named 'id'
// 3. If none of the above requirements is met, an error will be returned
func (r resourceInfo) getResourceIdentifier() (string, error) {
	isIDPropertyPresent := false
	for propertyName, property := range r.schemaDefinition.Properties {
		if propertyName == "id" {
			isIDPropertyPresent = true
			continue
		}
		// field with extTfID metadata takes preference over 'id' fields as the service provider is the one acknowledging
		// the fact that this field should be used as identifier of the resource
		if terraformID, ok := property.Extensions.GetBool(extTfID); ok && terraformID {
			return propertyName, nil
		}
	}
	// if the id field is missing and there isn't any properties set with extTfID, there is not way for the resource
	// to be identified and therefore an error is returned
	if !isIDPropertyPresent {
		return "", fmt.Errorf("could not find any identifier property in the resource payload swagger definition. Please make sure the payload definition has either one property named 'id' or one property that contains %s metadata", extTfID)
	}
	return "id", nil
}
