package githook

import (
	"os"
	"path"

	logy "github.com/apex/log"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

var (
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

// WithCommandData option.
func WithCommandData(cd *CommandData) Option {
	return func(g *Githook) {
		g.CommandData = cd
	}
}

// Install will create hard links from local git_hooks to the corresponding git hooks
func (g *Githook) Install() error {
	for _, h := range g.CommandData.Hooks {
		hookGitPath := path.Join(g.CommandData.Path, ".git", "hooks", h)
		hookRepoPath := path.Join(g.CommandData.Path, repoHookDir, h)
		if !exists(hookGitPath) {
			dir := path.Dir(hookGitPath)
			logy.Debugf("Path '%s' created", dir)
			err := createDirIfNotExist(dir)
			if err != nil {
				logy.WithError(err).Error("Could not create directory")
			}
		} else {
			os.Remove(hookGitPath)
		}
		if exists(hookRepoPath) {
			logy.Debugf("Create symlink old: %s, new: %s", hookGitPath, hookRepoPath)
			err := os.Link(hookRepoPath, hookGitPath)
			if err != nil {
				logy.WithError(err).Errorf("Could not link hook %s", h)
				return err
			}
			logy.Debugf("hook '%s' installed", h)
		} else {
			logy.Infof("template for hook '%s' could not be found in %s", h, path.Join(g.CommandData.Path, repoHookDir))
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
				Message: "What is the destination?",
				Default: ".",
			},
		},
		{
			Name:     "Hooks",
			Validate: survey.Required,
			Prompt: &survey.MultiSelect{
				Message: "Which hooks should be installed?",
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
	return g.Install()
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func createDirIfNotExist(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
