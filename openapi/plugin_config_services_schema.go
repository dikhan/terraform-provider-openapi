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

// GetSwaggerURL returns the URL where the service swagger doc is exposed
func (s *ServiceConfigV1) GetSwaggerURL() string {
	return s.SwaggerURL
}

// IsInsecureSkipVerifyEnabled returns true if the given provider's service configuration has InsecureSkipVerify enabled; false
// otherwise
func (s *ServiceConfigV1) IsInsecureSkipVerifyEnabled() bool {
	return s.InsecureSkipVerify
}
