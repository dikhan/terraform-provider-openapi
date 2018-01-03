package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
)

const resourceVersionRegex = "(/v[0-9]*/)"
const resourceNameRegex = "(/\\w*/)+{.*}"
const restfulEndpointRegex = "((?:.*)){.*}"

type apiSpecAnalyser struct {
	d *loads.Document
}

func (asa apiSpecAnalyser) getCrudResources() crudResourcesInfo {
	resources := crudResourcesInfo{}
	for pathName, pathItem := range asa.d.Spec().Paths.Paths {
		if !asa.isEndPointTerraformResourceCompliant(pathName, asa.d.Spec().Paths.Paths) {
			continue
		}
		rootPath := asa.findMatchingRootPath(pathName)
		ref := asa.d.Spec().Paths.Paths[rootPath].Post.Parameters[0].Schema.Ref.String()
		resourceName := asa.getResourceName(pathName)
		r := resourceInfo{
			name:             resourceName,
			basePath:         asa.d.BasePath(),
			path:             rootPath,
			host:             asa.d.Spec().Host,
			httpSchemes:      asa.d.Spec().Schemes,
			schemaDefinition: asa.d.Spec().Definitions[asa.getRefName(ref)],
			createPathInfo:   asa.d.Spec().Paths.Paths[rootPath],
			pathInfo:         pathItem,
		}
		resources[resourceName] = r
	}
	return resources
}

func (asa apiSpecAnalyser) getRefName(ref string) string {
	reg, err := regexp.Compile("(\\w+)[^//]*$")
	if err != nil {
		log.Fatalf("something really wrong happened if the ref reg can't be compiled...")
	}
	return reg.FindStringSubmatch(ref)[0]
}

// isEndPointTerraformResourceCompliant returns true only if the path given 'p' exposes POST and GET operations.
// PUT and DELETE are optional operations.
// For instance, if p was "/users/{username}" and paths contained the following entries and implementations:
// - "/users"
// 		- POST
// - "/users/{username}"
// 		- GET
// 		- PUT (optional)
// 		- DELETE (optional)
// then the expected returned value is true. Otherwise if the above criteria is not met, it is considered that
// the path provided is not terraform resource compliant.
func (asa apiSpecAnalyser) isEndPointTerraformResourceCompliant(p string, paths map[string]spec.PathItem) bool {
	if asa.isResourceInstanceEndPoint(p) {
		endPoint := paths[p]
		if endPoint.Get == nil {
			return false
		}
		endPointRootPath := asa.findMatchingRootPath(p)
		if endPointRootPath == "" {
			log.Printf("could not find root path for end point - %+s", p)
			return false
		}
		postExist, err := asa.postIsPresent(endPointRootPath, paths)
		if err != nil {
			log.Println(err)
			return false
		}
		return postExist
	}
	return false
}

// postIsPresent checks if a given path has a POST implementation. The given path
func (asa apiSpecAnalyser) postIsPresent(p string, paths map[string]spec.PathItem) (bool, error) {
	b := paths[p]
	if &b == nil || b.Post == nil {
		return false, fmt.Errorf("end point %s missing POST operation", p)
	}
	return true, nil
}

// restfulEndPointRegex loads up the regex specified in const restfulEndpointRegex
// If the regex is not able to compile the regular expression the function exists calling os.Exit(1) as
// there is the regex is completely busted
func (asa apiSpecAnalyser) restfulEndPointRegex() *regexp.Regexp {
	r, err := regexp.Compile(restfulEndpointRegex)
	if err != nil {
		log.Fatalf("Something is really wrong with the resful endpoint regex [%s] %s", restfulEndpointRegex, err)
	}
	return r
}

// isResourceInstanceEndPoint checks if the given path is of form /resource/{id}
func (asa apiSpecAnalyser) isResourceInstanceEndPoint(p string) bool {
	r := asa.restfulEndPointRegex()
	return r.MatchString(p)
}

// isResourceInstanceEndPoint checks if the given path is of form /resource/{id}
func (asa apiSpecAnalyser) getResourceName(p string) string {
	nameRegex, err := regexp.Compile(resourceNameRegex)
	if err != nil {
		log.Fatalf("Something is really wrong with the resource name regex [%s] %s", resourceNameRegex, err)
	}
	resourceName := strings.Replace(nameRegex.FindStringSubmatch(p)[1], "/", "", -1)

	versionRegex, err := regexp.Compile(resourceVersionRegex)
	if err != nil {
		log.Fatalf("Something is really wrong with the resource version regex [%s] %s", resourceVersionRegex, err)
	}
	versionMatches := versionRegex.FindStringSubmatch(p)
	if len(versionMatches) != 0 {
		version := strings.Replace(versionRegex.FindStringSubmatch(p)[1], "/", "", -1)
		resourceNameWithVersion := fmt.Sprintf("%s_%s", resourceName, version)
		return resourceNameWithVersion
	}
	return resourceName
}

// findMatchingRootPath returns the corresponding root path for a given endpoint
// Example: Given 'p' being "/users/{username}" the result will be "/users"
// Otherwise an error is returned
func (asa apiSpecAnalyser) findMatchingRootPath(p string) string {
	r := asa.restfulEndPointRegex()
	result := r.FindStringSubmatch(p)
	if len(result) != 2 {
		return ""
	}
	return strings.TrimRight(result[1], "/")
}
