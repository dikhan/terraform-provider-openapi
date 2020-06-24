package openapi

import "fmt"

// ParentResourceInfo contains the information related to the parent information. For instance, a subresource would have
// this struct populated with the parent info so the resource name and corresponding parent properties can be configured in the
// resource schema
type ParentResourceInfo struct {
	parentResourceNames    []string
	fullParentResourceName string
	parentURIs             []string
	parentInstanceURIs     []string
}

// GetParentPropertiesNames is responsible to building the parent properties names for a resource that is a subresource
func (info *ParentResourceInfo) GetParentPropertiesNames() []string {
	parentPropertyNames := []string{}
	for _, parentName := range info.parentResourceNames {
		parentPropertyNames = append(parentPropertyNames, fmt.Sprintf("%s_id", parentName))
	}
	return parentPropertyNames
}
