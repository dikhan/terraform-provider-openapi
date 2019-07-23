package openapi

import (
	"fmt"

	"github.com/dikhan/terraform-provider-openapi/openapi/openapiutils"
)

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

func newStubBackendMultiRegionConfiguration(host string, regions []string) *specStubBackendConfiguration {
	isMultiRegion, _ := openapiutils.IsMultiRegionHost(host)
	if !isMultiRegion {
		return nil
	}
	return &specStubBackendConfiguration{
		host:    host,
		regions: regions,
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

//TODO: delete me
func (s *specStubBackendConfiguration) getHTTPSchemes() []string {
	panic("quit calling me")
}

//TODO: rename me to getHTTPScheme
// TODO: make this configurable via the specStubBackendConfiguration, instead of httpSchemes the stub should contain
//  field httpScheme string that this method returns
func (s *specStubBackendConfiguration) getHTTPSchemes2() (string, error) {
	return s.httpSchemes[0], nil
}

func (s *specStubBackendConfiguration) getHostByRegion(region string) (string, error) {
	if s.hostByRegionErr != nil {
		return "", s.hostByRegionErr
	}
	return fmt.Sprintf(s.host, region), nil
}

func (s *specStubBackendConfiguration) getDefaultRegion(regions []string) (string, error) {
	if s.defaultRegionErr != nil {
		return "", s.defaultRegionErr
	}
	if len(regions) == 0 {
		return "", fmt.Errorf("empty regions provided")
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
