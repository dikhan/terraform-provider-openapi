package openapi

import (
	"errors"
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi/openapierr"
	"github.com/hashicorp/terraform/helper/schema"
	"io/ioutil"
	"net/http"
)

func checkHTTPStatusCode(openAPIResource SpecResource, res *http.Response, expectedHTTPStatusCodes []int) error {
	if !responseContainsExpectedStatus(expectedHTTPStatusCodes, res.StatusCode) {
		var resBody string
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("[resource='%s'] HTTP Reponse Status Code %d - Error '%s' occurred while reading the response body", openAPIResource.getResourceName(), res.StatusCode, err)
		}
		if b != nil && len(b) > 0 {
			resBody = string(b)
		}
		switch res.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d - Unauthorized: API access is denied due to invalid credentials (%s)", openAPIResource.getResourceName(), res.StatusCode, resBody)
		case http.StatusNotFound:
			return &openapierr.NotFoundError{OriginalError: fmt.Errorf("HTTP Reponse Status Code %d - Not Found. Could not find resource instance: %s", res.StatusCode, resBody)}
		default:
			return fmt.Errorf("[resource='%s'] HTTP Response Status Code %d not matching expected one %v (%s)", openAPIResource.getResourceName(), res.StatusCode, expectedHTTPStatusCodes, resBody)
		}
	}
	return nil
}

func responseContainsExpectedStatus(expectedStatusCodes []int, responseStatusCode int) bool {
	for _, expectedStatusCode := range expectedStatusCodes {
		if expectedStatusCode == responseStatusCode {
			return true
		}
	}
	return false
}

func getParentIDsAndResourcePath(openAPIResource SpecResource, data *schema.ResourceData) (parentIDs []string, resourcePath string, err error) {
	parentIDs, err = getParentIDs(openAPIResource, data)
	if err != nil {
		return nil, "", err
	}
	resourcePath, err = openAPIResource.getResourcePath(parentIDs)
	if err != nil {
		return nil, "", err
	}
	return
}

func getParentIDs(openAPIResource SpecResource, data *schema.ResourceData) ([]string, error) {
	if openAPIResource == nil {
		return []string{}, errors.New("can't get parent ids from a resourceFactory with no openAPIResource")
	}

	parentResourceInfo := openAPIResource.getParentResourceInfo()
	if parentResourceInfo != nil {
		parentResourceNames := parentResourceInfo.getParentPropertiesNames()

		parentIDs := []string{}
		for _, parentResourceName := range parentResourceNames {
			parentResourceID := data.Get(parentResourceName)
			if parentResourceID == nil {
				return nil, fmt.Errorf("could not find ID value in the state file for subresource parent property '%s'", parentResourceName)
			}
			parentIDs = append(parentIDs, parentResourceID.(string))
		}
		return parentIDs, nil
	}

	return []string{}, nil
}
