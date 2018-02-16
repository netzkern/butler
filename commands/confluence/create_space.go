package confluence

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/pinzolo/casee"

	logy "github.com/apex/log"
	"github.com/pkg/errors"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

type (
	// spaceRequest the request payload
	spaceRequest struct {
		Key         string    `json:"key"`
		Name        string    `json:"name"`
		Description spaceDesc `json:"description"`
	}
	spaceDesc struct {
		Plain spaceDescPlain `json:"plain"`
	}
	spaceDescPlain struct {
		Value          string `json:"value"`
		Representation string `json:"representation"`
	}
	// SpaceResponse the response payload
	SpaceResponse struct {
		ID          string `json:"id"`
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
	// SpaceOption function.
	SpaceOption func(*Space)
	// Space command to create git hooks
	Space struct {
		client      *Client
		endpoint    *url.URL
		CommandData *CommandData
	}
	// CommandData contains all command related data
	CommandData struct {
		Key         string
		Name        string
		Description string
		Private     bool
	}
)

var (
	errBadRequest   = errors.New("the space already has a value with the given key, or no property value was provided, or the value is too long")
	errForbidden    = errors.New("the user does not have permission to edit the space with the given key")
	errUnauthorized = errors.New("the user does not have permission to create the space")
)

// NewSpace with the given options.
func NewSpace(options ...SpaceOption) *Space {
	v := &Space{}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithClient option.
func WithClient(client *Client) SpaceOption {
	return func(c *Space) {
		c.client = client
	}
}

// WithEndpoint option.
func WithEndpoint(location string) SpaceOption {
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
func WithCommandData(cd *CommandData) SpaceOption {
	return func(g *Space) {
		g.CommandData = cd
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
	s.CommandData = cmd

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
				Message: "Please enter the name of the space.",
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
				Message: "Please enter the description of the space.",
			},
		},
		{
			Name: "Private",
			Prompt: &survey.Confirm{
				Message: "Do you want to create a private space?",
			},
		},
	}

	return qs
}

// https://docs.atlassian.com/atlassian-confluence/REST/6.6.0/#space-createSpace
func (s *Space) create(reqBody *spaceRequest) (*SpaceResponse, error) {
	jsonbody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := s.endpoint.String() + "/space"

	if s.CommandData.Private {
		endpoint += "/_private"
	}

	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, errors.Wrap(err, "endpoint could not be parsed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logy.Debugf("create space request to %s", url.String())

	req, err := http.NewRequest("POST", url.String(), strings.NewReader(string(jsonbody)))
	if err != nil {
		return nil, errors.Wrap(err, "request could not be created")
	}

	req.Header.Add("Content-Type", "application/json")

	req = req.WithContext(ctx)

	resp, err := s.client.sendRequest(req)
	if err != nil {
		return nil, errors.Wrap(err, "request could not be executed")
	}

	logy.Debugf("create space request returned (%s)", resp.Status)

	// provide detail error messages
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return nil, errBadRequest
	case http.StatusForbidden:
		return nil, errForbidden
	case http.StatusUnauthorized:
		return nil, errUnauthorized
	case http.StatusServiceUnavailable:
		return nil, errors.Errorf("service is not available (%s)", resp.Status)
	case http.StatusInternalServerError:
		return nil, errors.Errorf("internal server error: %s", resp.Status)
	}

	var space SpaceResponse
	err = json.Unmarshal(resp.Payload, &space)
	if err != nil {
		return nil, err
	}

	return &space, nil
}

// Run the command
func (s *Space) Run() (*SpaceResponse, error) {
	return s.create(&spaceRequest{
		Key:  s.CommandData.Key,
		Name: s.CommandData.Name,
		Description: spaceDesc{
			Plain: spaceDescPlain{
				Representation: "plain",
				Value:          s.CommandData.Description,
			},
		},
	})
}

// buildSpaceKey create a space key from a string
func buildSpaceKey(spaceName string) string {
	return casee.ToChainCase(spaceName)
}

// spaceKeyValidator check if string is a valid space key
// https://confluence.atlassian.com/display/CONF58/Create+a+Space
func spaceKeyValidator(val interface{}) error {
	if str, ok := val.(string); ok {
		reg, err := regexp.Compile("([^a-zA-Z0-9]{1-255})")
		if err != nil {
			return err
		}
		if reg.MatchString(str) {
			return errors.New("invalid name")
		}
	}
	return nil
}
