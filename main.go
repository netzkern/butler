package main

import (
	"fmt"
	"runtime"

	apexLog "github.com/apex/log"
	"github.com/netzkern/butler/commands"
	"github.com/netzkern/butler/config"
	"github.com/netzkern/butler/logger"
	update "github.com/tj/go-update"
	"github.com/tj/go-update/progress"
	"github.com/tj/go-update/stores/github"
	"gopkg.in/AlecAivazis/survey.v1"
)

var (
	log     *logger.Logger
	cfg     *config.Config
	version = "master"
	qs      = []*survey.Question{
		{
			Name:     "action",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "How can I help you, Sir?",
				Options: []string{"Project Templates", "Auto Update"},
			},
		},
	}
)

func init() {
	// go-update logger
	apexLog.SetLevel(apexLog.ErrorLevel)

	cfg = config.ParseConfig()
	if cfg.Logger == "file" {
		log = logger.NewFileLogger("butler.log", true, false, false, true)
	} else {
		log = logger.NewStdLogger(true, true, false, false, true)
	}
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
		log.Errorf(err.Error())
		return
	}

	switch taskType := answers.Action; taskType {
	case "Project Templates":
		command := commands.Templating{
			Templates: cfg.Templates,
			Variables: cfg.Variables,
			Logger:    log,
		}
		err := command.Run()
		if err != nil {
			log.Errorf(err.Error())
		}
	case "Auto Update":
		updateApp()
	default:
		log.Noticef("Command %s not implemented!", taskType)
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
		log.Fatalf("error fetching releases: %s", err)
	}

	// no updates
	if len(releases) == 0 {
		log.Noticef("no updates")
		return
	}

	// latest release
	latest := releases[0]

	// find the tarball for this system
	a := latest.FindTarball(runtime.GOOS, runtime.GOARCH)
	if a == nil {
		log.Noticef("no binary for your system")
		return
	}

	// whitespace
	fmt.Println()

	// download tarball to a tmp dir
	tarball, err := a.DownloadProxy(progress.Reader)
	if err != nil {
		log.Fatalf("error downloading: %s", err)
	}

	// install it
	if err := m.Install(tarball); err != nil {
		log.Fatalf("error installing: %s", err)
	}

	fmt.Printf("Updated to %s\n", latest.Version)
}
