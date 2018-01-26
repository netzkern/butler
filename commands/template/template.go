package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	"github.com/netzkern/butler/utils"
	"github.com/pinzolo/casee"
	"github.com/pkg/errors"
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
		Templates    []config.Template
		Variables    map[string]string
		CommandData  *CommandData
		TemplateData *TemplateData

		configName      string
		excludedDirs    map[string]struct{}
		excludedExts    map[string]struct{}
		ch              chan func()
		wg              sync.WaitGroup
		surveyResult    map[string]interface{}
		templateFuncMap template.FuncMap
		surveys         *Survey
		dirRenamings    map[string]string
		dirRemovings    []string
		gitDir          string
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
		dirRenamings: map[string]string{},
		dirRemovings: []string{},
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

// WithGitDir option.
func WithGitDir(dir string) Option {
	return func(t *Templating) {
		t.gitDir = dir
	}
}

// WithVariables option.
func WithVariables(s map[string]string) Option {
	return func(t *Templating) {
		t.Variables = s
	}
}

// SetConfigName option.
func SetConfigName(s string) Option {
	return func(t *Templating) {
		t.configName = s
	}
}

// WithTemplates option.
func WithTemplates(s []config.Template) Option {
	return func(t *Templating) {
		t.Templates = s
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

// unpackTemplate clone a repo to the dst
func (t *Templating) unpackTemplate(repoURL string, dest string) error {
	_, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL: repoURL,
	})

	if err != nil {
		return err
	}

	// remove git files
	err = os.RemoveAll(filepath.Join(dest, ".git"))
	if err != nil {
		return errors.Wrap(err, "remove all failed")
	}

	return err
}

func (t *Templating) packTemplate(tempDir, dest string) error {
	// remove butler files
	butlerSurveyFile := path.Clean(filepath.Join(tempDir, t.configName))
	if utils.Exists(butlerSurveyFile) {
		err := os.Remove(butlerSurveyFile)
		if err != nil {
			return errors.Wrap(err, "remove butler failes failed")
		}
	}

	// move files from temp to cd
	if dest == "." {
		err := utils.MoveDir(tempDir, ".")
		if err != nil {
			return errors.Wrap(err, "move failed")
		}
		err = os.RemoveAll(tempDir)
		if err != nil {
			return errors.Wrap(err, "remove all failed")
		}
	} else {
		err := os.Rename(tempDir, dest)
		if err != nil {
			return errors.Wrap(err, "rename failed")
		}
	}

	return nil
}

func (t *Templating) cleanTemplate(tempDir string) error {
	err := os.RemoveAll(tempDir)
	if err != nil {
		return errors.Wrap(err, "remove all failed")
	}

	return nil
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
	name := info.Name()
	// ignore hidden dirs and files
	if len(name) > 1 && strings.HasPrefix(name, ".") {
		if info.IsDir() {
			return false, filepath.SkipDir
		}
		return true, nil
	}

	// skip blacklisted directories
	if info.IsDir() {
		_, ok := t.excludedDirs[name]
		if ok {
			return false, filepath.SkipDir
		}
	}

	// skip blacklisted extensions
	if !info.IsDir() {
		_, ok := t.excludedExts[filepath.Ext("."+name)]
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
		return errors.Wrap(err, "command survey")
	}

	t.CommandData = cd
	t.CommandData.Path = path.Clean(cd.Path)

	return nil
}

func (t *Templating) confirmPackTemplate() (bool, error) {
	packTemplate := false
	prompt := &survey.Confirm{
		Message: fmt.Sprintf("Do you really want to checkout to '%s' ?", t.CommandData.Path),
	}

	err := survey.AskOne(prompt, &packTemplate, nil)
	if err != nil {
		return false, errors.Wrap(err, "confirm failed")
	}

	return packTemplate, nil
}

func (t *Templating) startTemplateSurvey(surveys *Survey) error {
	questions, err := BuildSurveys(surveys)
	if err != nil {
		return errors.Wrap(err, "build surveys")
	}

	t.surveyResult = map[string]interface{}{}
	err = survey.Ask(questions, &t.surveyResult)
	if err != nil {
		return errors.Wrap(err, "start template survey")
	}

	logy.Debugf("survey results %+v", t.surveyResult)

	return nil
}

// runSurveyTemplateHooks run all template hooks
func (t *Templating) runSurveyTemplateHooks(cmdDir string) error {
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
				return errors.Wrap(err, "create template")
			}

			var buf bytes.Buffer
			tpl.Execute(&buf, t.TemplateData)
			if buf.String() != "true" {
				continue
			}
			ctx.Debug("skipped")
		}

		cmd := exec.Command(hook.Cmd, hook.Args...)
		cmd.Dir = cmdDir
		// inherit process env
		cmd.Env = append(mapToEnvArray(t.surveyResult, envPrefix), os.Environ()...)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin
		err := cmd.Run()
		if err != nil {
			return errors.Wrapf(err, "Command %d ('%s') could not be executed", i, hook.Cmd)
		}
	}

	return nil
}

