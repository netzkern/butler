package space

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/netzkern/butler/commands/confluence"
	"github.com/pinzolo/casee"

	logy "github.com/apex/log"
	"github.com/pkg/errors"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

type (
	// request the request payload
	request struct {
		Key         string      `json:"key"`
		Name        string      `json:"name"`
		Description requestDesc `json:"description"`
	}
	requestDesc struct {
		Plain requestContent `json:"plain"`
	}
	requestContent struct {
		Value          string `json:"value"`
		Representation string `json:"representation"`
	}
	// Response the response payload
	Response struct {
		ID          int    `json:"id"`
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description struct {
			Plain struct {
				Value          string `json:"value"`
				Representation string `json:"representation"`
			} `json:"plain"`
		} `json:"description"`
		metadata interface{}
		Links    struct {
			Collection string `json:"collection"`
			Base       string `json:"base"`
			Context    string `json:"context"`
			Self       string `json:"self"`
		} `json:"_links"`
	}
	// Option function.
	Option func(*Space)
	// Space command to create git hooks
	Space struct {
		client      *confluence.Client
		endpoint    *url.URL
		commandData *CommandData
	}
	// CommandData contains all command related data
	CommandData struct {
		Key         string
		Name        string
		Description string
		Public      bool
	}
)

var (
	errBadRequest   = errors.New("there is already a space with the given key, or no property value was provided, or the value is too long")
	errForbidden    = errors.New("the user does not have permission to edit the space with the given key")
	errUnauthorized = errors.New("the user does not have permission to create the space")
	errNotFound     = errors.New("the user does not have permission to view the requested content")
)

// NewSpace with the given options.
func NewSpace(options ...Option) *Space {
	v := &Space{}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithClient option.
func WithClient(client *confluence.Client) Option {
	return func(c *Space) {
		c.client = client
	}
}

// WithEndpoint option.
func WithEndpoint(location string) Option {
	return func(c *Space) {
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
	return func(g *Space) {
		g.commandData = cd
	}
}

// StartCommandSurvey collect all required informations from user
func (s *Space) StartCommandSurvey() error {
	var cmd = &CommandData{}

	// start command prompts
	err := survey.Ask(s.getQuestions(), cmd)
	if err != nil {
		return err
	}

	cmd.Key = buildSpaceKey(cmd.Name)
	s.commandData = cmd

	return nil
}

// getQuestions return all required prompts
func (s *Space) getQuestions() []*survey.Question {
	qs := []*survey.Question{
		{
			Name: "Name",
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(1),
				survey.MaxLength(255),
			),
			Prompt: &survey.Input{
				Message: "What's the name of the space?",
			},
		},
		{
			Name: "Description",
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(1),
				survey.MaxLength(255),
			),
			Prompt: &survey.Input{
				Message: "What's the description of the space?",
			},
		},
		{
			Name: "Public",
			Prompt: &survey.Confirm{
				Message: "Do you want to create a public space?",
			},
		},
	}

	return qs
}

// https://docs.atlassian.com/atlassian-confluence/REST/6.6.0/#space-createSpace
func (s *Space) create(reqBody *request) (*Response, error) {
	jsonbody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := s.endpoint.String()

	if s.commandData.Public {
		endpoint += "/space"
	} else {
		endpoint += "/space/_private"
	}

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

	logy.Debugf("create space request returned (%s)", resp.Status)

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

	var space Response
	err = json.Unmarshal(resp.Payload, &space)
	if err != nil {
		return nil, err
	}

	return &space, nil
}

// Run the command
func (s *Space) Run() (*Response, error) {
	req := &request{
		Key:  s.commandData.Key,
		Name: s.commandData.Name,
		Description: requestDesc{
			Plain: requestContent{
				Representation: "plain",
				Value:          s.commandData.Description,
			},
		},
	}

	logy.Debugf("create space request payload: %+v\n", req)

	return s.create(req)
}

// buildSpaceKey create a space key from a string
func buildSpaceKey(spaceName string) string {
	return casee.ToCamelCase(spaceName)
}
