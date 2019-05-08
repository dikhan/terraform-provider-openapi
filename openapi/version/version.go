package version

import (
	"fmt"
)

var (
	// Version specifies the version of the OpenAPI Terraform provider
	Version = "dev"
	// Commit specifies the commit hash of the OpenAPI Terraform provider at the time of building the binary
	Commit = "none"
	// Date specifies the data which the binary was build
	Date = "unknown"
)

// BuildUserAgent creates based on the Version, Commit, runtime and arch info the user agent string that will
// be send along all the API requests
func BuildUserAgent(runtime, arch string) string {
	return fmt.Sprintf("OpenAPI Terraform Provider/%s-%s (%s/%s)", Version, Commit, runtime, arch)
}
