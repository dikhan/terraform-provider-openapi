package openapi

import (
	"io/ioutil"
	"net/http"
	"strings"
)

// clientOpenAPIStub is a stubbed client used for testing purposes that implements the ClientOpenAPI interface
type clientOpenAPIStub struct {
	ClientOpenAPI
	responsePayload map[string]interface{}
	error           error
	returnHTTPCode  int
}

func (c *clientOpenAPIStub) Post(resource SpecResource, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	switch p := responsePayload.(type) {
	case *map[string]interface{}:
		*p = c.responsePayload
	default:
		panic("unexpected type")
	}
	return c.generateStubResponse(http.StatusCreated), nil
}

func (c *clientOpenAPIStub) Put(resource SpecResource, id string, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	switch p := responsePayload.(type) {
	case *map[string]interface{}:
		*p = c.responsePayload
	default:
		panic("unexpected type")
	}
	return c.generateStubResponse(http.StatusOK), nil
}

func (c *clientOpenAPIStub) Get(resource SpecResource, id string, responsePayload interface{}) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	switch p := responsePayload.(type) {
	case *map[string]interface{}:
		*p = c.responsePayload
	default:
		panic("unexpected type")
	}

	return c.generateStubResponse(http.StatusOK), nil
}

func (c *clientOpenAPIStub) Delete(resource SpecResource, id string) (*http.Response, error) {
	if c.error != nil {
		return nil, c.error
	}
	delete(c.responsePayload, id)
	return c.generateStubResponse(http.StatusNoContent), nil
}

func (c *clientOpenAPIStub) generateStubResponse(defaultHTTPCode int) *http.Response {
	return &http.Response{
		StatusCode: c.returnCode(defaultHTTPCode),
		Body:       ioutil.NopCloser(strings.NewReader("")),
	}
}

func (c *clientOpenAPIStub) returnCode(defaultHTTPCode int) int {
	if c.returnHTTPCode != 0 {
		return c.returnHTTPCode
	}
	return defaultHTTPCode
}
