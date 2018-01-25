package main

import (
	"fmt"
	"os"
	"runtime"

	logy "github.com/apex/log"
	"github.com/netzkern/butler/commands/githook"
	"github.com/netzkern/butler/commands/template"
	"github.com/netzkern/butler/config"
	"github.com/netzkern/butler/updater"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/core"
)

const (
	debug          = false
	trace          = false
	appName        = "Butler"
	appDesc        = "Welcome to Butler, your personal assistent to scaffold new projects.\n"
	author         = "netzkern AG"
	repository     = "netzkern/butler"
	surveyFilename = "butler-survey.yml"
	configName     = "butler.yml"
)

var (
	cfg      *config.Config
	version  = "master"
	commands = []string{
		"Project Templates",
		"Install Git Hooks",
		"Auto Update",
		"Version",
	}
	qs = []*survey.Question{
		{
			Name:     "action",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "How can I help you, Sir?",
				Options: commands,
			},
		},
	}
)

func init() {
	logy.SetLevel(logy.ErrorLevel)
	cfg = config.ParseConfig(configName)

	// Windows comaptible symbols
	if runtime.GOOS == "windows" {
		core.ErrorIcon = "X"
		core.HelpIcon = "????"
		core.QuestionIcon = "?"
		core.SelectFocusIcon = ">"
		core.MarkedOptionIcon = "[x]"
		core.UnmarkedOptionIcon = "[ ]"
	}
}

func interactiveCliMode() {
	fmt.Println(appDesc)

	answers := struct {
		Action string
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		logy.Errorf(err.Error())
		return
	}

	switch taskType := answers.Action; taskType {
	case "Project Templates":
		command := template.New(
			template.WithTemplates(cfg.Templates),
			template.WithVariables(cfg.Variables),
			template.SetConfigName(surveyFilename),
		)
		command.StartCommandSurvey()
		err := command.Run()
		if err != nil {
			logy.Errorf(err.Error())
		}
	case "Install Git Hooks":
		command := githook.New()
		command.StartCommandSurvey()
		err := command.Run()
		if err != nil {
			logy.Errorf(err.Error())
		}
	case "Auto Update":
		updater.ConfirmAndSelfUpdate(repository, version)
	case "Version":
		fmt.Printf("Version: %s\n", version)
	default:
		logy.Infof("Command %s not implemented!", taskType)
	}
}

// WIP
func cliMode() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "your personal assistent to scaffold new projects."
	app.Author = author
	app.Version = version
	app.Description = appDesc
	app.Commands = []cli.Command{
		{
			Name:    "template",
			Aliases: []string{"t"},
			Usage:   "options for task templates",
			Subcommands: []cli.Command{
				{
					Name:  "create",
					Usage: "checkout a new template",
					Action: func(c *cli.Context) error {
						fmt.Println("new task template: ", c.Args().First())
						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func main() {
	if len(os.Args[1:]) > 0 {
		cliMode()
		return
	}

	interactiveCliMode()
}
