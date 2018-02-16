package builder

import (
	"net/url"
	"sort"

	"github.com/netzkern/butler/commands/confluence"
	"github.com/netzkern/butler/commands/confluence/page"
	"github.com/netzkern/butler/config"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

type (
	// Option function.
	Option func(*TreeBuilder)
	// TreeBuilder command to create git hooks
	TreeBuilder struct {
		client      *confluence.Client
		endpoint    *url.URL
		spaceKey    string
		templates   []config.ConfluenceTemplate
		commandData *commandData
	}
	// commandData contains all command related data
	commandData struct {
		SpaceKey string
		Template string
		Tree     []config.ConfluencePage
	}
)

// NewTreeBuilder with the given options.
func NewTreeBuilder(options ...Option) *TreeBuilder {
	v := &TreeBuilder{}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithClient option.
func WithClient(client *confluence.Client) Option {
	return func(c *TreeBuilder) {
		c.client = client
	}
}

// WithEndpoint option.
func WithEndpoint(location string) Option {
	return func(c *TreeBuilder) {
		u, err := url.ParseRequestURI(location)
		if err != nil {
			panic(err)
		}
		c.endpoint = u
	}
}

// WithTemplates option.
func WithTemplates(s []config.ConfluenceTemplate) Option {
	return func(t *TreeBuilder) {
		t.templates = s
	}
}

// WithSpaceKey option.
func WithSpaceKey(key string) Option {
	return func(g *TreeBuilder) {
		g.spaceKey = key
	}
}

// getQuestions return all required prompts
func (s *TreeBuilder) getQuestions() []*survey.Question {
	qs := []*survey.Question{
		{
			Name:     "Template",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "Please select a template",
				Options: s.getTemplateOptions(),
				Help:    "You can add additional templates in your config",
			},
		},
	}

	return qs
}

// StartCommandSurvey collect all required informations from user
func (s *TreeBuilder) StartCommandSurvey() error {
	var cmd = &commandData{}

	// start command prompts
	err := survey.Ask(s.getQuestions(), cmd)
	if err != nil {
		return err
	}

	for _, tpl := range s.templates {
		if tpl.Name == cmd.Template {
			cmd.Tree = tpl.Pages
			break
		}
	}

	s.commandData = cmd

	return nil
}

// getTemplateOptions return an array with all template names
func (s *TreeBuilder) getTemplateOptions() []string {
	tpls := []string{}

	for _, tpl := range s.templates {
		tpls = append(tpls, tpl.Name)
	}

	sort.Strings(tpls)

	return tpls
}

func (s *TreeBuilder) createTree(ancestorID string, children []config.ConfluencePage) error {
	if children == nil || len(children) == 0 {
		return nil
	}

	for _, p := range children {
		page := page.NewPage(
			page.WithClient(s.client),
			page.WithEndpoint(s.endpoint.String()),
			page.WithCommandData(
				&page.CommandData{
					AncestorID: ancestorID,
					SpaceKey:   s.spaceKey,
					Title:      p.Name,
					Type:       "page",
				},
			),
		)
		pageResp, err := page.Run()
		if err != nil {
			return err
		}
		s.createTree(pageResp.ID, p.Children)
	}
	return nil
}

// Run the command
func (s *TreeBuilder) Run() error {
	return s.createTree("", s.commandData.Tree)
}
