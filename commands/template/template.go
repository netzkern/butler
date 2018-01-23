package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	logy "github.com/apex/log"
	"github.com/briandowns/spinner"
	"github.com/netzkern/butler/config"
	"github.com/pinzolo/casee"
	survey "gopkg.in/AlecAivazis/survey.v1"
	git "gopkg.in/src-d/go-git.v4"
)

const (
	startDelim = "butler{"
	endDelim   = "}"
)

var (
	allowedExtensions = []string{
		// Templates / Text
		".md",
		".txt",
		".html",
		".htm",
		".cshtml",
		".rtf",
		// Configuration
		".json",
		".yml",
		// Project
		".csproj",
		".sln",
		// Code
		".cs",
		".js",
	}
	excludedDirs = []string{
		"node_modules",
		"bower_components",
		"jspm_packages",
		"dist",
		"build",
		"log",
		"logs",
		"bin",
		"typings",

		"__tests__",
		"test",
		"tests",
		"powered-test",
		"images",
		"assets",
		"coverage",
	}
)

type (
	// ProjectData contains all project data
	ProjectData struct {
		Name        string
		Path        string
		Template    string
		Description string
	}
	// Templating command
	Templating struct {
		Templates         []config.Template
		Variables         map[string]string
		allowedExtensions map[string]struct{}
		excludedDirs      map[string]struct{}
		ch                chan func()
		wg                sync.WaitGroup
	}
)

// Option function.
type Option func(*Templating)

// New with the given options.
func New(options ...Option) *Templating {
	v := &Templating{
		allowedExtensions: toMap(allowedExtensions),
		excludedDirs:      toMap(excludedDirs),
		ch:                make(chan func()),
	}

	for _, o := range options {
		o(v)
	}

	return v
}

// WithVariables option.
func WithVariables(s map[string]string) Option {
	return func(v *Templating) {
		v.Variables = s
	}
}

// WithTemplates option.
func WithTemplates(s []config.Template) Option {
	return func(v *Templating) {
		v.Templates = s
	}
}

// cloneRepo clone a repo to the dst
func (t *Templating) cloneRepo(repoURL string, dest string) error {
	_, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL: repoURL,
	})

	if err == git.ErrRepositoryAlreadyExists {
		return fmt.Errorf("respository already exists. Remove '%s' directory", dest)
	}

	return err
}

// getTemplateByName returns the template by name
func (t *Templating) getTemplateByName(name string) *config.Template {
	for _, tpl := range t.Templates {
		if tpl.Name == name {
			return &tpl
		}
	}

	return nil
}

// getTemplateOptions return an array with all template names
func (t *Templating) getTemplateOptions() []string {
	tpls := []string{}

	for _, tpl := range t.Templates {
		tpls = append(tpls, tpl.Name)
	}

	sort.Strings(tpls)

	return tpls
}

// prompts activate interactive CLI
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

	project.Name = casee.ToPascalCase(project.Name)

	return project, nil
}

// Run the command
func (t *Templating) Run() error {
	project, err := t.prompts()

	if err != nil {
		return err
	}

	startTime := time.Now()

	tpl := t.getTemplateByName(project.Template)

	// clone repository
	if tpl != nil {
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Suffix = "Cloning repository..."
		s.FinalMSG = "Repository cloned!\n"
		s.Start()
		err := t.cloneRepo(tpl.Url, project.Path)
		s.Stop()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("template %s could not be found", project.Template)
	}

	// spinner progress
	spinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spinner.Suffix = "Processing templates..."
	spinner.FinalMSG = "Templates proceed!\n"
	spinner.Start()

	// start multiple routines
	t.startN(runtime.NumCPU())
	// close sync.WaitGroup and spinner when finished
	defer func() {
		t.stop()
		spinner.Stop()
		fmt.Printf("\nTotal: %s sec \n", strconv.FormatFloat(time.Since(startTime).Seconds(), 'f', 2, 64))
	}()

	walkErr := filepath.Walk(project.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		ctx := logy.WithFields(logy.Fields{
			"path": path,
			"size": info.Size(),
			"dir":  info.IsDir(),
		})

		// ignore hidden dirs and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// skip blacklisted directories
		if info.IsDir() {
			_, ok := t.excludedDirs[info.Name()]
			if ok {
				return filepath.SkipDir
			}
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

		// template file
		t.ch <- func() {
			ctx.Debugf("template")

			utilFuncMap := template.FuncMap{
				"toCamelCase":  casee.ToCamelCase,
				"toPascalCase": casee.ToPascalCase,
				"toSnakeCase":  casee.ToSnakeCase,
			}

			tmplPath, err := template.New(path).
				Delims(startDelim, endDelim).
				Funcs(utilFuncMap).
				Parse(path)

			if err != nil {
				ctx.WithError(err).Error("create template for filename")
				return
			}

			var pathBuffer bytes.Buffer
			err = tmplPath.Execute(&pathBuffer, templateData)
			if err != nil {
				ctx.WithError(err).Error("template for filename")
				return
			}

			newPath := pathBuffer.String()

			// check for valid file extension
			ext := filepath.Ext(newPath)
			_, ok := t.allowedExtensions[ext]
			if ok {
				dat, err := ioutil.ReadFile(path)
				tmpl, err := template.New(newPath).
					Delims(startDelim, endDelim).
					Funcs(utilFuncMap).
					Parse(string(dat))

				f, err := os.Create(newPath)

				defer f.Close()

				if err != nil {
					ctx.WithError(err).Error("read file")
					return
				}

				err = tmpl.Execute(f, templateData)

				if err != nil {
					ctx.WithError(err).Error("template file")
					return
				}

				if path != newPath {
					os.Remove(path)
				}
			} else {
				os.Rename(path, newPath)
			}
		}

		return nil
	})

	if walkErr != nil {
		return walkErr
	}

	return nil
}

// startN starts n loops.
func (t *Templating) startN(n int) {
	for i := 0; i < n; i++ {
		t.wg.Add(1)
		go t.start()
	}
}

// start loop.
func (t *Templating) start() {
	defer t.wg.Done()
	for fn := range t.ch {
		fn()
	}
}

// stop loop.
func (t *Templating) stop() {
	close(t.ch)
	t.wg.Wait()
}

// toMap returns a map from slice.
func toMap(s []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
