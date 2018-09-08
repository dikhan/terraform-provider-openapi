package openapi

import (
	"github.com/go-openapi/spec"
	"log"
)

const resourceVersionRegex = "(/v[0-9]*/)"
const resourceNameRegex = "((/\\w*/){\\w*})+$"
const resourceInstanceRegex = "((?:.*)){.*}"
const swaggerResourcePayloadDefinitionRegex = "(\\w+)[^//]*$"

const extTfImmutable = "x-terraform-immutable"
const extTfForceNew = "x-terraform-force-new"
const extTfSensitive = "x-terraform-sensitive"
const extTfFieldName = "x-terraform-field-name"
const extTfID = "x-terraform-id"
const extTfExcludeResource = "x-terraform-exclude-resource"

type specV2Resource struct {
	// name defines the name of the resource (including the version if applicable). This name is used along with
	// the provider name to build the terraform resource name that will be used in the terraform configuration file
	name string
	// path contains the full relative path to the resource e,g: /v1/resource
	path string
	// schema definition represents the representational state (aka model) of the resource
	schemaDefinition spec.Schema
	// rootPathItem contains info about the resource root path e,g: /resource, including the POST operation used to create instances of this resource
	rootPathItem spec.PathItem
	// instancePathItem contains info about the resource's instance /resource/{id}, including GET, PUT and REMOVE operations if applicable
	instancePathItem spec.PathItem
}

func (o *specV2Resource) getResourceName() string {
	return o.name
}

func (o *specV2Resource) getResourcePath() string {
	return o.path
}

func (o *specV2Resource) getResourcePostOperation() *ResourceOperation {
	if o.rootPathItem.Post == nil {
		return nil
	}
	return o.createResourceOperation(o.rootPathItem.Post)
}

func (o *specV2Resource) getResourceGetOperation() *ResourceOperation {
	if o.instancePathItem.Get == nil {
		return nil
	}
	return o.createResourceOperation(o.instancePathItem.Get)
}

func (o *specV2Resource) getResourcePutOperation() *ResourceOperation {
	if o.instancePathItem.Put == nil {
		return nil
	}
	return o.createResourceOperation(o.instancePathItem.Put)
}

func (o *specV2Resource) getResourceDeleteOperation() *ResourceOperation {
	if o.instancePathItem.Delete == nil {
		return nil
	}
	return o.createResourceOperation(o.instancePathItem.Delete)
}

// shouldIgnoreResource checks whether the POST operation for a given resource as the 'x-terraform-exclude-resource' extension
// defined with true value. If so, the resource will not be exposed to the OpenAPI Terraform provder; otherwise it will
// be exposed and users will be able to manage such resource via terraform.
func (o *specV2Resource) shouldIgnoreResource() bool {
	if extensionExists, ignoreResource := o.rootPathItem.Post.Extensions.GetBool(extTfExcludeResource); extensionExists && ignoreResource {
		return true
	}
	return false
}

func (o *specV2Resource) createResourceOperation(operation *spec.Operation) *ResourceOperation {
	headerParameters := getHeaderConfigurations(operation.Parameters)
	securitySchemes := createSecuritySchemes(operation.Security)
	return &ResourceOperation{
		HeaderParameters: headerParameters,
		SecuritySchemes:  securitySchemes,
	}
}

func (o *specV2Resource) getResourceSchema() (SchemaDefinition, error) {
	schemaDefinition := SchemaDefinition{}
	schemaDefinition.Properties = map[string]SchemaDefinitionProperty{}
	for propertyName, property := range o.schemaDefinition.Properties {
		schemaDefinitionProperty, err := o.createSchemaDefinitionProperty(propertyName, property)
		if err != nil {
			return SchemaDefinition{}, err
		}
		schemaDefinition.Properties[schemaDefinitionProperty.Name] = schemaDefinitionProperty
	}
	return schemaDefinition, nil
}

func (o *specV2Resource) createSchemaDefinitionProperty(propertyName string, property spec.Schema) (SchemaDefinitionProperty, error) {
	schemaDefinitionProperty := SchemaDefinitionProperty{}

	schemaDefinitionProperty.Name = propertyName

	if preferredPropertyName, exists := property.Extensions.GetString(extTfFieldName); exists {
		schemaDefinitionProperty.PreferredName = preferredPropertyName
	}

	if o.isArrayProperty(property) {
		schemaDefinitionProperty.Type = typeList
	} else if property.Type.Contains("string") {
		schemaDefinitionProperty.Type = typeString
	} else if property.Type.Contains("integer") {
		schemaDefinitionProperty.Type = typeInt
	} else if property.Type.Contains("number") {
		schemaDefinitionProperty.Type = typeFloat
	} else if property.Type.Contains("boolean") {
		schemaDefinitionProperty.Type = typeBool
	}

	// Set the property as required or optional
	required := o.isRequired(propertyName, o.schemaDefinition.Required)
	if required {
		schemaDefinitionProperty.Required = true
	}

	// If the value of the property is changed, it will force the deletion of the previous generated resource and
	// a new resource with this new value will be created
	if forceNew, ok := property.Extensions.GetBool(extTfForceNew); ok && forceNew {
		schemaDefinitionProperty.ForceNew = true
	}

	// A readOnly property is the one that is not used to create a resource (property is not exposed to the user); but
	// it comes back from the api and is stored in the state. This properties are mostly informative.
	if property.ReadOnly {
		schemaDefinitionProperty.ReadOnly = true
	}

	// A sensitive property means that the value will not be disclosed in the state file, preventing secrets from
	// being leaked
	if sensitive, ok := property.Extensions.GetBool(extTfSensitive); ok && sensitive {
		schemaDefinitionProperty.Sensitive = true
	}

	// field with extTfID metadata takes preference over 'id' fields as the service provider is the one acknowledging
	// the fact that this field should be used as identifier of the resource
	if terraformID, ok := property.Extensions.GetBool(extTfID); ok && terraformID {
		schemaDefinitionProperty.IsIdentifier = true
	}

	if immutable, ok := property.Extensions.GetBool(extTfImmutable); ok && immutable {
		schemaDefinitionProperty.Immutable = true
	}

	if property.Default != nil {
		if property.ReadOnly {
			// Below we just log a warn message; however, the validateFunc will take care of throwing an error if the following happens
			// Check r.validateFunc which will handle this use case on runtime and provide the user with a detail description of the error
			log.Printf("[WARN] '%s.%s' is readOnly and can not have a default value. The value is expected to be computed by the API. Terraform will fail on runtime when performing the property validation check", o.name, propertyName)
		} else {
			schemaDefinitionProperty.Default = property.Default
		}
	}
	return schemaDefinitionProperty, nil
}

func (o *specV2Resource) isRequired(propertyName string, requiredProps []string) bool {
	var required = false
	for _, f := range requiredProps {
		if f == propertyName {
			required = true
		}
	}
	return required
}

func (o *specV2Resource) isArrayProperty(property spec.Schema) bool {
	return property.Type.Contains("array")
}
