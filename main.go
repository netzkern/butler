package main

import (
	"fmt"
	"runtime"

	logy "github.com/apex/log"
	"github.com/netzkern/butler/commands"
	"github.com/netzkern/butler/config"
	update "github.com/tj/go-update"
	"github.com/tj/go-update/progress"
	"github.com/tj/go-update/stores/github"
	"gopkg.in/AlecAivazis/survey.v1"
)

const (
	debug = false
	trace = false
)

var (
	cfg     *config.Config
	version = "master"
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
		command := commands.Templating{
			Templates: cfg.Templates,
			Variables: cfg.Variables,
		}
		err := command.Run()
		if err != nil {
			logy.Errorf(err.Error())
		}
	case "Auto Update":
		updateApp()
	case "Version":
		fmt.Printf("Version: %s\n", version)
	default:
		logy.Infof("Command %s not implemented!", taskType)
	}
}

func updateApp() {
	command := "butler"

	if runtime.GOOS == "windows" {
		command = "butler.exe"
	}

	m := &update.Manager{
		Command: command,
		Store: &github.Store{
			Owner: "netzkern",
			Repo:  "butler",
		},
	}

	// fetch the new releases
	releases, err := m.LatestReleases()
	if err != nil {
		logy.Fatalf("error fetching releases: %s", err)
	}

	// no updates
	if len(releases) == 0 {
		logy.Infof("no updates")
		return
	}

	// latest release
	latest := releases[0]

	// find the tarball for this system
	a := latest.FindTarball(runtime.GOOS, runtime.GOARCH)
	if a == nil {
		logy.Infof("no binary for your system")
		return
	}

	// whitespace
	fmt.Println()

	// download tarball to a tmp dir
	tarball, err := a.DownloadProxy(progress.Reader)
	if err != nil {
		logy.Fatalf("error downloading: %s", err)
	}

	// install it
	if err := m.Install(tarball); err != nil {
		logy.Fatalf("error installing: %s", err)
	}

	fmt.Printf("Updated to %s\n", latest.Version)
}
