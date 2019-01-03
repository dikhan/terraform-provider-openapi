package openapi

import "fmt"

type specStubBackendConfiguration struct {
	host             string
	basePath         string
	httpSchemes      []string
	regions          []string
	err              error
	hostErr          error
	defaultRegionErr error
	hostByRegionErr  error
}

func newStubBackendConfiguration(host, basePath string, httpSchemes []string) *specStubBackendConfiguration {
	return &specStubBackendConfiguration{
		host:        host,
		basePath:    basePath,
		httpSchemes: httpSchemes,
	}
}

func (s *specStubBackendConfiguration) getHost() (string, error) {
	if s.hostErr != nil {
		return "", s.hostErr
	}
	return s.host, nil
}
func (s *specStubBackendConfiguration) getBasePath() string {
	return s.basePath
}

func (s *specStubBackendConfiguration) getHTTPSchemes() []string {
	return s.httpSchemes
}

func (s *specStubBackendConfiguration) getHostByRegion(region string) (string, error) {
	if s.hostByRegionErr != nil {
		return "", s.hostByRegionErr
	}
	return fmt.Sprintf(s.host, region), nil
}

func (s *specStubBackendConfiguration) getDefaultRegion() (string, error) {
	if s.defaultRegionErr != nil {
		return "", s.defaultRegionErr
	}
	return s.regions[0], nil
}

func (s *specStubBackendConfiguration) isMultiRegion() (bool, string, []string, error) {
	if s.err != nil {
		return false, "", nil, s.err
	}
	if len(s.regions) > 0 {
		return true, s.host, s.regions, nil
	}
	return false, "", nil, nil
}
