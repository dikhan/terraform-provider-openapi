package openapi

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTelemetryProviderGraphite_Validate(t *testing.T) {
	testCases := []struct {
		testName    string
		host        string
		port        int
		expectedErr error
	}{
		{
			testName: "happy path - host and port populated",
			host:     "telemetry.myhost.com",
			port:     8125,
		},
		{
			testName:    "crappy path - host is empty",
			host:        "",
			port:        8125,
			expectedErr: errors.New("graphite telemetry configuration is missing a value for the 'host property'"),
		},
		{
			testName:    "crappy path - port is 0",
			host:        "telemetry.myhost.com",
			port:        0,
			expectedErr: errors.New("graphite telemetry configuration is missing a valid value (>0) for the 'port' property'"),
		},
	}

	for _, tc := range testCases {
		tpg := TelemetryProviderGraphite{
			Host: tc.host,
			Port: tc.port,
		}
		err := tpg.Validate()
		assert.Equal(t, tc.expectedErr, err, tc.testName)
	}
}
