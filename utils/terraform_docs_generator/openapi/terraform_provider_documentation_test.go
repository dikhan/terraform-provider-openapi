package openapi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResource_BuildImportIDsExample(t *testing.T) {
	testCases := []struct {
		name              string
		parentProperties  []string
		expectedImportIDs string
	}{
		{
			name:              "resource configured with resource parent properties",
			parentProperties:  []string{"parent_id"},
			expectedImportIDs: "parent_id/fw_id",
		},
		{
			name:              "resource configured with NO resource parent properties",
			parentProperties:  nil,
			expectedImportIDs: "id",
		},
	}
	for _, tc := range testCases {
		resource := Resource{
			Name:             "fw",
			ParentProperties: tc.parentProperties,
		}
		result := resource.BuildImportIDsExample()
		assert.Equal(t, tc.expectedImportIDs, result)
	}
}

func TestProperty_ContainsComputedSubProperties(t *testing.T) {
	testCases := []struct {
		name           string
		property       Property
		expectedResult bool
	}{
		{
			name: "property does not have schema",
			property: Property{
				Name:   "some primitive property",
				Schema: nil,
			},
			expectedResult: false,
		},
		{
			name: "property does have a schema",
			property: Property{
				Name: "some property with schema (eg: object or array of objects) containing computed props",
				Schema: []Property{
					{
						Name:     "subProperty",
						Computed: true,
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "property does have a schema",
			property: Property{
				Name: "some property with schema (eg: object or array of objects) with no computed props",
				Schema: []Property{
					{
						Name:     "subProperty",
						Computed: false,
					},
				},
			},
			expectedResult: false,
		},
	}
	for _, tc := range testCases {
		result := tc.property.ContainsComputedSubProperties()
		assert.Equal(t, tc.expectedResult, result)
	}
}
