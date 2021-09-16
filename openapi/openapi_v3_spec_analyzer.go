package openapi

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
)

// specV3Analyser defines an SpecAnalyser implementation for OpenAPI v3 specification
// Forcing creation of this object via constructor so proper input validation is performed before creating the struct
// instance
type specV3Analyser struct {
	openAPIDocumentURL string
	d                  *openapi3.T
}

var _ SpecAnalyser = (*specV3Analyser)(nil)

// newSpecAnalyserV3 creates an instance of specV2Analyser which implements the SpecAnalyser interface
// This implementation provides an analyser that understands an OpenAPI v2 document
func newSpecAnalyserV3(openAPIDocumentFilename string) (*specV3Analyser, error) {
	if openAPIDocumentFilename == "" {
		return nil, errors.New("open api document filename argument empty, please provide the url of the OpenAPI document")
	}
	openAPIDocumentURL, err := url.Parse(openAPIDocumentFilename)
	if err != nil {
		return nil, fmt.Errorf("invalid URL to retrieve OpenAPI document: '%s' - error = %s", openAPIDocumentFilename, err)
	}
	apiSpec, err := openapi3.NewLoader().LoadFromURI(openAPIDocumentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve the OpenAPI document from '%s' - error = %s", openAPIDocumentFilename, err)
	}
	return &specV3Analyser{
		d:                  apiSpec,
		openAPIDocumentURL: openAPIDocumentFilename,
	}, nil
}

func (s *specV3Analyser) GetTerraformCompliantResources() ([]SpecResource, error) {
	return []SpecResource{}, nil
}

func (s specV3Analyser) GetTerraformCompliantDataSources() []SpecResource {
	return []SpecResource{}
}

func (s specV3Analyser) GetSecurity() SpecSecurity {
	// TODO: replace this stub
	return &specSecurityStub{
		securityDefinitions: &SpecSecurityDefinitions{
			newAPIKeyHeaderSecurityDefinition("apikey_auth", "Authorization"),
		},
		globalSecuritySchemes: createSecuritySchemes([]map[string][]string{}),
	}
}

func (s specV3Analyser) GetAllHeaderParameters() SpecHeaderParameters {
	// TODO: add support for header params
	return []SpecHeaderParam{
		{
			Name:          "X-Request-ID",
			TerraformName: "x_request_id",
			IsRequired:    true,
		},
	}
}

func (s specV3Analyser) GetAPIBackendConfiguration() (SpecBackendConfiguration, error) {
	return newOpenAPIBackendConfigurationV3(s.d, s.openAPIDocumentURL)
}
