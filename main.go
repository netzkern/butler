package main

import (
	"fmt"

	"github.com/netzkern/butler/commands"
	"gopkg.in/AlecAivazis/survey.v1"
)

var version = "master"

var qs = []*survey.Question{
	{
		Name:     "action",
		Validate: survey.Required,
		Prompt: &survey.Select{
			Message: "How can I help you, Sir?",
			Options: []string{"Templating", "Jira", "Tfs"},
		},
	},
}

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
		command := commands.Templating{}
		err := command.Run()
		if err != nil {
			fmt.Println(err.Error())
		}
	case "Jira":
		fmt.Println("Jira command not implemented!")
	case "Tfs":
		fmt.Println("Tfs command not implemented!")
	default:
		fmt.Printf("Command %s not implemented!", taskType)
	}
}
