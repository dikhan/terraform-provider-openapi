package openapi

import "fmt"

type subResourceInfo struct {
	parentResourceNames    []string
	fullParentResourceName string
	parentURIs             []string
	parentInstanceURIs     []string
}

// getParentPropertiesNames is responsible to building the parent properties names for a resource that is a subresource
func (sub *subResourceInfo) getParentPropertiesNames() []string {
	parentPropertyNames := []string{}
	for _, parentName := range sub.parentResourceNames {
		parentPropertyNames = append(parentPropertyNames, fmt.Sprintf("%s_id", parentName))
	}
	return parentPropertyNames
}
