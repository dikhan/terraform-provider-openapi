package openapi

import (
	"fmt"
	"github.com/go-openapi/spec"
)

type specV2Security struct {
	SecurityDefinitions spec.SecurityDefinitions
	GlobalSecurity      []map[string][]string
}

// GetAPIKeySecurityDefinitions returns a list of SpecSecurityDefinition after looping through the SecurityDefinitions
// and selecting only the SecurityDefinitions of type apiKey
func (s *specV2Security) GetAPIKeySecurityDefinitions() SpecSecurityDefinitions {
	securityDefinitions := SpecSecurityDefinitions{}
	for secDefName, secDef := range s.SecurityDefinitions {
		if secDef.Type == "apiKey" {
			securityDefinitions = append(securityDefinitions, SpecSecurityDefinition{Type: "apiKey", In: secDef.In, Name: secDefName})
		}
	}
	return securityDefinitions
}

// GetGlobalSecuritySchemes returns a list of SpecSecuritySchemes that have their corresponding SpecSecurityDefinition
func (s *specV2Security) GetGlobalSecuritySchemes() (SpecSecuritySchemes, error) {
	securitySchemes := createSecuritySchemes(s.GlobalSecurity)
	for _, securityScheme := range securitySchemes {
		secDef := s.GetAPIKeySecurityDefinitions().findSecurityDefinitionFor(securityScheme.Name)
		if secDef == nil {
			return nil, fmt.Errorf("global security scheme '%s' not found or not matching supported 'apiKey' type", securityScheme.Name)
		}
	}
	return securitySchemes, nil
}
