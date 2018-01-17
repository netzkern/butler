package main

import (
	"fmt"

	"github.com/netzkern/butler/commands"
	"github.com/netzkern/butler/config"
	"gopkg.in/AlecAivazis/survey.v1"
)

var (
	version = "master"
	qs      = []*survey.Question{
		{
			Name:     "action",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "How can I help you, Sir?",
				Options: []string{"Templating", "Jira", "Tfs"},
			},
		},
	}
)

func main() {
	fmt.Println("  ____        _   _           ")
	fmt.Println(" |  _ \\      | | | |          ")
	fmt.Println(" | |_) |_   _| |_| | ___ _ __ ")
	fmt.Println(" |  _ <| | | | __| |/ _ \\ '__|")
	fmt.Println(" | |_) | |_| | |_| |  __/ |   ")
	fmt.Println(" |____/ \\__,_|\\__|_|\\___|_|   ")
	fmt.Println("                              ")
	fmt.Println("Welcome to Butler, your personal assistent to scaffolding your projects.")
	fmt.Println("Version: ", version)

	cfg := config.ParseConfig()

	answers := struct {
		Action string
	}{}

	err := survey.Ask(qs, &answers)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	switch taskType := answers.Action; taskType {
	case "Templating":
		command := commands.Templating{Templates: cfg.Templates}
		err := command.Run()
		if err != nil {
			fmt.Println("butler: " + err.Error())
		}
	case "Jira":
		fmt.Println("butler: Jira command not implemented!")
	case "Tfs":
		fmt.Println("butler: Tfs command not implemented!")
	default:
		fmt.Printf("butler: Command %s not implemented!", taskType)
	}
}
