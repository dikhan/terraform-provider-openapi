package openapi

type specResponses map[int]*specResponse

type specResponse struct {
	isPollingEnabled    bool
	pollTargetStatuses  []string
	pollPendingStatuses []string
}

func (s specResponses) getResponse(responseStatusCode int) *specResponse {
	response, exists := s[responseStatusCode]
	if !exists {
		return nil
	}
	return response
}
