package page

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	logy "github.com/apex/log"
	"github.com/netzkern/butler/commands/confluence"
	"github.com/pkg/errors"
)

type (
	// request the request payload
	request struct {
		Title     string         `json:"title"`
		Type      string         `json:"type"`
		Space     page           `json:"space"`
		Ancestors []pageAncestor `json:"ancestors"`
	}
	page struct {
		Key string `json:"key"`
	}
	pageAncestor struct {
		ID string `json:"id"`
	}
	// Response the response payload
	Response struct {
		ID     string `json:"id"`
		Type   string `json:"type"`
		Status string `json:"status"`
	}
	// Option function.
	Option func(*Page)
	// Page command to create git hooks
	Page struct {
		client      *confluence.Client
		endpoint    *url.URL
		CommandData *CommandData
	}
	// CommandData contains all command related data
	CommandData struct {
		AncestorID string
		Title      string
		Type       string
		SpaceKey   string
	}
)

var (
	errBadRequest   = errors.New("there is already a page with the given key, or no property value was provided, or the value is too long")
	errForbidden    = errors.New("the user does not have permission to edit the page in the space")
	errUnauthorized = errors.New("the user does not have permission to create the page in the space")
	errNotFound     = errors.New("the user does not have permission to view the requested content")
)

// NewPage with the given options.
func NewPage(options ...Option) *Page {
	v := &Page{}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithClient option.
func WithClient(client *confluence.Client) Option {
	return func(c *Page) {
		c.client = client
	}
}

// WithEndpoint option.
func WithEndpoint(location string) Option {
	return func(c *Page) {
		u, err := url.ParseRequestURI(location)
		if err != nil {
			panic(err)
		}
		u.Path += "rest/api"
		c.endpoint = u
	}
}

// WithCommandData option.
func WithCommandData(cd *CommandData) Option {
	return func(g *Page) {
		g.CommandData = cd
	}
}

// https://developer.atlassian.com/cloud/confluence/rest/#api-content-post
func (s *Page) create(reqBody *request) (*Response, error) {
	jsonbody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := s.endpoint.String() + "/content"

	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "endpoint could not be parsed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequest("POST", url.String(), strings.NewReader(string(jsonbody)))
	if err != nil {
		return nil, errors.Wrap(err, "request could not be created")
	}

	req.Header.Add("Content-Type", "application/json")

	req = req.WithContext(ctx)

	logy.Debugf("new create space request to %s", url.String())

	resp, err := s.client.SendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "request could not be executed")
	}

	logy.Debugf("create page request returned (%s)", resp.Status)

	// provide detail error messages
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, errBadRequest
	case http.StatusNotFound:
		return nil, errNotFound
	case http.StatusForbidden:
		return nil, errForbidden
	case http.StatusUnauthorized:
		return nil, errUnauthorized
	case http.StatusServiceUnavailable:
		return nil, errors.Errorf("service is not available (%s)", resp.Status)
	case http.StatusInternalServerError:
		return nil, errors.Errorf("internal server error: %s", resp.Status)
	}

	var page Response
	err = json.Unmarshal(resp.Payload, &page)
	if err != nil {
		return nil, err
	}

	return &page, nil
}

// Run the command
func (s *Page) Run() (*Response, error) {
	req := &request{
		Title: s.CommandData.Title,
		Type:  s.CommandData.Type,
		Space: page{Key: s.CommandData.SpaceKey},
	}

	if s.CommandData.AncestorID != "" {
		req.Ancestors = []pageAncestor{
			pageAncestor{ID: s.CommandData.AncestorID},
		}
	}

	logy.Debugf("create page request payload: %+v\n", req)

	return s.create(req)
}
