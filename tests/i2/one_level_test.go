package i2

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/hashicorp/terraform/helper/schema"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"github.com/dikhan/terraform-provider-openapi/openapi"
	"github.com/stretchr/testify/assert"
)

type fakeServiceSchemaPropertyConfiguration struct {
	openapi.ServiceSchemaPropertyConfiguration
}

type fakeServiceConfiguration struct {
	openapi.ServiceConfiguration
	getSwaggerURL func() string
}

func (c fakeServiceConfiguration) GetSwaggerURL() string {
	return c.getSwaggerURL()
}

func makeAPIServer(apiServerBehaviors map[string]http.HandlerFunc) (*httptest.Server, string) {
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("apiServer request>>>>", r.URL, r.Method)
		apiServerBehaviors[r.Method](w, r)
	}))
	fmt.Println("apiHost>>>>", apiServer.URL[7:])
	return apiServer, apiServer.URL[7:]
}

func mSW(swaggerfile string, apiHost string) func() *httptest.Server {
	return func() *httptest.Server {
		swaggerServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			swaggerReturned := fmt.Sprintf(swaggerfile, apiHost)
			fmt.Println("swaggerReturned>>>>", swaggerReturned)
			w.Write([]byte(swaggerReturned))
		}))
		fmt.Println("swaggerServer URL>>>>", swaggerServer.URL)
		return swaggerServer
	}
}

func createProvider(mSW func() *httptest.Server) *schema.Provider {
	swagServer := mSW()
	provider, e := openapi.CreateSchemaProviderFromServiceConfiguration(&openapi.ProviderOpenAPI{ProviderName: "openapi"}, fakeServiceConfiguration{
		getSwaggerURL: func() string {
			return swagServer.URL
		},
	})
	defer swagServer.Close()
	fmt.Println(e)
	return provider
}

func Test_OneLevel_CDN_Existing_CDN_and_Firewall_only_GET_are_sent(t *testing.T) {
	/*   API SERVER BEHAVIORS */
	apiServerBehaviors := map[string]http.HandlerFunc{}
	apiServerBehaviors[http.MethodGet] = func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/v1/cdns/42/v1/firewalls/1337":
			bs, e := ioutil.ReadAll(r.Body)
			require.NoError(t, e)
			fmt.Println("GET request body >>>", string(bs))
			apiResponse := `{"id":1337,"label":"FW #1337"}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(apiResponse))
		case "/v1/cdns/42":
			bs, e := ioutil.ReadAll(r.Body)
			require.NoError(t, e)
			fmt.Println("GET request body >>>", string(bs))
			apiResponse := `{"id":42,"label":"CDN #42"}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(apiResponse))
		default:
			assert.Fail(t, "rx unexpected GET to "+r.RequestURI)
		}
	}
	apiHost, apiHostURL := makeAPIServer(apiServerBehaviors)
	defer apiHost.Close()

	/* Provider creation based on the swagger passed as first arg in  mSW() */
	cdnSW := mSW(cdnSwaggerYAMLTemplate, apiHostURL)
	provider := createProvider(cdnSW)

	/* Assertion over the Provider created starting from the Swagger */
	assert.Nil(t, provider.ResourcesMap["openapi_cdn_v1"].Schema["id"]) //TODO: this needs to be not nil
	assert.NotNil(t, provider.ResourcesMap["openapi_cdn_v1"].Schema["label"])
	assert.Nil(t, provider.ResourcesMap["openapi_cdns_v1_firewalls_v1"].Schema["id"]) //TODO: this needs to be not nil
	assert.NotNil(t, provider.ResourcesMap["openapi_cdns_v1_firewalls_v1"].Schema["label"])
	assert.Nil(t, provider.ResourcesMap["openapi_cdns_v1_firewalls_v1"].Schema["cdn_v1_id"]) //TODO: this needs to be not nil

	/* TF file definition */
	tfFileContents := `# URI /v1/cdns/
		resource "openapi_cdn_v1" "my_cdn" {
		  label = "CDN #42"
		}
		# URI /v1/cdns/{parent_id}/v1/firewalls/
        resource "openapi_cdns_v1_firewalls_v1" "my_cdn_firewall_v1" {
           cdns_v1_id = openapi_cdn_v1.my_cdn.id
           label = "FW #1337"
        }`

	/* Assertion on Terraform operations using the given tfFileContent and the Provider above*/
	var testAccProviders = map[string]terraform.ResourceProvider{"openapi": provider}

	resource.Test(t, resource.TestCase{
		IsUnitTest:                true,
		PreCheck:                  nil,
		Providers:                 testAccProviders,
		CheckDestroy:              nil,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: tfFileContents,
				Check: resource.ComposeTestCheckFunc(
					//testAccCheckResourceExistCDN(),
					resource.TestCheckResourceAttr(
						"openapi_cdn_v1.my_cdn", "label", "CDN #42"),
					resource.TestCheckResourceAttr(
						"openapi_cdns_v1_firewalls_v1.my_cdn_firewall_v1", "cdns_v1_id", "42"),
				),
			},
		},
	})

}
