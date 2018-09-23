package openapi

type specSecurityStub struct {
	securityDefinitions   *SpecSecurityDefinitions
	globalSecuritySchemes SpecSecuritySchemes
	error                 error
}

func (s *specSecurityStub) GetAPIKeySecurityDefinitions() (*SpecSecurityDefinitions, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.securityDefinitions, nil
}

func (s *specSecurityStub) GetGlobalSecuritySchemes() (SpecSecuritySchemes, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.globalSecuritySchemes, nil
}
