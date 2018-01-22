package confluence

import (
	"encoding/json"
	"fmt"

	logy "github.com/apex/log"
	"github.com/parnurzeal/gorequest"
)

// https://docs.atlassian.com/atlassian-confluence/REST/6.5.0

type (
	Confluence struct {
		Space Space
		Auth  BasicAuth
	}
	BasicAuth struct {
		Username, Password string
	}
	Space struct {
		Name string `json:"name"`
	}
)

// Option function.
type Option func(*Confluence)

// New Confluence command
func New(options ...Option) *Confluence {
	c := &Confluence{}

	for _, o := range options {
		o(c)
	}

	return c
}

// WithSpace option.
func WithSpace(space Space) Option {
	return func(c *Confluence) {
		c.Space = space
	}
}

// WithAuth option.
func WithAuth(auth BasicAuth) Option {
	return func(c *Confluence) {
		c.Auth = auth
	}
}

// Run the command
func (c *Confluence) Run() error {
	mJSON, _ := json.Marshal(c.Space)

	request := gorequest.New().SetBasicAuth(c.Auth.Username, c.Auth.Password)
	resp, _, errs := request.Post("http://example.com").
		Send(string(mJSON)).
		End()

	if len(errs) > 0 {
		logy.Errorf("could not create Space %+v", errs)
		return fmt.Errorf("could not create Space")
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("could not create Space Status %v", resp.StatusCode)
	}

	return nil
}
