package openapi

// specStubResource is a stub implementation of SpecResource interface which is used for testing purposes
type specStubResource struct {
	name                    string
	path                    string
	shouldIgnore            bool
	schemaDefinition        *SchemaDefinition
	resourceGetOperation    *ResourceOperation
	resourcePostOperation   *ResourceOperation
	resourcePutOperation    *ResourceOperation
	resourceDeleteOperation *ResourceOperation
	timeouts                *specTimeouts
}

func newSpecStubResource(name, path string, shouldIgnore bool, schemaDefinition *SchemaDefinition) *specStubResource {
	return newSpecStubResourceWithOperations(name, path, shouldIgnore, schemaDefinition, nil, nil, nil, nil)
}

func newSpecStubResourceWithOperations(name, path string, shouldIgnore bool, schemaDefinition *SchemaDefinition, resourcePostOperation, resourcePutOperation, resourceGetOperation, resourceDeleteOperation *ResourceOperation) *specStubResource {
	return &specStubResource{
		name:                    name,
		path:                    path,
		schemaDefinition:        schemaDefinition,
		shouldIgnore:            shouldIgnore,
		resourcePostOperation:   resourcePostOperation,
		resourceGetOperation:    resourceGetOperation,
		resourceDeleteOperation: resourceDeleteOperation,
		resourcePutOperation:    resourcePutOperation,
		timeouts:                &specTimeouts{},
	}
}

func (s *specStubResource) getResourceName() string { return s.name }

func (s *specStubResource) getResourcePath() string { return s.path }

func (s *specStubResource) getResourceSchema() (*SchemaDefinition, error) {
	return s.schemaDefinition, nil
}

func (s *specStubResource) shouldIgnoreResource() bool { return s.shouldIgnore }

func (s *specStubResource) getResourcePostOperation() *ResourceOperation {
	return s.resourcePostOperation
}
func (s *specStubResource) getResourceGetOperation() *ResourceOperation {
	return s.resourceGetOperation
}
func (s *specStubResource) getResourcePutOperation() *ResourceOperation {
	return s.resourcePutOperation
}
func (s *specStubResource) getResourceDeleteOperation() *ResourceOperation {
	return s.resourceDeleteOperation
}

func (s *specStubResource) getTimeouts() (*specTimeouts, error) {
	return s.timeouts, nil
}
