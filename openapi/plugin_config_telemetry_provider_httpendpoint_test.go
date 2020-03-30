package openapi

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTelemetryProviderHttpEndpoint_Validate(t *testing.T) {
	testCases := []struct {
		testName    string
		url         string
		expectedErr error
	}{
		{
			testName:    "happy path - host and port populated",
			url:         "http://telemetry.myhost.com/v1/metrics",
			expectedErr: nil,
		},
		{
			testName:    "url is empty",
			url:         "",
			expectedErr: errors.New("http endpoint telemetry configuration is missing a value for the 'url property'"),
		},
		{
			testName:    "url is wrongly formatter",
			url:         "htop://something-wrong.com",
			expectedErr: errors.New("http endpoint telemetry configuration does not have a valid URL 'htop://something-wrong.com'"),
		},
	}

	for _, tc := range testCases {
		tpg := TelemetryProviderHttpEndpoint{
			URL: tc.url,
		}
		err := tpg.Validate()
		assert.Equal(t, tc.expectedErr, err, tc.testName)
	}
}
