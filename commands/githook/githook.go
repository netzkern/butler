package githook

import (
	"os"
	"path"

	logy "github.com/apex/log"
	"github.com/netzkern/butler/utils"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

var (
	// Hooks a list of all available git hooks
	Hooks = []string{
		"applypatch-msg",
		"commit-msg",
		"post-commit",
		"post-receive",
		"post-update",
		"pre-applypatch",
		"pre-commit",
		"prepare-commit-msg",
		"pre-rebase",
		"update",
	}
	// convention
	repoHookDir = "git_hooks"
)

// CommandData contains all command related data
type CommandData struct {
	Path  string
	Hooks []string
}

// Githook command to create git hooks
type Githook struct {
	Path        string
	Cwd         string
	CommandData *CommandData
}

// Option function.
type Option func(*Githook)

// New with the given options.
func New(options ...Option) *Githook {
	v := &Githook{}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithCwd option.
func WithCwd(dir string) Option {
	return func(g *Githook) {
		g.Cwd = dir
	}
}

// WithCommandData option.
func WithCommandData(cd *CommandData) Option {
	return func(g *Githook) {
		g.CommandData = cd
	}
}

// Install will create hard links from local git_hooks to the corresponding git hooks
func (g *Githook) install() error {
	for _, h := range g.CommandData.Hooks {
		hookGitPath := path.Join(g.CommandData.Path, ".git", "hooks", h)
		hookRepoPath := path.Join(g.CommandData.Path, repoHookDir, h)

		if !utils.Exists(hookGitPath) {
			dir := path.Dir(hookGitPath)
			logy.Debugf("path '%s' created", dir)
			// when git wasn't initialized with a hook folder
			err := utils.CreateDirIfNotExist(dir)
			if err != nil {
				logy.WithError(err).Error("could not create directory")
			}
		} else {
			// remove existing hooks
			// should be no problem because all hooks are versioned
			os.Remove(hookGitPath)
		}

		if utils.Exists(hookRepoPath) {
			logy.Debugf("create symlink old: '%s', new: '%s'", hookGitPath, hookRepoPath)
			err := os.Link(hookRepoPath, hookGitPath)
			if err != nil {
				logy.WithError(err).Errorf("could not link hook '%s'", h)
			} else {
				logy.Infof("hook '%s' installed", h)
			}
		} else {
			logy.Debugf("template for hook '%s' could not be found in '%s'", h, path.Join(g.CommandData.Path, repoHookDir))
		}

	}
	return nil
}

// getQuestions return all required prompts
func (g *Githook) getQuestions() []*survey.Question {
	qs := []*survey.Question{
		{
			Name:     "Path",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "What's the root directory of your git repository?",
				Default: g.Cwd,
			},
		},
		{
			Name:     "Hooks",
			Validate: survey.Required,
			Prompt: &survey.MultiSelect{
				Message: "Please select your hooks",
				Options: Hooks,
			},
		},
	}

	return qs
}

// StartCommandSurvey collect all required informations from user
func (g *Githook) StartCommandSurvey() error {
	var cmd = &CommandData{}

	// start command prompts
	err := survey.Ask(g.getQuestions(), cmd)
	if err != nil {
		return err
	}

	g.CommandData = cmd

	return nil
}

// Run the command
func (g *Githook) Run() error {
	return g.install()
}
