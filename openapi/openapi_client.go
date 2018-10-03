package openapi

import (
	"fmt"
	"github.com/dikhan/http_goclient"
	"log"
	"net/http"
	"strings"
)

type httpMethodSupported string

const (
	httpGet    httpMethodSupported = "GET"
	httpPost   httpMethodSupported = "POST"
	httpPut    httpMethodSupported = "PUT"
	httpDelete httpMethodSupported = "DELETE"
)

// ClientOpenAPI defines the behaviour expected to be implemented for the OpenAPI Client used in the Terraform OpenAPI Provider
type ClientOpenAPI interface {
	Post(resource SpecResource, requestPayload interface{}, responsePayload interface{}) (*http.Response, error)
	Put(resource SpecResource, id string, requestPayload interface{}, responsePayload interface{}) (*http.Response, error)
	Get(resource SpecResource, id string, responsePayload interface{}) (*http.Response, error)
	Delete(resource SpecResource, id string) (*http.Response, error)
}

// ProviderClient defines a client that is configured based on the OpenAPI server side documentation
// The CRUD operations accept an OpenAPI operation which defines among other things the security scheme applicable to
// the API when making the HTTP request
type ProviderClient struct {
	openAPIBackendConfiguration SpecBackendConfiguration
	httpClient                  http_goclient.HttpClient
	providerConfiguration       providerConfiguration
	apiAuthenticator            specAuthenticator
}

// Post performs a POST request to the server API based on the resource configuration and the payload passed in
func (o *ProviderClient) Post(resource SpecResource, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	resourceURL, err := o.getResourceURL(resource)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Post
	return o.performRequest(httpPost, resourceURL, operation, requestPayload, responsePayload)
}

// Put performs a PUT request to the server API based on the resource configuration and the payload passed in
func (o *ProviderClient) Put(resource SpecResource, id string, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(resource, id)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Put
	return o.performRequest(httpPut, resourceURL, operation, requestPayload, responsePayload)
}

// Get performs a GET request to the server API based on the resource configuration and the resource instance id passed in
func (o *ProviderClient) Get(resource SpecResource, id string, responsePayload interface{}) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(resource, id)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Get
	return o.performRequest(httpGet, resourceURL, operation, nil, responsePayload)
}

// Delete performs a DELETE request to the server API based on the resource configuration and the resource instance id passed in
func (o *ProviderClient) Delete(resource SpecResource, id string) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(resource, id)
	if err != nil {
		return nil, err
	}
	operation := resource.getResourceOperations().Delete
	return o.performRequest(httpDelete, resourceURL, operation, nil, nil)
}

func (o *ProviderClient) performRequest(method httpMethodSupported, resourceURL string, operation *specResourceOperation, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {

	reqContext, err := o.apiAuthenticator.prepareAuth(resourceURL, operation.SecuritySchemes, o.providerConfiguration)
	if err != nil {
		return nil, err
	}
	reqContext.headers = o.appendOperationHeaders(operation.HeaderParameters, o.providerConfiguration, reqContext.headers)
	log.Printf("[DEBUG] Performing %s %s", method, reqContext.url)

	log.Printf("[DEBUG] Headers %s %s", method, reqContext.headers)

	switch method {
	case httpPost:
		return o.httpClient.PostJson(reqContext.url, reqContext.headers, requestPayload, &responsePayload)
	case httpPut:
		return o.httpClient.PutJson(reqContext.url, reqContext.headers, requestPayload, &responsePayload)
	case httpGet:
		return o.httpClient.Get(reqContext.url, reqContext.headers, &responsePayload)
	case httpDelete:
		return o.httpClient.Delete(reqContext.url, reqContext.headers)
	}
	return nil, fmt.Errorf("method '%s' not supported", method)
}

// appendOperationHeaders returns a maps containing the headers passed in and adds whatever headers the operation requires. The values
// are retrieved from the provider configuration.
func (o ProviderClient) appendOperationHeaders(operationHeaders []SpecHeaderParam, providerConfig providerConfiguration, headers map[string]string) map[string]string {
	if operationHeaders != nil && len(operationHeaders) > 0 {
		for _, headerParam := range operationHeaders {
			// Setting the actual name of the header with the expectedValue coming from the provider configuration
			headers[headerParam.Name] = providerConfig.getHeaderValueFor(headerParam)
		}
	}
	return headers
}

func (o ProviderClient) getResourceURL(resource SpecResource) (string, error) {
	host, err := o.openAPIBackendConfiguration.getHost()
	if err != nil {
		return "", err
	}

	basePath := o.openAPIBackendConfiguration.getBasePath()
	resourceRelativePath := resource.getResourcePath()

	// Fall back to override the host if value is not empty; otherwise global host will be used as usual
	hostOverride, err := resource.getHost()
	if err != nil {
		return "", err
	}
	if hostOverride != "" {
		log.Printf("[INFO] resource '%s' is configured with host override, API calls will be made against '%s' instead of '%s'", resourceRelativePath, hostOverride, host)
		host = hostOverride
	}

	if host == "" || resourceRelativePath == "" {
		return "", fmt.Errorf("host and path are mandatory attributes to get the resource URL - host['%s'], path['%s']", host, resourceRelativePath)
	}

	// TODO: use resource operation schemes if specified
	defaultScheme := "http"
	for _, scheme := range o.openAPIBackendConfiguration.getHTTPSchemes() {
		if scheme == "https" {
			defaultScheme = "https"
		}
	}
	path := resourceRelativePath
	if strings.Index(resourceRelativePath, "/") != 0 {
		path = fmt.Sprintf("/%s", resourceRelativePath)
	}

	if basePath != "" && basePath != "/" {
		if strings.Index(basePath, "/") == 0 {
			return fmt.Sprintf("%s://%s%s%s", defaultScheme, host, basePath, path), nil
		}
		return fmt.Sprintf("%s://%s/%s%s", defaultScheme, host, basePath, path), nil
	}
	return fmt.Sprintf("%s://%s%s", defaultScheme, host, path), nil
}

func (o ProviderClient) getResourceIDURL(resource SpecResource, id string) (string, error) {
	url, err := o.getResourceURL(resource)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", url, id), nil
}
