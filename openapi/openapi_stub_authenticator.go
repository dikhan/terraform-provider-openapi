package openapi

type specStubAuthenticator struct {
	authContext *authContext
	err         error
}

func newStubAuthenticator(expectedHeader, expectedHeaderValue string, err error) *specStubAuthenticator {
	return &specStubAuthenticator{
		authContext: &authContext{
			url: "",
			headers: map[string]string{
				expectedHeader: expectedHeaderValue,
			},
		},
		err: err,
	}
}

func (s *specStubAuthenticator) prepareAuth(url string, operationSecuritySchemes SpecSecuritySchemes, providerConfig providerConfiguration) (*authContext, error) {
	// mimicking api key header auth which does not change the url at all
	if s.authContext.url == "" {
		s.authContext.url = url
	}
	return s.authContext, s.err
}
