package confluence

import (
	"io/ioutil"
	"net/http"
)

type (
	// Client command to create git hooks
	Client struct {
		authMethod AuthMethod
		client     *http.Client
	}
	// ClientOption function.
	ClientOption func(*Client)
	// AuthMethod the authentication interface
	AuthMethod interface {
		auth(req *http.Request)
	}
	// RequestResult represent the result of the json request
	RequestResult struct {
		StatusCode int
		Status     string
		Payload    []byte
	}
)

// NewClient with the given options.
func NewClient(options ...ClientOption) *Client {
	v := &Client{client: &http.Client{}}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithAuth option.
func WithAuth(auth AuthMethod) ClientOption {
	return func(c *Client) {
		c.authMethod = auth
	}
}

// sendRequest make a request with an auhentication schema and
// return the whole request
func (c *Client) sendRequest(req *http.Request) (*RequestResult, error) {
	result := &RequestResult{}
	req.Header.Add("Accept", "application/json, */*")
	c.authMethod.auth(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return result, err
	}

	res, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	result.Payload = res
	result.StatusCode = resp.StatusCode
	result.Status = resp.Status

	if err != nil {
		return result, err
	}

	return result, nil
}

type basicAuthCallback func() (username, password string)

func (cb basicAuthCallback) auth(req *http.Request) {
	username, password := cb()
	req.SetBasicAuth(username, password)
}

// BasicAuth method
func BasicAuth(username, password string) AuthMethod {
	return basicAuthCallback(func() (string, string) { return username, password })
}
