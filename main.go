package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"sort"

	"github.com/skratchdot/open-golang/open"

	logy "github.com/apex/log"
	"github.com/netzkern/butler/commands/confluence"
	"github.com/netzkern/butler/commands/confluence/builder"
	"github.com/netzkern/butler/commands/confluence/space"
	"github.com/netzkern/butler/commands/githook"
	"github.com/netzkern/butler/commands/template"
	"github.com/netzkern/butler/config"
	"github.com/netzkern/butler/updater"
	"github.com/urfave/cli"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/AlecAivazis/survey.v1/core"
	yaml "gopkg.in/yaml.v2"
)

const (
	appName         = "Butler"
	appDesc         = "Welcome to Butler, your personal assistent to scaffold new projects.\n"
	githubIssueLink = "https://github.com/netzkern/butler/issues/new"
	author          = "netzkern AG"
	repository      = "netzkern/butler"
	surveyFilename  = "butler-survey.yml"
	configName      = "butler.yml"
)

type commandSelection struct {
	Action string
}

var (
	cfg             *config.Config
	version         = "0.9.0"
	primaryCommands = []string{
		"Create Project",
		"Create Confluence Space",
		"Create Git Hooks",
		"Maintenance",
		"Exit",
	}
	devCommands = []string{
		"Dump config",
		"Auto Update",
		"Report a bug",
		"Version",
	}
	primaryQs = []*survey.Question{
		{
			Name:     "action",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "How can I help you, Sir?",
				Options: primaryCommands,
			},
		},
	}
	devQs = []*survey.Question{
		{
			Name:     "action",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "How can I help you, Sir?",
				Options: devCommands,
			},
		},
	}
)

func init() {
	logy.SetLevel(logy.InfoLevel)
	cfg = config.ParseConfig(configName)

	// TERM contains a identifier for the text window’s capabilities (UNIX).
	if os.Getenv("TERM") == "xterm-256color" {
		return
	}

	// Fallback for Windows
	if runtime.GOOS == "windows" {
		core.ErrorIcon = "x"
		core.HelpIcon = "????"
		core.QuestionIcon = "?"
		core.SelectFocusIcon = ">"
		core.MarkedOptionIcon = "[x]"
		core.UnmarkedOptionIcon = "[ ]"
	}
}

func showPrimaryCommands() *commandSelection {
	answer := &commandSelection{}

	err := survey.Ask(primaryQs, answer)
	if err != nil {
		logy.Fatal(err.Error())
	}

	return answer
}

func showMaintananceCommands() *commandSelection {
	answer := &commandSelection{}

	err := survey.Ask(devQs, answer)
	if err != nil {
		logy.Fatal(err.Error())
	}

	return answer
}

func listMaintananceCommands() {
	answer := showMaintananceCommands()

	switch taskType := answer.Action; taskType {
	case "Dump config":
		dumpConfig()
	case "Auto Update":
		updater.ConfirmAndSelfUpdate(repository, version)
	case "Report a bug":
		err := open.Start(githubIssueLink)
		if err != nil {
			logy.WithError(err).Error("report a bug")
			return
		}
	case "Version":
		fmt.Printf("Version: %s\n", version)
	default:
		logy.Infof("Command %s is not implemented!", taskType)
	}
}

func interactiveCliMode() {
	fmt.Println(appDesc)

	answer := showPrimaryCommands()

	cd, err := os.Getwd()
	if err != nil {
		logy.WithError(err).Error("getwd")
		return
	}

	switch taskType := answer.Action; taskType {
	case "Create Project":
		command := template.New(
			template.WithTemplates(cfg.Templates),
			template.WithVariables(cfg.Variables),
			template.SetConfigName(surveyFilename),
			template.WithButlerVersion(version),
			template.WithCwd(cd),
		)

		err := command.StartCommandSurvey()
		if err != nil {
			logy.WithError(err).Error("start survey")
			return
		}

		err = command.Run()
		if err != nil {
			logy.WithError(err).Error("run command")
			return
		}

		fmt.Println()
		command.TaskTracker.PrintSummary(os.Stdout)
	case "Create Confluence Space":
		// validate confluence configuration
		if cfg.ConfluenceAuthMethod != "" {
			if cfg.ConfluenceAuthMethod == "basic" {
				if len(cfg.ConfluenceBasicAuth) != 2 {
					logy.WithField("ENV", "CONFLUENCE_BASIC_AUTH").
						Fatalf("invalid basic auth credentials")
				}
			} else {
				logy.WithField("ENV", "CONFLUENCE_AUTH_METHOD").
					Fatalf("only basic auth is currently supported")
			}
		} else {
			logy.Fatalf(
				"invalid confluence settings. For more information see https://github.com/netzkern/butler/tree/master/docs/confluence.md",
			)
		}

		// validate confluence url
		if cfg.ConfluenceURL != "" {
			_, err := url.ParseRequestURI(cfg.ConfluenceURL)
			if err != nil {
				logy.WithField("ENV", "BUTLER_CONFLUENCE_URL").
					Fatalf("invalid url")
			}
		}

		client := confluence.NewClient(
			confluence.WithAuth(
				confluence.BasicAuth(
					cfg.ConfluenceBasicAuth[0],
					cfg.ConfluenceBasicAuth[1],
				),
			),
		)

		createSpaceCmd := space.NewSpace(
			space.WithEndpoint(cfg.ConfluenceURL),
			space.WithClient(client),
		)

		err := createSpaceCmd.StartCommandSurvey()
		if err != nil {
			logy.WithError(err).Error("start survey")
			return
		}

		spaceData, err := createSpaceCmd.Run()
		if err != nil {
			logy.WithError(err).Error("run command")
			return
		}

		// skip tree builder when no template exist
		if len(cfg.Confluence.Templates) > 0 {
			treeBuilderCmd := builder.NewTreeBuilder(
				builder.WithTemplates(cfg.Confluence.Templates),
				builder.WithClient(client),
				builder.WithEndpoint(cfg.ConfluenceURL),
				builder.WithSpaceKey(spaceData.Key),
			)

			err = treeBuilderCmd.StartCommandSurvey()
			if err != nil {
				logy.WithError(err).Error("start survey")
				return
			}

			if err := treeBuilderCmd.Run(); err != nil {
				logy.WithError(err).Error("run command")
				return
			}
		}
	case "Create Git Hooks":
		command := githook.New(githook.WithCwd(cd))

		err := command.StartCommandSurvey()
		if err != nil {
			logy.WithError(err).Error("start survey")
			return
		}

		err = command.Run()
		if err != nil {
			logy.WithError(err).Error("run command")
			return
		}
	case "Maintenance":
		listMaintananceCommands()
	case "Exit":
		os.Exit(0)
	default:
		logy.Infof("Command '%s' is not implemented!", taskType)
		return
	}

	fmt.Println("Command executed successfully")
}

func dumpConfig() error {
	str, err := yaml.Marshal(cfg)

	if err != nil {
		return err
	}

	fmt.Println(string(str))

	return nil
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
		{
			Name:   "dump-config",
			Usage:  "Dumps the final config file",
			Action: func(c *cli.Context) error { return dumpConfig() },
		},
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	err := app.Run(os.Args)

	if err != nil {
		logy.Fatalf("Failed executing command, see %s", err)
	}
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

func main() {
	if len(os.Args[1:]) > 0 {
		cliMode()
		return
	}

	interactiveCliMode()
}
