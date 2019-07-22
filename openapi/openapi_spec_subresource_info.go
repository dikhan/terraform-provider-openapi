package openapi

import "fmt"

type parentResourceInfo struct {
	parentResourceNames    []string
	fullParentResourceName string
	parentURIs             []string
	parentInstanceURIs     []string
}

// getParentPropertiesNames is responsible to building the parent properties names for a resource that is a parentResource
func (info *parentResourceInfo) getParentPropertiesNames() []string {
	parentPropertyNames := []string{}
	for _, parentName := range info.parentResourceNames {
		parentPropertyNames = append(parentPropertyNames, fmt.Sprintf("%s_id", parentName))
	}
	return parentPropertyNames
}
