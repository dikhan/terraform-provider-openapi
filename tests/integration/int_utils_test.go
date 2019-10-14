package integration

import (
	"fmt"
	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/go-openapi/loads"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func createPluginConfigFile(content string) *os.File {
	file, err := ioutil.TempFile("", "terraform-provider-openapi.yaml")
	if err != nil {
		log.Fatal(err)
	}
	file.Write([]byte(content))
	return file
}

func floatToString(number float32) string {
	if number == 0 {
		return "0"
	}
	return fmt.Sprintf("%.2f", number)
}

func arrayToString(value []string) string {
	var result = "["
	for _, v := range value {
		result += fmt.Sprintf("%s,", v)
	}
	result = strings.TrimRight(result, ",")
	result += "]"
	return result
}

// Check if resource DOES NOT exists remotely
func testAccCheckResourceDoesNotExist(openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName string) resource.TestCheckFunc {
	return testAccCheckWhetherResourceExist(openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName, false)
}

// Check if resource exists remotely
func testAccCheckResourceExist(openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName string) resource.TestCheckFunc {
	return testAccCheckWhetherResourceExist(openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName, true)
}

func testAccCheckWhetherResourceExist(openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName string, resourceShouldExist bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		err := testCheckDestroy(s, openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName)

		if err != nil {
			if resourceShouldExist {
				if strings.Contains(err.Error(), "still exists") {
					return nil
				}
			}
			return err
		}

		// if resourceShouldExist is set to false and we reach this point; returning nil as the testCheckDestroy would have return nil which means that the resource no longer exists which fulfils this premise
		if !resourceShouldExist {
			return nil
		}

		// if resourceShouldExist is set to true and we reach this point; the premise will not be fulfilled as the expectation was for the resource to exist
		return fmt.Errorf("resource no longer exists")
	}
}

// Acceptance test resource-destruction for openapi_{resourceName}:
//
// Check all resources of the type specified in the configuration have been destroyed.
func testCheckDestroy(state *terraform.State, openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName string) error {
	return testCheckDestroyWithDelay(state, openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName, 0)
}

// Acceptance test resource-destruction for openapi_{resourceName}:
//
// Check all resources of the type specified in the configuration have been destroyed but delay the check {delayCheck} seconds
func testCheckDestroyWithDelay(state *terraform.State, openAPIResourceName, resourceName, resourcePath, resourceSchemaDefinitionName string, delayCheck int) error {
	for _, res := range state.RootModule().Resources {
		if res.Type != openAPIResourceName {
			continue
		}
		cdnID := res.Primary.ID
		openAPIClient := testAccProvider.Meta().(openapi.ClientOpenAPI)
		apiSpec, err := loads.JSONSpec(otfVarSwaggerURLEnvVariableValue)
		if err != nil {
			return err
		}

		instancePath := fmt.Sprintf("%s/{id}", resourcePath)

		specResource := &openapi.SpecV2Resource{
			Name:             resourceName,
			Path:             resourcePath,
			SchemaDefinition: apiSpec.Spec().Definitions[resourceSchemaDefinitionName],
			InstancePathItem: apiSpec.Spec().Paths.Paths[instancePath],
			RootPathItem:     apiSpec.Spec().Paths.Paths[resourcePath],
		}

		time.Sleep(time.Duration(delayCheck) * time.Second)

		resp, err := openAPIClient.Get(specResource, cdnID, nil)
		if err != nil {
			return err
		}
		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("%s '%s' still exists", resourceName, cdnID)
		}
	}
	return nil
}

func testAccPreCheck(t *testing.T) {
	versionEndpoint := "https://localhost:8443/version"
	res, err := http.Get(versionEndpoint)
	if err != nil {
		t.Fatalf("error occured when verifying if the API is up and running: %s", err)
	}
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned not expected response status code %d", versionEndpoint, res.StatusCode)
	}
}
