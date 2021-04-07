package openapiterraformdocsgenerator

// Property defines the attributes for describing a given property for a resource
type Property struct {
	Name               string
	Type               string
	ArrayItemsType     string
	Required           bool
	Computed           bool
	IsOptionalComputed bool
	IsSensitive        bool
	IsParent           bool
	Description        string
	Default            interface{}
	Schema             []Property // This is used to describe the schema for array of objects or object properties
}

// ContainsComputedSubProperties checks if a schema contains properties that are computed recursively
func (p Property) ContainsComputedSubProperties() bool {
	for _, s := range p.Schema {
		if s.Computed || s.ContainsComputedSubProperties() {
			return true
		}
	}
	return false
}

// DefaultNotNil checks whether the Default value is nil. If the value is populated it returns true, false otherwise
func (p Property) DefaultNotNil() bool {
	return p.Default != nil
}
