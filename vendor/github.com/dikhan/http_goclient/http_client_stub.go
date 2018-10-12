package http_goclient

import (
	"net/http"
)

// HttpClientStub implements the HttpClientIface and should be used for unit testing purposes
type HttpClientStub struct {
	// Properties that tests can run assertions against to
	URL      string
	Headers  map[string]string
	In       interface{}
	Out      interface{}
	// Stub response
	Response *http.Response
	// Stub error
	Error    error
}



func (c *HttpClientStub) Get(url string, headers map[string]string, out interface{}) (*http.Response, error) {
	c.updateStub(url, headers, nil, out)
	return c.Response, c.Error
}

func (c *HttpClientStub) updateStub(url string, headers map[string]string, in interface{}, out interface{}) {
	c.URL = url
	c.Headers = headers
	c.In = in
	c.Out = out
}

func (c *HttpClientStub) PostJson(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	c.updateStub(url, headers, in, out)
	return c.Response, c.Error
}

func (c *HttpClientStub) Post(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	c.updateStub(url, headers, in, out)
	return c.Response, c.Error
}

func (c *HttpClientStub) PutJson(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	c.updateStub(url, headers, in, out)
	return c.Response, c.Error
}

func (c *HttpClientStub) Put(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	c.updateStub(url, headers, in, out)
	return c.Response, c.Error
}

func (c *HttpClientStub) Delete(url string, headers map[string]string) (*http.Response, error) {
	c.updateStub(url, headers, nil, nil)
	return c.Response, c.Error
}

