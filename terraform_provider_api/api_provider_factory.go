package terraform_provider_api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	"github.com/hashicorp/terraform/helper/schema"
	"log"
	"os"
	"regexp"
	"strings"
)

const RESTFUL_ENDPOINT_REGEX = "(/\\w*/)+{.*}"

type DynamicProviderFactory struct {
	Name string
}

func (f DynamicProviderFactory) createProviderDynamically() *schema.Provider {
	apiDiscoveryUrl := os.Getenv("API_DISCOVERY_URL")
	apiSpec, err := f.getApiSpecification(apiDiscoveryUrl)
	if err != nil {
		log.Fatalf("error occurred while retrieving api specification. Error=%s", err)
	}
	PrettyPrint(apiSpec.Spec())

	schemaProvider, err := f.generateProviderSchemaFromApiSpec(apiSpec)
	if err != nil {
		log.Fatalf("error occurred while creating schema provider. Error=%s", err)
	}
	return schemaProvider
}

func (f DynamicProviderFactory) getApiSpecification(apiDiscoveryUrl string) (*loads.Document, error) {
	if apiDiscoveryUrl == "" {
		return nil, errors.New("required param 'apiDiscoveryUrl' missing")
	}
	apiSpecDoc, err := loads.JSONDoc(apiDiscoveryUrl)
	if err != nil {
		return nil, fmt.Errorf("error occurred when retrieving api spec from %s. Error=%s", apiDiscoveryUrl, err)
	}
	// load embedded swagger file
	swaggerSpec, err := loads.Analyzed(apiSpecDoc, "")
	if err != nil {
		return nil, fmt.Errorf("could not load api spec from %s. Error=%s", apiDiscoveryUrl, err)
	}
	return swaggerSpec, nil
}

func (f DynamicProviderFactory) generateProviderSchemaFromApiSpec(d *loads.Document) (*schema.Provider, error) {
	resourceMap := map[string]*schema.Resource{}
	for pathName, _ := range d.Spec().Paths.Paths {
		if !f.isEndPointCrudCompliant(pathName, d.Spec().Paths.Paths) {
			log.Printf("end point %s does not implement all CRUD operations so it won't be considered as a resource", pathName)
			continue
		}
		log.Println("FOUND API that supports all CRUD operations", pathName)
		r := ResourceFactory{
			Schema: d.Spec().Definitions["User"], // TODO: get key name from $ref object
			Rud:    d.Spec().Paths.Paths[pathName],
			Create: d.Spec().Paths.Paths[f.findMatchingRootPath(pathName)],
		}
		resource := r.createSchemaResource()
		resourceName, _ := f.getResourceName(pathName)

		log.Printf("NEW REOUSRCE REGISTERED %s", resourceName)
		resourceMap[resourceName] = resource
	}
	provider := &schema.Provider{
		ResourcesMap: resourceMap,
	}
	return provider, nil
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
func (f DynamicProviderFactory) isEndPointCrudCompliant(p string, paths map[string]spec.PathItem) bool {
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
func (f DynamicProviderFactory) postIsPresent(p string, paths map[string]spec.PathItem) (bool, error) {
	b := paths[p]
	if &b == nil || b.Post == nil {
		return false, fmt.Errorf("end point %s missing POST operation", p)
	}
	return true, nil
}

// restfulEndPointRegex loads up the regex specified in const RESTFUL_ENDPOINT_REGEX
// If the regex is not able to compile the regular expression the function exists calling os.Exit(1) as
// there is the regex is completely busted
func (f DynamicProviderFactory) restfulEndPointRegex() *regexp.Regexp {
	r, err := regexp.Compile(RESTFUL_ENDPOINT_REGEX)
	if err != nil {
		log.Fatalf("Something is really wrong with the resful endpoint regex [%s] %s", RESTFUL_ENDPOINT_REGEX, err)
	}
	return r
}

// isPotentialCrudEndPoint checks if the given path is of form /rootPath/{id}
func (f DynamicProviderFactory) isPotentialCrudEndPoint(p string) bool {
	r := f.restfulEndPointRegex()
	return r.MatchString(p)
}

// findMatchingRootPath returns the corresponding root path for a given endpoint
// Example: Given 'p' being "/users/{username}" the result will be "/users"
// Otherwise an error is returned
func (f DynamicProviderFactory) findMatchingRootPath(p string) string {
	r := f.restfulEndPointRegex()
	result := r.FindStringSubmatch(p)
	if len(result) != 2 {
		return ""
	}
	return strings.TrimRight(result[1], "/")
}

func (f DynamicProviderFactory) getResourceName(p string) (string, error) {
	r := f.findMatchingRootPath(p)
	resourceName := fmt.Sprintf("%s_%s", f.Name, strings.TrimLeft(r, "/"))
	return resourceName, nil
}

func PrettyPrint(v interface{}) {
	b, _ := json.MarshalIndent(v, "", "  ")
	log.Printf(string(b))
	log.Println()
}
