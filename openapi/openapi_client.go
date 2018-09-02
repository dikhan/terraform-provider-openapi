package openapi

import (
	"fmt"
	"github.com/dikhan/http_goclient"
	"log"
	"net/http"
	"strings"
)

// OpenAPIClient defines a client that is configured based on the OpenAPI server side documentation
// The CRUD operations accept an OpenAPI operation which defines among other things the security scheme applicable to
// the API when making the HTTP request
type OpenAPIClient struct {
	openAPIBackendConfiguration openAPIBackendConfiguration
	httpClient                  http_goclient.HttpClient
	providerConfig              providerConfig
	apiAuthenticator            apiAuthenticator
}

// Post performs a post request to the server API operation based on its definition (e,g: auth required)
func (o *OpenAPIClient) Post(path string, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	resourceURL, err := o.getResourceURL(path)
	if err != nil {
		return nil, err
	}
	operationSecuritySchemes := o.openAPIBackendConfiguration.getSecurityForPathPostOperation(path)
	operationHeaders := o.openAPIBackendConfiguration.getHeaderParamsForPathPostOperation(path)
	return o.performRequest(HttpPost, resourceURL, operationSecuritySchemes, operationHeaders, requestPayload, responsePayload)
}

func (o *OpenAPIClient) Put(path, id string, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(path, id)
	if err != nil {
		return nil, err
	}
	operationSecuritySchemes := o.openAPIBackendConfiguration.getSecurityForPathPutOperation(path)
	operationHeaders := o.openAPIBackendConfiguration.getHeaderParamsForPathPutOperation(path)
	return o.performRequest(HttpPut, resourceURL, operationSecuritySchemes, operationHeaders, requestPayload, responsePayload)
}

func (o *OpenAPIClient) Get(path, id string, responsePayload interface{}) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(path, id)
	if err != nil {
		return nil, err
	}
	operationSecuritySchemes := o.openAPIBackendConfiguration.getSecurityForPathGetOperation(path)
	operationHeaders := o.openAPIBackendConfiguration.getHeaderParamsForPathGetOperation(path)
	return o.performRequest(HttpGet, resourceURL, operationSecuritySchemes, operationHeaders, nil, responsePayload)
}

func (o *OpenAPIClient) Delete(path, id string) (*http.Response, error) {
	resourceURL, err := o.getResourceIDURL(path, id)
	if err != nil {
		return nil, err
	}
	operationSecuritySchemes := o.openAPIBackendConfiguration.getSecurityForPathDeleteOperation(path)
	operationHeaders := o.openAPIBackendConfiguration.getHeaderParamsForPathDeleteOperation(path)
	return o.performRequest(HttpDelete, resourceURL, operationSecuritySchemes, operationHeaders, nil, nil)
}

func (o *OpenAPIClient) performRequest(method HttpMethodSupported, resourceURL string, operationSecuritySchemes []map[string][]string, operationHeaders []HeaderParam, requestPayload interface{}, responsePayload interface{}) (*http.Response, error) {
	reqContext, err := o.apiAuthenticator.prepareAuth(method, resourceURL, operationSecuritySchemes, o.providerConfig)
	if err != nil {
		return nil, err
	}
	reqContext.headers = o.appendOperationHeaders(operationHeaders, o.providerConfig, reqContext.headers)
	log.Printf("[DEBUG] Performing %s %s", method, reqContext.url)

	switch method {
	case HttpPost:
		return o.httpClient.PostJson(reqContext.url, reqContext.headers, requestPayload, &responsePayload)
	case HttpPut:
		return o.httpClient.PutJson(reqContext.url, reqContext.headers, requestPayload, &responsePayload)
	case HttpGet:
		return o.httpClient.Get(reqContext.url, reqContext.headers, &responsePayload)
	case HttpDelete:
		return o.httpClient.Delete(reqContext.url, reqContext.headers)
	}
	return nil, fmt.Errorf("method '%s' not supported", method)
}

// appendOperationHeaders returns a maps containing the headers passed in and adds whatever headers the operation requires. The values
// are retrieved from the provider configuration.
func (o OpenAPIClient) appendOperationHeaders(operationHeaders []HeaderParam, providerConfig providerConfig, headers map[string]string) map[string]string {
	if operationHeaders != nil && len(operationHeaders) > 0 {
		for _, headerParam := range operationHeaders {
			// Setting the actual name of the header with the value coming from the provider configuration
			headers[headerParam.Name] = providerConfig.Headers[headerParam.GetHeaderTerraformConfigurationName()]
		}
	}
	return headers
}

func (o OpenAPIClient) getResourceURL(resourceRelativePath string) (string, error) {
	host := o.openAPIBackendConfiguration.getHost()
	basePath := o.openAPIBackendConfiguration.getBasePath()

	if host == "" || resourceRelativePath == "" {
		return "", fmt.Errorf("host and path are mandatory attributes to get the resource URL - host['%s'], path['%s']", host, resourceRelativePath)
	}
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

func (o OpenAPIClient) getResourceIDURL(resourceRelativePath, id string) (string, error) {
	url, err := o.getResourceURL(resourceRelativePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", url, id), nil
}
