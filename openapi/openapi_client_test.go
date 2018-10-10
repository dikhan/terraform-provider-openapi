package openapi

import (
	"github.com/dikhan/http_goclient"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProviderClient(t *testing.T) {
	Convey("Given a SpecBackendConfiguration, HttpClient, providerConfiguration and specAuthenticator", t, func() {
		var openAPIBackendConfiguration SpecBackendConfiguration
		httpClient := http_goclient.HttpClient{}
		providerConfiguration := providerConfiguration{}
		var apiAuthenticator specAuthenticator
		Convey("When ProviderClient method is constructed", func() {
			providerClient := &ProviderClient{
				openAPIBackendConfiguration: openAPIBackendConfiguration,
				httpClient:                  httpClient,
				providerConfiguration:       providerConfiguration,
				apiAuthenticator:            apiAuthenticator,
			}
			Convey("Then the providerClient should comply with ClientOpenAPI interface", func() {
				var _ ClientOpenAPI = providerClient
			})
		})
	})
}
