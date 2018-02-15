package confluence

import (
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
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

func (c *Client) sendRequest(req *http.Request) ([]byte, error) {
	req.Header.Add("Accept", "application/json, */*")
	c.authMethod.auth(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPartialContent:
		res, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		return res, nil
	case http.StatusNoContent, http.StatusResetContent:
		return nil, nil
	case http.StatusUnauthorized:
		return nil, errors.New("authentication failed")
	case http.StatusServiceUnavailable:
		return nil, errors.Errorf("service is not available (%s)", resp.Status)
	case http.StatusInternalServerError:
		return nil, errors.Errorf("Internal server error: %s", resp.Status)
	}

	return nil, errors.Wrapf(err, "unknown response status %s", resp.Status)
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
