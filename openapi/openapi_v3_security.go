package openapi

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

type specV3Security struct {
	SecuritySchemes openapi3.SecuritySchemes
	GlobalSecurity  openapi3.SecurityRequirements
}

// GetAPIKeySecurityDefinitions returns a list of SpecSecurityDefinition after looping through the SecuritySchemes
// and selecting only the SecuritySchemes of type apiKey
func (s *specV3Security) GetAPIKeySecurityDefinitions() (*SpecSecurityDefinitions, error) {
	securityDefinitions := &SpecSecurityDefinitions{}
	for secDefName, secDef := range s.SecuritySchemes {
		// TODO: support .Ref
		secDefSchema := secDef.Value
		if secDefSchema.Type == "apiKey" {
			var securityDefinition SpecSecurityDefinition
			switch secDefSchema.In {
			case "header":
				if refreshTokenURL := s.isRefreshTokenAuth(secDefSchema); refreshTokenURL != "" {
					securityDefinition = newAPIKeyHeaderRefreshTokenSecurityDefinition(secDefName, refreshTokenURL)
				} else if s.isBearerScheme(secDefSchema) {
					securityDefinition = newAPIKeyHeaderBearerSecurityDefinition(secDefName)
				} else {
					securityDefinition = newAPIKeyHeaderSecurityDefinition(secDefName, secDefSchema.Name)
				}
			case "query":
				if s.isBearerScheme(secDefSchema) {
					securityDefinition = newAPIKeyQueryBearerSecurityDefinition(secDefName)
				} else {
					securityDefinition = newAPIKeyQuerySecurityDefinition(secDefName, secDefSchema.Name)
				}
			default:
				return nil, fmt.Errorf("apiKey In value '%s' not supported, only 'header' and 'query' values are valid", secDefSchema.In)
			}
			if err := securityDefinition.validate(); err != nil {
				return nil, err
			}
			*securityDefinitions = append(*securityDefinitions, securityDefinition)
		}
	}
	return securityDefinitions, nil
}

func (s *specV3Security) isBearerScheme(secDef *openapi3.SecurityScheme) bool {
	authScheme, enabled := getExtensionAsJsonBool(secDef.Extensions, extTfAuthenticationSchemeBearer)
	if authScheme && enabled {
		return true
	}
	return false
}

func (s *specV3Security) isRefreshTokenAuth(secDef *openapi3.SecurityScheme) string {
	refreshTokenURL, isRefreshTokenAuth := getExtensionAsJsonString(secDef.Extensions, extTfAuthenticationRefreshToken)
	if isRefreshTokenAuth {
		return refreshTokenURL
	}
	return ""
}

// GetGlobalSecuritySchemes returns a list of SpecSecuritySchemes that have their corresponding SpecSecurityDefinition
func (s *specV3Security) GetGlobalSecuritySchemes() (SpecSecuritySchemes, error) {
	var secSchemes []map[string][]string
	for _, sec := range s.GlobalSecurity {
		secSchemes = append(secSchemes, sec)
	}
	securitySchemes := createSecuritySchemes(secSchemes)
	for _, securityScheme := range securitySchemes {
		secDef, err := s.GetAPIKeySecurityDefinitions()
		if err != nil {
			return SpecSecuritySchemes{}, nil
		}
		secDefFound := secDef.findSecurityDefinitionFor(securityScheme.Name)
		if secDefFound == nil {
			return nil, fmt.Errorf("global security scheme '%s' not found or not matching supported 'apiKey' type", securityScheme.Name)
		}
	}
	return securitySchemes, nil
}
