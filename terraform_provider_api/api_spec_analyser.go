package terraform_provider_api

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

const RESOURCE_NAME_REGEX = "(/\\w*/)+{.*}"
const RESTFUL_ENDPOINT_REGEX = "((?:.*)){.*}"

type SpecAnalyser interface {
}

type ApiSpecAnalyser struct {
	d *loads.Document
}

func (a ApiSpecAnalyser) getCrudResources() CrudResourcesInfo {
	resources := CrudResourcesInfo{}
	for pathName, pathItem := range a.d.Spec().Paths.Paths {
		if !a.isEndPointCrudCompliant(pathName, a.d.Spec().Paths.Paths) {
			continue
		}
		rootPath := a.findMatchingRootPath(pathName)
		ref := a.d.Spec().Paths.Paths[rootPath].Post.Parameters[0].Schema.Ref.String()
		resourceName := a.getResourceName(pathName)
		r := ResourceInfo{
			Name:             resourceName,
			Host:             a.d.Spec().Host,
			SchemaDefinition: a.d.Spec().Definitions[a.getRefName(ref)],
			CreatePathInfo:   a.d.Spec().Paths.Paths[rootPath],
			PathInfo:         pathItem,
		}
		resources[resourceName] = r
	}
	return resources
}

func (a ApiSpecAnalyser) getRefName(ref string) string {
	reg, err := regexp.Compile("(\\w+)[^//]*$")
	if err != nil {
		log.Fatalf("something really wrong happened if the ref reg can't be compiled...")
	}
	return reg.FindStringSubmatch(ref)[0]
}

// isEndPointCrudCompliant returns true only if the path given 'p' exposes all CRUD operations.
// For instance, if p was "/users/{username}" and paths contained the following entries and implementations:
// - "/users"
// 		- POST
// - "/users/{username}"
// 		- GET
// 		- PUT
// 		- DELETE
// then the expected returned value is true. Otherwise if the above criteria is not met, it is considered that
// the path provided is not fully compliant.
func (f ApiSpecAnalyser) isEndPointCrudCompliant(p string, paths map[string]spec.PathItem) bool {
	if f.isPotentialCrudEndPoint(p) {
		endPoint := paths[p]
		if endPoint.Get == nil || endPoint.Put == nil || endPoint.Delete == nil {
			return false
		}
		endPointRootPath := f.findMatchingRootPath(p)
		if endPointRootPath == "" {
			log.Printf("could not find root path for end point - %+s", p)
			return false
		}
		postExist, err := f.postIsPresent(endPointRootPath, paths)
		if err != nil {
			log.Println(err)
			return false
		}
		return postExist
	}
	return false
}

// postIsPresent checks if a given path has a POST implementation. The given path
func (f ApiSpecAnalyser) postIsPresent(p string, paths map[string]spec.PathItem) (bool, error) {
	b := paths[p]
	if &b == nil || b.Post == nil {
		return false, fmt.Errorf("end point %s missing POST operation", p)
	}
	return true, nil
}

// restfulEndPointRegex loads up the regex specified in const RESTFUL_ENDPOINT_REGEX
// If the regex is not able to compile the regular expression the function exists calling os.Exit(1) as
// there is the regex is completely busted
func (f ApiSpecAnalyser) restfulEndPointRegex() *regexp.Regexp {
	r, err := regexp.Compile(RESTFUL_ENDPOINT_REGEX)
	if err != nil {
		log.Fatalf("Something is really wrong with the resful endpoint regex [%s] %s", RESTFUL_ENDPOINT_REGEX, err)
	}
	return r
}

// isPotentialCrudEndPoint checks if the given path is of form /resource/{id}
func (f ApiSpecAnalyser) isPotentialCrudEndPoint(p string) bool {
	r := f.restfulEndPointRegex()
	return r.MatchString(p)
}

// isPotentialCrudEndPoint checks if the given path is of form /resource/{id}
func (f ApiSpecAnalyser) getResourceName(p string) string {
	r, err := regexp.Compile(RESOURCE_NAME_REGEX)
	if err != nil {
		log.Fatalf("Something is really wrong with the resource name regex [%s] %s", RESOURCE_NAME_REGEX, err)
	}
	return strings.Replace(r.FindStringSubmatch(p)[1], "/", "", -1)
}

// findMatchingRootPath returns the corresponding root path for a given endpoint
// Example: Given 'p' being "/users/{username}" the result will be "/users"
// Otherwise an error is returned
func (f ApiSpecAnalyser) findMatchingRootPath(p string) string {
	r := f.restfulEndPointRegex()
	result := r.FindStringSubmatch(p)
	if len(result) != 2 {
		return ""
	}
	return strings.TrimRight(result[1], "/")
}
