# HTTP Go Client

Http Go Client represents an HTTP wrapper that reduces the boiler plate needed to marshall/un-marshall request/response
bodies by providing friendly CRUD operations matching their corresponding HTTP verbs that allow in/out interfaces.

The client supports CRUD operations in the form of GET, POST, PUT and DELETE. The client also supports convenient json 
operations like PutJson/PostJson so the user does not need to set up the content type of the payload.  

Additionally, the client provides a handy testing package that allows for the creation of a fake http client and server.
The server side can be configured with a request matcher that verifies whether the incoming request matches the expected 
one and if so replies with a configured (payload and HTTP code response) HTTP response. 

# How to use the client?

Http Go Client is really easy to use, it is first configured with the actual underlying http.client which is the one
that eventually performs the http calls.

Here is an example on how you to build the struct:

```
import (
	"github.com/dikhan/http_goclient"
)

func main() {
    httpClient := &HttpClient{&http.Client{}}	
}
```

Once the client is created, the different CRUD operations available can be invoked:

``` 
    res, err := c.httpClient.Get(url, c.requestHeaders(), in)
    res, err := c.httpClient.PostJson(url, c.requestHeaders(), in, out)
    res, err := c.httpClient.PutJson(url, c.requestHeaders(), in, out)
    res, err := c.httpClient.Delete(url, c.requestHeaders())
```

Please note that this client does not inspect the response at all. The user of the library should take care of the
different http response codes returned by using the *http.Response object returned by the CRUD operations. The following
[logentries_goclient](https://github.com/dikhan/logentries_goclient/blob/master/log_entries_client.go) can serve as an 
example on how to use the library.

# How to use the client's testing package?

The following snippet of code shows how to create a new RequestMatcher, initialize the clientServer struct and subsequently 
how to instantiate the TestClientServer().
 

```
import (
	"github.com/dikhan/http_goclient/testutils"
)

type TestStruct struct {
	Name string `json:"name"`
	Username string `json:"username"`
}

func TestLogSets_GetLogSets(t *testing.T) {
    expectedPayload := &TestStruct{
        Name: "Dani",
        Username: "dikhan",
    }

    requestMatcher := testutils.NewRequestMatcher(http.MethodGet, "/api/resource", nil, http.StatusOK, expectedPayload)
    
	testClientServer := testutils.TestClientServer {
		RequestMatcher: requestMatcher,
	}
	httpClient, httpServer := testClientServer.TestClientServer()
    ...
    // init the struct which depends on *http.Client
}
```

The client will end up sending the HTTP request to the 'fake' server which will perform some verifications on the
incoming request (as configured previously), and reply with the expected response. 

The following [tests from the logentries_goclient](https://github.com/dikhan/logentries_goclient/blob/master/logs_test.go#L42) 
can be used as reference to better understand how to use the testing package.


## Contributing

- Fork it!
- Create your feature branch: git checkout -b my-new-feature
- Commit your changes: git commit -am 'Add some feature'
- Push to the branch: git push origin my-new-feature
- Submit a pull request :D

## Authors

Daniel I. Khan Ramiro

See also the list of [contributors](https://github.com/dikhan/http_goclient/graphs/contributors) who participated in this project.