// generateTempFuncs create helper funcs and getters based on the survey result
func (t *Templating) generateTempFuncs() {
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

func (t *Templating) walkDirectories(path string, info os.FileInfo, err error) error {
	ctx := logy.WithFields(logy.Fields{
		"path": path,
		"size": info.Size(),
		"dir":  info.IsDir(),
	})

	if err != nil {
		err := errors.Wrap(err, "walk failed")
		ctx.WithError(err)
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
			ctx.Error("directory templating panic")
		}
	}()

	// Template directory
	tplDir, err := template.New(path).
		Delims(startNameDelim, endNameDelim).
		Funcs(t.templateFuncMap).
		Parse(info.Name())

	if err != nil {
		err := errors.Wrap(err, "create template for directory")
		ctx.WithError(err)
		return err
	}

	var dirNameBuffer bytes.Buffer
	err = tplDir.Execute(&dirNameBuffer, t.TemplateData)
	if err != nil {
		err := errors.Wrap(err, "execute template for directory")
		ctx.WithError(err)
		return err
	}

	newDirectory := dirNameBuffer.String()
	newPath := filepath.Join(filepath.Dir(path), newDirectory)

	// when directory template expression was evaluated to empty
	if strings.TrimSpace(newPath) == "" {
		t.dirRemovings = append(t.dirRemovings, newPath)
	} else if path != newPath {
		t.dirRenamings[path] = newPath
	}

	return nil
}

func (t *Templating) walkFiles(path string, info os.FileInfo, err error) error {
	ctx := logy.WithFields(logy.Fields{
		"path": path,
		"size": info.Size(),
		"dir":  info.IsDir(),
	})

	if err != nil {
		err := errors.Wrap(err, "walk failed")
		ctx.WithError(err)
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
				ctx.Error("file templating panic")
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
}

// Run the command
func (t *Templating) Run() (err error) {
	tpl := t.getTemplateByName(t.CommandData.Template)

	if tpl == nil {
		err = errors.Errorf("template %s could not be found", t.CommandData.Template)
		return
	}

	startTimeClone := time.Now()
	cloneSpinner := defaultSpinner("Cloning repository...")
	cloneSpinner.Start()

	// unpack template
	tempDir, err := ioutil.TempDir(t.gitDir, "butler")
	if err != nil {
		err = errors.Wrap(err, "create temp folder failed")
	}

	err = t.unpackTemplate(tpl.Url, tempDir)

	endTimeClone := time.Since(startTimeClone).Seconds()

	defer func() {
		r := recover()
		if err != nil || r != nil {
			err = t.cleanTemplate(tempDir)
			if err != nil {
				logy.WithError(err).Error("clean template failed")
			}
			logy.Debug("clean template")
		}
	}()

	cloneSpinner.Stop()

	if err != nil {
		logy.WithError(err).Error("clone")
		return
	}

	surveyFilePath := path.Join(tempDir, t.configName)
	ctx := logy.WithFields(logy.Fields{
		"path": surveyFilePath,
	})

	surveys, err := ReadSurveyConfig(surveyFilePath)
	if err != nil {
		ctx.WithError(err).Error("read survey config")
		return
	}
	t.surveys = surveys

	err = t.startTemplateSurvey(surveys)
	if err != nil {
		ctx.WithError(err).Error("start template survey")
		return
	}

	// spinner progress
	templatingSpinner := defaultSpinner("Processing templates...")
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

	// create getter for survey result
	if t.surveyResult != nil {
		t.generateTempFuncs()
	}

	logy.Debugf("dir walk in path %s", tempDir)

	// iterate through all directorys
	walkDirErr := filepath.Walk(tempDir, t.walkDirectories)

	if walkDirErr != nil {
		logy.WithError(walkDirErr).Error("walk dir")
		err = walkDirErr
		return
	}

	// rename and remove changed dirs from walk
	for oldPath, newPath := range t.dirRenamings {
		os.Rename(oldPath, newPath)
		os.RemoveAll(oldPath)
	}

	// remove directories which are evaluated to empty string from walk
	for _, path := range t.dirRemovings {
		os.RemoveAll(path)
	}

	logy.Debugf("file walk in path %s", tempDir)

	// iterate through all files
	walkErr := filepath.Walk(tempDir, t.walkFiles)

	if walkErr != nil {
		err = walkErr
		return
	}

	t.stop()
	templatingSpinner.Stop()

	endTimeTemplating := time.Since(startTimeTemplating).Seconds()

	startTimeHooks := time.Now()

	if t.surveyResult != nil {
		logy.Debug("execute template hooks")
		err = t.runSurveyTemplateHooks(tempDir)
		if err != nil {
			logy.WithError(err).Error("template hooks failed")
			return
		}
	} else {
		logy.Debug("skip template hooks")
	}

	endTimeHooks := time.Since(startTimeHooks).Seconds()

	confirmed, err := t.confirmPackTemplate()
	if confirmed {
		err = t.packTemplate(tempDir, t.CommandData.Path)
		if err != nil {
			logy.WithError(err).Error("pack template failed")
			return
		}
	} else {
		err = errors.New("abort templating")
		return
	}

	commandGitHook := githook.New(
		githook.WithGitDir(t.gitDir),
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
		logy.WithError(err).Error("could not create git hooks")
		return
	}

	// print summary
	totalDuration := endTimeClone + endTimeTemplating + endTimeHooks
	fmt.Printf("\nClone: %s sec \nTemplating: %s sec\nHooks: %s\nTotal: %s sec",
		strconv.FormatFloat(endTimeClone, 'f', 2, 64),
		strconv.FormatFloat(endTimeTemplating, 'f', 2, 64),
		strconv.FormatFloat(endTimeHooks, 'f', 2, 64),
		strconv.FormatFloat(totalDuration, 'f', 2, 64),
	)

	return
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

func defaultSpinner(suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = suffix
	return s
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
