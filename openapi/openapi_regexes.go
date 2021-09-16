package openapi

const pathParameterRegex = "/({[\\w]*})*/"

// resourceVersionRegexTemplate is used to identify the version attached to the given resource. The parameter in the
// template will be replaced with the actual resource name so if there is a match the version grabbed is assured to belong
// to the resource in question and not any other version showing in the path before the resource name
const resourceVersionRegexTemplate = "/(v[\\d]*)/%s"

const resourceNameRegex = "((/[\\w-]*[/]?))+$"

// resourceParentNameRegex is the regex used to identify the different parents from a path that is a sub-resource. If used
// calling FindStringSubmatch, any match will contain the following groups in the corresponding array index:
// Index 0: This value will represent the full match containing also the path parameter (e,g: /v1/cdns/{id})
// Index 1: This value will represent the resource path (without the instance path parameter) - e,g: /v1/cdns
// Index 2: This value will represent version if it exists in the path (e,g: v1)
// Index 3: This value will represent the resource path name (e,g: cdns)
//
// - Example calling FindAllStringSubmatch with '/v1/cdns/{id}/v1/firewalls' path:
// matches, _ := resourceParentRegex.FindAllStringSubmatch("/v1/cdns/{id}/v1/firewalls", -1)
// matches[0][0]: Full match /v1/cdns/{id}
// matches[0][1]: Group 1. /v1/cdns
// matches[0][2]: Group 2. v1
// matches[0][3]: Group 3. cdns

// - Example calling FindAllStringSubmatch with '/v1/cdns/{id}/v2/firewalls/{id}/v3/rules' path
// matches, _ := resourceParentRegex.FindAllStringSubmatch("/v1/cdns/{id}/v2/firewalls/{id}/v3/rules", -1)
// matches[0][0]: Full match /v1/cdns/{id}
// matches[0][1]: Group 1. /v1/cdns
// matches[0][2]: Group 2. v1
// matches[0][3]: Group 3. cdns
// matches[1][0]: Full match /v2/firewalls/{id}
// matches[1][1]: Group 1. /v2/firewalls
// matches[1][2]: Group 2. v2
// matches[1][3]: Group 3. firewalls
const resourceParentNameRegex = `(\/(?:\w+\/)?(?:v\d+\/)?\w+)\/{\w+}`

const resourceInstanceRegex = "((?:.*)){.*}"
