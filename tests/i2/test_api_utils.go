package i2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertExpectedRequestURI(t *testing.T, expectedRequestURI string, r *http.Request) {
	if r.RequestURI != expectedRequestURI {
		assert.Fail(t, fmt.Sprintf("%s request URI '%s' does not match the expected one '%s'", r.Method, r.RequestURI, expectedRequestURI))
	}
}

func apiPostResponse(t *testing.T, responseBody string, w http.ResponseWriter, r *http.Request) {
	apiResponse(t, responseBody, http.StatusCreated, w, r)
}

func apiGetResponse(t *testing.T, responseBody string, w http.ResponseWriter, r *http.Request) {
	apiResponse(t, responseBody, http.StatusOK, w, r)
}

func apiDeleteResponse(t *testing.T, w http.ResponseWriter, r *http.Request) {
	apiResponse(t, "", http.StatusNoContent, w, r)
}

func apiResponse(t *testing.T, responseBody string, httpResponseStatusCode int, w http.ResponseWriter, r *http.Request) {
	if r.Body != nil {
		bs, e := ioutil.ReadAll(r.Body)
		require.NoError(t, e)
		fmt.Printf("%s request body >>> %s\n", r.Method, string(bs))
	}
	w.WriteHeader(httpResponseStatusCode)
	if responseBody != "" {
		w.Write([]byte(responseBody))
	}
}
