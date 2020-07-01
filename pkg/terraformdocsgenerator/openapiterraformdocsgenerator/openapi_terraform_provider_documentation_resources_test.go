package openapiterraformdocsgenerator

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestContainsResourcesWithSecretProperties(t *testing.T) {
	testCases := []struct {
		name                    string
		providerResources       ProviderResources
		expectedContainsSecrets bool
	}{
		{
			name: "no resources with sensitive props",
			providerResources: ProviderResources{
				Resources: []Resource{
					{Properties: []Property{{IsSensitive: false}}},
				},
			},
			expectedContainsSecrets: false,
		},
		{
			name: "resource has sensitive prop",
			providerResources: ProviderResources{
				Resources: []Resource{
					{Properties: []Property{{IsSensitive: true}}},
				},
			},
			expectedContainsSecrets: true,
		},
	}

	for _, tc := range testCases {
		actualContainsSecrets := tc.providerResources.ContainsResourcesWithSecretProperties()
		assert.Equal(t, tc.expectedContainsSecrets, actualContainsSecrets, tc.name)
	}
}

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
