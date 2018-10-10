package openapi

type specStubBackendConfiguration struct {
	host        string
	basePath    string
	httpSchemes []string
}

func (s *specStubBackendConfiguration) getHost() (string, error) {
	return s.host, nil
}
func (s *specStubBackendConfiguration) getBasePath() string {
	return s.basePath
}

func (s *specStubBackendConfiguration) getHTTPSchemes() []string {
	return s.httpSchemes
}
