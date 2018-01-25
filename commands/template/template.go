package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
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
	"github.com/netzkern/butler/commands/githook"
	"github.com/netzkern/butler/config"
	"github.com/pinzolo/casee"
	uuid "github.com/satori/go.uuid"
	survey "gopkg.in/AlecAivazis/survey.v1"
	git "gopkg.in/src-d/go-git.v4"
)

const (
	startContentDelim = "butler{"
	endContentDelim   = "}"

	startNameDelim = "{"
	endNameDelim   = "}"

	envPrefix = "BUTLER"
)

type (
	// CommandData contains all project data
	CommandData struct {
		Name        string
		Path        string
		Template    string
		Description string
	}
	// Templating command
	Templating struct {
		Templates       []config.Template
		Variables       map[string]string
		configName      string
		excludedDirs    map[string]struct{}
		excludedExts    map[string]struct{}
		ch              chan func()
		wg              sync.WaitGroup
		surveyResult    map[string]interface{}
		CommandData     *CommandData
		TemplateData    *TemplateData
		templateFuncMap template.FuncMap
		surveys         *Survey
	}
	// TemplateData basic template data
	TemplateData struct {
		Project *CommandData
		Date    string
		Year    int
		Vars    map[string]string
	}
)

// Option function.
type Option func(*Templating)

