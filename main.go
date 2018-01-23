package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	logy "github.com/apex/log"
	"github.com/blang/semver"
	"github.com/netzkern/butler/commands/template"
	"github.com/netzkern/butler/config"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/core"
)

const (
	debug   = false
	trace   = false
	appName = "Butler"
	appDesc = "Welcome to Butler🤵, your personal assistent to scaffolding your projects.\n"
)

var (
	cfg      *config.Config
	version  = "master"
	commands = []string{
		"Project Templates",
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
	// go-update logger
	logy.SetLevel(logy.InfoLevel)
	cfg = config.ParseConfig()

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

func main() {
	if len(os.Args[1:]) > 0 {
		cliMode()
		return
	}

	interactiveCliMode()
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
		)
		err := command.Run()
		if err != nil {
			logy.Errorf(err.Error())
		}
	case "Auto Update":
		confirmAndSelfUpdate()
	case "Version":
		fmt.Printf("Version: %s\n", version)
	default:
		logy.Infof("Command %s not implemented!", taskType)
	}
}

func cliMode() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "your personal assistent to scaffolding your projects."
	app.Author = "netzkern AG"
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

func confirmAndSelfUpdate() {
	latest, found, err := selfupdate.DetectLatest("netzkern/butler")
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return
	}

	fmt.Printf("Version: %s\n", version)
	v := semver.MustParse(version)
	if !found || latest.Version.Equals(v) {
		log.Println("Current version is the latest")
		return
	}

	update := false
	prompt := &survey.Confirm{
		Message: "Do you want to update to " + latest.Version.String() + "?",
	}
	survey.AskOne(prompt, &update, nil)

	if !update {
		return
	}

	cmdPath, err := os.Executable()
	if err != nil {
		logy.WithError(err).Error("os executable")
		return
	}

	if err := selfupdate.UpdateTo(latest.AssetURL, cmdPath); err != nil {
		log.Println("Error occurred while updating binary:", err)
		return
	}
	log.Println("Successfully updated to version", latest.Version)
}
