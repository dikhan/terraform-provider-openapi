package version

import (
	"fmt"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func BuildUserAgent(runtime, arch string) string {
	return fmt.Sprintf("OpenAPI Terraform Provider/%s-%s (%s/%s)", Version, Commit, runtime, arch)
}