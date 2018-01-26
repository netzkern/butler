package main

import (
	"fmt"
	"os"
	"runtime"
	"sort"

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
	logy.SetLevel(logy.DebugLevel)
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

func cliMode() {

	type surveyResult map[string]interface{}

	app := cli.NewApp()
	app.Name = appName
	app.Usage = "your personal assistent to scaffold new projects."
	app.Author = author
	app.Version = version
	app.Description = appDesc

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "logLevel",
			Value:  "info",
			Usage:  "Log level",
			EnvVar: "BUTLER_LOG_LEVEL",
		},
	}
	app.Commands = []cli.Command{
		{
			Name:    "interactive",
			Aliases: []string{"ui"},
			Usage:   "Enable interactive cli",
			Action: func(c *cli.Context) error {
				setLogLevel(c.GlobalString("logLevel"))
				interactiveCliMode()
				return nil
			},
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	app.Run(os.Args)
}

func main() {
	if len(os.Args[1:]) > 0 {
		cliMode()
		return
	}

	interactiveCliMode()
}

func setLogLevel(level string) {
	switch level {
	case "info":
		logy.SetLevel(logy.InfoLevel)
	case "debug":
		logy.SetLevel(logy.DebugLevel)
	case "fatal":
		logy.SetLevel(logy.FatalLevel)
	case "error":
		logy.SetLevel(logy.ErrorLevel)
	case "warn":
		logy.SetLevel(logy.WarnLevel)
	default:
		logy.Fatalf("Invalid log level '%s'", level)
	}
}
