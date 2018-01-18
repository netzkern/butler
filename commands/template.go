package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/briandowns/spinner"
	"github.com/netzkern/butler/config"
	log "github.com/netzkern/butler/logger"
	"github.com/netzkern/butler/util"
	survey "gopkg.in/AlecAivazis/survey.v1"
	git "gopkg.in/src-d/go-git.v4"
)

const (
	startDelim = "butler{"
	endDelim   = "}"
)

var (
	allowedExtensions = [...]string{".md", ".txt", ".html", ".htm", ".rtf", ".json", ".yml", ".csproj", ".sln"}
	blacklistDirs     = map[string]bool{
		"node_modules":     true,
		"bower_components": true,
		"jspm_packages":    true,
		"dist":             true,
		"logs":             true,
		"bin":              true,
	}
)

type (
	ProjectData struct {
		Name        string
		Path        string
		Template    string
		Description string
	}
	Templating struct {
		Templates []config.Template
		Variables map[string]string
		Logger    *log.Logger
	}
)

func (t *Templating) cloneRepo(repoURL string, dest string) error {
	_, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:      repoURL,
		Progress: os.Stdout,
	})

	if err == git.ErrRepositoryAlreadyExists {
		return fmt.Errorf("respository already exists. Remove %s", dest)
	}

	return err
}

func (t *Templating) getTemplateByName(name string) *config.Template {
	for _, tpl := range t.Templates {
		if tpl.Name == name {
			return &tpl
		}
	}

	return nil
}

func (t *Templating) getTemplateOptions() []string {
	tpls := []string{}

	for _, tpl := range t.Templates {
		tpls = append(tpls, tpl.Name)
	}

	sort.Strings(tpls)

	return tpls
}

func (t *Templating) prompts() (*ProjectData, error) {
	var simpleQs = []*survey.Question{
		{
			Name:     "Template",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "What system are you using?",
				Options: t.getTemplateOptions(),
				Help:    "You can add additional templates in your config",
			},
		},
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "What is the project name?",
				Help:    "Allowed character 0-9, A-Z, _-",
			},
			Validate: survey.Required,
		},
		{
			Name: "Description",
			Prompt: &survey.Input{
				Message: "What is the project description?",
			},
		},
		{
			Name:     "Path",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "What is the destination?",
				Default: "src",
				Help:    "The place of your new project",
			},
		},
	}

	var project = &ProjectData{}

	// ask the question
	err := survey.Ask(simpleQs, project)

	if err != nil {
		return nil, err
	}

	project.Name = util.NormalizeProjectName(project.Name)

	return project, nil
}

// Run - Runs the command
func (t *Templating) Run() error {
	project, err := t.prompts()

	if err != nil {
		return err
	}

	tpl := t.getTemplateByName(project.Template)

	if tpl != nil {
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Suffix = " Cloning repository..."
		s.FinalMSG = "Complete!\n"
		s.Start()
		err := t.cloneRepo(tpl.Url, project.Path)
		s.Stop()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("template %s could not be found", project.Template)
	}

	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " Processing templates..."
	s.FinalMSG = "Complete!\n"
	s.Start()

	walkErr := filepath.Walk(project.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip blacklisted directories
		if info.IsDir() && blacklistDirs[info.Name()] {
			return filepath.SkipDir
		}

		// ignore hidden dirs
		if strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// skip directories but go ahead with traversing
		if info.IsDir() {
			return nil
		}

		// check for valid file extension
		fileExt := strings.ToLower(info.Name())
		validExt := false
		for _, ext := range allowedExtensions {
			if strings.HasSuffix(fileExt, ext) {
				validExt = true
				break
			}
		}

		if !validExt {
			return nil
		}

		var templateData = struct {
			Project *ProjectData
			Date    string
			Year    int
			Vars    map[string]string
		}{
			project,
			time.Now().Format(time.RFC3339),
			time.Now().Year(),
			t.Variables,
		}

		dat, err := ioutil.ReadFile(path)

		var b bytes.Buffer

		tmpl, err := template.New(path).Delims(startDelim, endDelim).Parse(string(dat))
		tmplPath, err := template.New(path).Delims(startDelim, endDelim).Parse(path)

		err = tmplPath.Execute(&b, templateData)

		f, err := os.Create(path)

		defer func() {
			err = os.Rename(path, b.String())
		}()

		defer f.Close()

		defer func() {
			if r := recover(); r != nil {
				t.Logger.Tracef("file %s was recovered due to template error! Error: %s \n", path, r)
				ioutil.WriteFile(path, dat, 0644)
			}
		}()

		if err != nil {
			return err
		}

		err = tmpl.Execute(f, templateData)

		if err != nil {
			return err
		}

		return nil
	})

	s.Stop()

	if walkErr != nil {
		return walkErr
	}

	return nil
}
