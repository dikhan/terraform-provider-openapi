package openapiutils

import (
	"regexp"
)

const fqdnInURLRegex = `\b(?:(?:[^.-/]{0,1})[\w-]{1,63}[-]{0,1}[.]{1})+(?:[a-zA-Z]{2,63})?|localhost(?:[:]\d+)?\b`

// GetHostFromURL returns the fqdn of a given string (localhost including port number is also handled).
// Example domains that would match:
// - http://domain.com/
// - domain.com/parameter
// - domain.com?anything
// - example.domain.com
// - example.domain-hyphen.com
// - www.domain.com
// - localhost
// - localhost:8080
// Example domains that would not match:
// - http://domain.com:8080 (this use case is not support at the moment, it is assumed that actual domains will use standard ports)
func GetHostFromURL(url string) string {
	re := regexp.MustCompile(fqdnInURLRegex)
	return re.FindString(url)
}
