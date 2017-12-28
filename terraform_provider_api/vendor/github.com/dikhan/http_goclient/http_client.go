package http_goclient

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// HttpClient represents an http wrapper which reduces the boiler plate needed to marshall/un-marshall request/response
// bodies by providing friendly CRUD http operations that allow in/out interfaces
type HttpClient struct {
	HttpClient *http.Client
}

// Get issues a GET HTTP request to the specified URL including the headers passed in.
//
// The 'out' param interface is the un-marshall representation of the http response returned
//
// Example on how to invoke the method:
//
//	type Out struct {
// 		Id string `json:"id"`
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//
//  out := &Out{}
//  headers := map[string]string{"header_example": "header_value"}
//  HttpClient.Get("http://api.com/resource", headers, out)
//
func (httpClient *HttpClient) Get(url string, headers map[string]string, out interface{}) (*http.Response, error) {
	if req, err := httpClient.prepareRequest(http.MethodGet, url, headers, nil); err != nil {
		return nil, err
	} else {
		return httpClient.performRequest(req, out)
	}
}

// PostJson issues a POST to the specified URL including the headers passed in.
// The content type of the body is set to application/json so it doesn't need to be added to the headers passed in
//
// The 'in' param interface is marshall and added to the htp request body.
// The 'out' param interface is the un-marshall representation of the http response returned
//
// Example on how to invoke the method:
//
//	type In struct {
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//
//	type Out struct {
//		Id string `json:"id"`
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//  in := &In{}
//  out := &Out{}
//  headers := map[string]string{"header_example": "header_value"}
//  HttpClient.PostJson("http://api.com/resource", headers, in, out)
//
func (httpClient *HttpClient) PostJson(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	httpClient.addJsonHeader(headers)
	return httpClient.Post(url, headers, in, out)
}

// Post issues a POST HTTP request to the specified URL including the headers passed in.
//
// The in interface is marshall and added to the htp request body.
// The out interface is the un-marshall representation of the http response returned
//
// Example on how to invoke the method:
//
//	type In struct {
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//
//	type Out struct {
//		Id string `json:"id"`
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//  in := &In{}
//  out := &Out{}
//  headers := map[string]string{"header_example": "header_value"}
//  HttpClient.Post("http://api.com/resource", headers, in, out)
//
func (httpClient *HttpClient) Post(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	if req, err := httpClient.prepareRequest(http.MethodPost, url, headers, in); err != nil {
		return nil, err
	} else {
		return httpClient.performRequest(req, out)
	}
}

// PutJson issues a PUT HTTP request to the specified URL including the headers passed in.
// The content type of the body is set to application/json
//
// The 'in' param interface is marshall and added to the htp request body.
// The 'out' param interface is the un-marshall representation of the http response returned
//
// Example on how to invoke the method:
//
//	type In struct {
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//
//	type Out struct {
//		Id string `json:"id"`
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//  in := &In{}
//  out := &Out{}
//  headers := map[string]string{"header_example": "header_value"}
//  HttpClient.PutJson("http://api.com/resource", headers, in, out)
//
func (httpClient *HttpClient) PutJson(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	httpClient.addJsonHeader(headers)
	return httpClient.Put(url, headers, in, out)
}

// Post issues a PUT HTTP request to the specified URL including the headers passed in.
//
// The in interface is marshall and added to the http request body.
// The out interface is the un-marshall representation of the http response returned
//
// Example on how to invoke the method:
//
//	type In struct {
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//
//	type Out struct {
//		Id string `json:"id"`
//		Name        string   `json:"name"`
//		Description string   `json:"description"`
//	}
//  in := &In{}
//  out := &Out{}
//  headers := map[string]string{"header_example": "header_value"}
//  HttpClient.Put("http://api.com/resource", headers, in, out)
//
func (httpClient *HttpClient) Put(url string, headers map[string]string, in interface{}, out interface{}) (*http.Response, error) {
	if req, err := httpClient.prepareRequest(http.MethodPut, url, headers, in); err != nil {
		return nil, err
	} else {
		return httpClient.performRequest(req, out)
	}
}

// Delete issues a DELETE HTTP request to the specified URL including the headers passed in.
func (httpClient *HttpClient) Delete(url string, headers map[string]string) (*http.Response, error) {
	if req, err := httpClient.prepareRequest(http.MethodDelete, url, headers, nil); err != nil {
		return nil, err
	} else {
		return httpClient.performRequest(req, nil)
	}
}

func (httpClient *HttpClient) prepareRequest(method, url string, headers map[string]string, in interface{}) (*http.Request, error) {

	var body []byte
	var err error
	if in != nil {
		body, err = json.Marshal(in)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	return req, nil
}

func (httpClient *HttpClient) performRequest(req *http.Request, out interface{}) (*http.Response, error) {
	resp, err := httpClient.HttpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("request %s %s %s failed. Response Error [%s]: '%s'", req.Method, req.URL, req.Proto, resp.Status, err.Error())
	}

	if out != nil {
		var body []byte
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			return nil, err
		}
		if len(body) > 0 {
			if err = json.Unmarshal(body, &out); err != nil {
				return nil, fmt.Errorf("unable to unmarshal response body ['%s'] for request = '%s %s %s'. Response = '%s'", err.Error(), req.Method, req.URL, req.Proto, resp.Status)
			}
		} else {
			return nil, fmt.Errorf("expected a response body but response body received was empty for request = '%s %s %s'. Response = '%s'", req.Method, req.URL, req.Proto, resp.Status)
		}
	}
	return resp, nil
}

func (httpClient *HttpClient) addJsonHeader(headers map[string]string) {
	if headers == nil {
		headers = map[string]string{}
	}
	headers["Content-Type"] = "application/json"
}
