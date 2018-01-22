package main

import (
	"fmt"
	"log"
	"os"

	logy "github.com/apex/log"
	"github.com/blang/semver"
	"github.com/netzkern/butler/commands/template"
	"github.com/netzkern/butler/config"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	debug = false
	trace = false
)

var (
	cfg     *config.Config
	version = "0.0.9"
	qs      = []*survey.Question{
		{
			Name:     "action",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "How can I help you, Sir?",
				Options: []string{"Project Templates", "Auto Update", "Version"},
			},
		},
	}
)

func init() {
	// go-update logger
	logy.SetLevel(logy.InfoLevel)
	cfg = config.ParseConfig()
}

func main() {
	fmt.Println("Welcome to Butler🤵, your personal assistent to scaffolding your projects.\n")

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

func confirmAndSelfUpdate() {
	selfupdate.EnableLog()

	latest, found, err := selfupdate.DetectLatest("netzkern/butler")
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return
	}

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