// New with the given options.
func New(options ...Option) *Templating {
	v := &Templating{
		excludedDirs: toMap(ExcludedDirs),
		excludedExts: toMap(BinaryFileExt),
		ch:           make(chan func(), runtime.NumCPU()),
	}

	v.templateFuncMap = template.FuncMap{
		"toCamelCase":  casee.ToCamelCase,
		"toPascalCase": casee.ToPascalCase,
		"toSnakeCase":  casee.ToSnakeCase,
		"join":         strings.Join,
		"uuid": func() (string, error) {
			ui, err := uuid.NewV4()
			if err != nil {
				return "", err
			}
			return ui.String(), nil
		},
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

// SetConfigName option.
func SetConfigName(s string) Option {
	return func(v *Templating) {
		v.configName = s
	}
}

// WithTemplates option.
func WithTemplates(s []config.Template) Option {
	return func(v *Templating) {
		v.Templates = s
	}
}

// WithCommandData option.
func WithCommandData(cd *CommandData) Option {
	return func(t *Templating) {
		t.CommandData = cd
	}
}

// WithTemplateSurveyResults option.
func WithTemplateSurveyResults(sr map[string]interface{}) Option {
	return func(t *Templating) {
		t.surveyResult = sr
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

	os.RemoveAll(filepath.Join(dest, ".git"))

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

// getQuestions return all required prompts
func (t *Templating) getQuestions() []*survey.Question {
	qs := []*survey.Question{
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

	return qs
}

// Skip returns an error when a directory should be skipped or true with a file
func (t *Templating) Skip(path string, info os.FileInfo) (bool, error) {
	// ignore hidden dirs and files
	if strings.HasPrefix(info.Name(), ".") {
		if info.IsDir() {
			return false, filepath.SkipDir
		}
		return true, nil
	}

	// skip blacklisted directories
	if info.IsDir() {
		_, ok := t.excludedDirs[info.Name()]
		if ok {
			return false, filepath.SkipDir
		}
	}

	// skip blacklisted extensions
	if !info.IsDir() {
		_, ok := t.excludedExts[filepath.Ext("."+info.Name())]
		if ok {
			return true, nil
		}
	}

	return false, nil
}

// StartCommandSurvey collect all required informations from user
func (t *Templating) StartCommandSurvey() error {
	var cd = &CommandData{}

	// start command prompts
	err := survey.Ask(t.getQuestions(), cd)
	if err != nil {
		logy.WithError(err).Error("command survey")
		return err
	}

	t.CommandData = cd

	return nil
}

func (t *Templating) startTemplateSurvey(surveys *Survey) error {
	questions, err := BuildSurveys(surveys)
	if err != nil {
		logy.WithError(err).Error("build surveys")
		return err
	}

	t.surveyResult = map[string]interface{}{}
	err = survey.Ask(questions, &t.surveyResult)
	if err != nil {
		logy.WithError(err).Error("start template survey")
		return err
	}

	logy.Debugf("Survey results %+v", t.surveyResult)

	return nil
}

// runSurveyTemplateHooks run all template hooks
func (t *Templating) runSurveyTemplateHooks() {
	for i, hook := range t.surveys.AfterHooks {
		ctx := logy.WithFields(logy.Fields{
			"cmd":  hook.Cmd,
			"args": hook.Args,
		})

		// check if cmd should be run
		if strings.TrimSpace(hook.Enabled) != "" {
			tpl, err := template.New(hook.Cmd).
				Delims("{", "}").
				Funcs(t.templateFuncMap).
				Parse("{if " + hook.Enabled + "}true{end}")

			if err != nil {
				ctx.WithError(err).Error("create template")
			}

			var buf bytes.Buffer
			tpl.Execute(&buf, t.TemplateData)
			if buf.String() != "true" {
				continue
			}
			ctx.Debug("skipped")
		}

		cmd := exec.Command(hook.Cmd, hook.Args...)
		cmd.Dir = path.Clean(t.CommandData.Path)
		// inherit process env
		cmd.Env = append(mapToEnvArray(t.surveyResult, envPrefix), os.Environ()...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			logy.WithError(err).Errorf("Command %d ('%s') could not be executed", i, hook.Cmd)
			log.Fatal(err)
		}
	}
}

// generateTempFuncs create helper funcs and getters based on the survey result
func (t *Templating) generateTempFuncs() {
	if t.surveyResult == nil {
		logy.Debug("could not generate getter functions due to empty survey results")
		return
	}

	// create getter functions for the survey results for easier access
	for key, val := range t.surveyResult {
		t.templateFuncMap["get"+casee.ToPascalCase(key)] = (func(v interface{}) func() interface{} {
			return func() interface{} {
				return v
			}
		})(val)
	}

	// create getter functions for the survey options for easier access
	for _, question := range t.surveys.Questions {
		t.templateFuncMap["get"+casee.ToPascalCase(question.Name+"Question")] = (func(v Question) func() interface{} {
			return func() interface{} {
				return question
			}
		})(question)
	}

}

// Run the command
func (t *Templating) Run() error {
	tpl := t.getTemplateByName(t.CommandData.Template)

	if tpl == nil {
		return fmt.Errorf("template %s could not be found", t.CommandData.Template)
	}

	// clone repository
	startTimeClone := time.Now()
	cloneSpinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	cloneSpinner.Suffix = "Cloning repository..."
	cloneSpinner.Start()
	err := t.cloneRepo(tpl.Url, t.CommandData.Path)
	cloneSpinner.Stop()

	if err != nil {
		logy.WithError(err).Error("clone")
		return err
	}

	surveyFile := path.Join(t.CommandData.Path, t.configName)
	ctx := logy.WithFields(logy.Fields{
		"path": surveyFile,
	})

	surveys, err := ReadSurveyConfig(surveyFile)
	if err != nil {
		ctx.WithError(err).Error("read survey config")
		return err
	}
	t.surveys = surveys

	err = t.startTemplateSurvey(surveys)
	if err != nil {
		ctx.WithError(err).Error("start template survey")
		return err

	}

	// spinner progress
	templatingSpinner := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	templatingSpinner.Suffix = "Processing templates..."
	templatingSpinner.Start()

	// start multiple routines
	t.startN(runtime.NumCPU())
	startTimeTemplating := time.Now()

	t.TemplateData = &TemplateData{
		t.CommandData,
		time.Now().Format(time.RFC3339),
		time.Now().Year(),
		t.Variables,
	}

	t.generateTempFuncs()

	renamings := map[string]string{}
	dirRemovings := []string{}

	// iterate through all directorys
	walkDirErr := filepath.Walk(
		t.CommandData.Path,
		func(path string, info os.FileInfo, err error) error {
			ctx := logy.WithFields(logy.Fields{
				"path": path,
				"size": info.Size(),
				"dir":  info.IsDir(),
			})

			if err != nil {
				ctx.WithError(err).Error("inside walk")
				return err
			}

			// preserve files from processing
			if !info.IsDir() {
				return nil
			}

			skipFile, skipDirErr := t.Skip(path, info)
			if skipFile {
				return nil
			}
			if skipDirErr != nil {
				return skipDirErr
			}

			defer func() {
				if r := recover(); r != nil {
					ctx.Error("directory templating error")
				}
			}()

			// Template directory
			tplDir, err := template.New(path).
				Delims(startNameDelim, endNameDelim).
				Funcs(t.templateFuncMap).
				Parse(info.Name())

			if err != nil {
				ctx.WithError(err).Error("create template for directory")
				return err
			}

			var dirNameBuffer bytes.Buffer
			err = tplDir.Execute(&dirNameBuffer, t.TemplateData)
			if err != nil {
				ctx.WithError(err).Error("execute template for directory")
				return err
			}

			newDirectory := dirNameBuffer.String()
			newPath := filepath.Join(filepath.Dir(path), newDirectory)

			// when directory contains a condition
			// order is irrelevant
			if strings.TrimSpace(newPath) == "" {
				dirRemovings = append(dirRemovings, newPath)
			} else if path != newPath {
				renamings[path] = newPath
			}

			return nil
		})

	if walkDirErr != nil {
		logy.WithError(walkDirErr).Error("walk dir")
		return walkDirErr
	}

	// rename and remove changed dirs
	for oldPath, newPath := range renamings {
		os.Rename(oldPath, newPath)
		os.RemoveAll(oldPath)
	}

	// remove directories which are evaluated to empty string
	for _, path := range dirRemovings {
		os.RemoveAll(path)
	}

	// iterate through all files
	walkErr := filepath.Walk(t.CommandData.Path,
		func(path string, info os.FileInfo, err error) error {
			ctx := logy.WithFields(logy.Fields{
				"path": path,
				"size": info.Size(),
				"dir":  info.IsDir(),
			})

			if err != nil {
				ctx.WithError(err).Error("inside walk")
				return err
			}

			skipFile, skipDirErr := t.Skip(path, info)
			if skipFile {
				ctx.Debug("skip file")
				return nil
			}
			if skipDirErr != nil {
				ctx.Debug("skip directory")
				return skipDirErr
			}

			// preserve directorys from processing
			if info.IsDir() {
				return nil
			}

			// template file
			t.ch <- func() {
				defer func() {
					if r := recover(); r != nil {
						ctx.Error("templating error")
					}
				}()

				// Template filename
				tplFilename, err := template.New(path).
					Delims(startNameDelim, endNameDelim).
					Funcs(t.templateFuncMap).
					Parse(info.Name())

				if err != nil {
					ctx.WithError(err).Error("create template for filename")
					return
				}

				var filenameBuffer bytes.Buffer
				err = tplFilename.Execute(&filenameBuffer, t.TemplateData)
				if err != nil {
					ctx.WithError(err).Error("execute template for filename")
					return
				}

				newFilename := filenameBuffer.String()
				newPath := filepath.Join(filepath.Dir(path), newFilename)

				// when filename contains a condition
				if strings.TrimSpace(newPath) == "" {
					err := os.Remove(path)
					if err != nil {
						ctx.WithError(err).Error("delete")
					}
					return
				}

				dat, err := ioutil.ReadFile(path)

				if err != nil {
					ctx.WithError(err).Error("read")
					return
				}

				// Template file content
				tmpl, err := template.New(newPath).
					Delims(startContentDelim, endContentDelim).
					Funcs(t.templateFuncMap).
					Parse(string(dat))

				if err != nil {
					ctx.WithError(err).Error("parse")
					return
				}

				f, err := os.Create(newPath)

				if err != nil {
					ctx.WithError(err).Error("create")
					return
				}

				defer f.Close()

				err = tmpl.Execute(f, t.TemplateData)

				if err != nil {
					ctx.WithError(err).Error("template")
					return
				}

				// remove old file when the name was changed
				if path != newPath {
					ctx.Debug("delete due to different filename")
					os.Remove(path)
				}
			}

			return nil
		})

	if walkErr != nil {
		return walkErr
	}

	t.stop()
	templatingSpinner.Stop()

	commandGitHook := githook.New(
		githook.WithCommandData(
			&githook.CommandData{
				Path:  t.CommandData.Path,
				Hooks: githook.Hooks,
			},
		),
	)

	logy.Debug("create git hooks")

	err = commandGitHook.Run()
	if err != nil {
		logy.WithError(err).Error("Could not create git hooks")
		return err
	}

	logy.Debug("execute template hooks")

	startTimeHooks := time.Now()

	if t.surveyResult != nil {
		t.runSurveyTemplateHooks()
	} else {
		logy.Debug("skip template survey")
	}

	// print summary
	totalCloneDuration := time.Since(startTimeClone).Seconds()
	totalTemplatingDuration := time.Since(startTimeTemplating).Seconds()
	totalHooksDuration := time.Since(startTimeHooks).Seconds()
	totalDuration := totalCloneDuration + totalTemplatingDuration + totalHooksDuration
	fmt.Printf("\nClone: %s sec \nTemplating: %s sec\nHooks: %s\nTotal: %s sec",
		strconv.FormatFloat(totalCloneDuration, 'f', 2, 64),
		strconv.FormatFloat(totalTemplatingDuration, 'f', 2, 64),
		strconv.FormatFloat(totalHooksDuration, 'f', 2, 64),
		strconv.FormatFloat(totalDuration, 'f', 2, 64),
	)

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

func mapToEnvArray(s map[string]interface{}, prefix string) []string {
	prefix = strings.ToUpper(prefix)
	array := []string{}
	for name, a := range s {
		envName := strings.ToUpper(name)
		switch v := a.(type) {
		case []string:
			array = append(array, fmt.Sprintf("%s_%s=%s", prefix, envName, strings.Join(v, ",")))
		case string:
			array = append(array, fmt.Sprintf("%s_%s=%s", prefix, envName, a))
		}
	}
	return array
}

// toMap returns a map from slice.
func toMap(s []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range s {
		m[v] = struct{}{}
	}
	return m
}
