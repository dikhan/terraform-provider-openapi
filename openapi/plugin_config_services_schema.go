package openapi

// ServiceConfiguration defines the interface/expected behaviour for ServiceConfiguration implementations.
type ServiceConfiguration interface {
	// GetSwaggerURL returns the URL where the service swagger doc is exposed
	GetSwaggerURL() string
	// IsInsecureSkipVerifyEnabled returns true if the given provider's service configuration has InsecureSkipVerify enabled; false
	// otherwise
	IsInsecureSkipVerifyEnabled() bool
}

// ServiceConfigV1 defines configuration for the service provider
type ServiceConfigV1 struct {
	// SwaggerURL defines where the swagger is located
	SwaggerURL string `yaml:"swagger-url"`
	// InsecureSkipVerify defines whether the internal http client used to fetch the swagger file should verify the server cert
	// or not. This should only be used purposefully if the server is using a self-signed cert and only if the server is trusted
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

// NewServiceConfigV1WithDefaults creates a new ServiceConfigV1 with the swaggerURL passed in and default configuration
// for the rest of the struct fields
func NewServiceConfigV1WithDefaults(swaggerURL string) *ServiceConfigV1 {
	return NewServiceConfigV1(swaggerURL, false)
}

// NewServiceConfigV1 creates a new instance of NewServiceConfigV1 struct with the values provided
func NewServiceConfigV1(swaggerURL string, insecureSkipVerifyEnabled bool) *ServiceConfigV1 {
	return &ServiceConfigV1{
		SwaggerURL:         swaggerURL,
		InsecureSkipVerify: insecureSkipVerifyEnabled,
	}
}

// GetSwaggerURL returns the URL where the service swagger doc is exposed
func (s *ServiceConfigV1) GetSwaggerURL() string {
	return s.SwaggerURL
}

// IsInsecureSkipVerifyEnabled returns true if the given provider's service configuration has InsecureSkipVerify enabled; false
// otherwise
func (s *ServiceConfigV1) IsInsecureSkipVerifyEnabled() bool {
	return s.InsecureSkipVerify
}
