package i2

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertExpectedRequestURI(t *testing.T, expectedRequestURI string, r *http.Request) {
	if r.RequestURI != expectedRequestURI {
		assert.Fail(t, fmt.Sprintf("%s request URI '%s' does not match the expected one '%s'", r.Method, r.RequestURI, expectedRequestURI))
	}
}
