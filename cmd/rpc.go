package cmd

import (
	"net/http"
	"github.com/Azure/azure-storage-azcopy/common"
)

// Global singleton for sending RPC requests from the frontend to the STE
var Rpc func(cmd common.RpcCmd, request interface{}, response interface{}) error = NewHttpClient("").send

// NewHttpClient returns the instance of struct containing an instance of http.client and url
func NewHttpClient(url string) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{},
		url:    url,
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// todo : use url in case of string
type HTTPClient struct {
	client *http.Client
	url    string
}

// Send method on HttpClient sends the data passed in the interface for given command type to the client url
func (httpClient *HTTPClient) send(rpcCmd common.RpcCmd, requestData interface{}, responseData interface{}) error {
	// Create HTTP request with command in query parameter & request data as JSON payload
	requestJson, err := json.Marshal(v)
	if err != nil {
		fmt.Println(fmt.Sprintf("error marshalling request payload for command type %q", rpcCmd.String()))
		return err
	}
	request, err := http.NewRequest("POST", httpClient.url, bytes.NewReader(requestJson))
	// adding the commandType as a query param
	q := request.URL.Query()
	q.Add("commandType", rpcCmd.String())
	request.URL.RawQuery = q.Encode()

	response, err := httpClient.client.Do(request)
	if err != nil {
		return err
	}

	// Read response data, deserialie it and return it (via out responseData parameter) & error
	responseJson, err := ioutil.ReadAll(response.Body)
	response.Body.Close()
	if err != nil {
		fmt.Println("error reading response for the request")
		return err
	}
	if err = json.Unmarshal(responseJson, responseData); err != nil {
		panic(err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
