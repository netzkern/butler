package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"text/template"
	"time"

	logy "github.com/apex/log"
	"github.com/blang/semver"
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

var (
	errManualTermination = errors.New("manual termination")
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
		Variables    map[string]interface{}
		CommandData  *CommandData
		TemplateData *TemplateData
		TaskTracker  *TaskTracker

		configName      string
		excludedDirs    map[string]struct{}
		excludedExts    map[string]struct{}
		ch              chan func()
		chErr           chan error
		wg              sync.WaitGroup
		surveyResult    map[string]interface{}
		templateFuncMap template.FuncMap
		templateConfig  *Survey
		dirRenamings    map[string]string
		dirRemovings    []string
		cwd             string
		butlerVersion   semver.Version
	}
	// TemplateData basic template data
	TemplateData struct {
		Project *CommandData
		Date    string
		Year    int
		Vars    map[string]interface{}
	}
)

// Option function.
type Option func(*Templating)

// New with the given options.
func New(options ...Option) *Templating {
	t := &Templating{
		excludedDirs: toMap(ExcludedDirs),
		excludedExts: toMap(BinaryFileExt),
		// the buffer size is equivalent to the worker size this reduce the chance of wasted (blocking) resources.
		ch:           make(chan func(), runtime.NumCPU()),
		chErr:        make(chan error, runtime.NumCPU()),
		dirRenamings: map[string]string{},
		dirRemovings: []string{},
		TaskTracker:  NewTaskTracker(),
	}

	for _, o := range options {
		o(t)
	}

	t.templateFuncMap = template.FuncMap{
		// string helper funcs
		"toCamelCase":  casee.ToCamelCase,
		"toPascalCase": casee.ToPascalCase,
		"toSnakeCase":  casee.ToSnakeCase,
		"toLowerCase":  strings.ToLower,
		"toUpperCase":  strings.ToUpper,
		"join":         strings.Join,
		"replace":      strings.Replace,
		"contains":     strings.Contains,
		"index":        strings.Index,
		"repeat":       strings.Repeat,
		"split":        strings.Split,
		// path
		"joinPath": filepath.Join,
		"relPath":  filepath.Rel,
		"basePath": filepath.Base,
		"extPath":  filepath.Ext,
		"absPath":  filepath.Abs,
		// regexp
		"regex": func(str string) *regexp.Regexp {
			return regexp.MustCompile(str)
		},
		// generators
		"uuid":      uuid.NewV4,
		"randomInt": func(min, max int) int { return rand.Intn(max-min) + min },
		//environment
		"cwd": func() string { return t.CommandData.Path },
		"env": func(name string) string { return os.Getenv(name) },
	}

	return t
}

// WithCwd option.
func WithCwd(dir string) Option {
	return func(t *Templating) {
		t.cwd = dir
	}
}

// WithVariables option.
func WithVariables(s map[string]interface{}) Option {
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

// WithButlerVersion option.
func WithButlerVersion(s string) Option {
	return func(t *Templating) {
		t.butlerVersion = semver.MustParse(s)
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

// unpackGitRepository clone a repo to the dst
func (t *Templating) unpackGitRepository(templatePath string, dest string) error {
	logy.Debugf("unpack template from %s to %s", templatePath, dest)

	_, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL: templatePath,
	})

	if err != nil {
		return err
	}

	// remove git files
	err = os.RemoveAll(filepath.Join(dest, ".git"))
	if err != nil {
		return errors.Wrap(err, "git files from remote repository could not be removed")
	}

	return err
}

// unpackLocalGitRepository copy a local repository to the dst
func (t *Templating) unpackLocalGitRepository(tempDir string, dest string) error {
	logy.Debugf("unpack template from %s to %s", tempDir, dest)

	err := utils.MoveDir(tempDir, dest)
	if err != nil {
		return errors.Wrap(err, "local repository could not be copied")
	}

	// remove git files
	err = os.RemoveAll(filepath.Join(dest, ".git"))
	if err != nil {
		return errors.Wrap(err, "git files from local repository could not be removed")
	}

	return err
}

func (t *Templating) packTemplate(tempDir, dest string) error {
	// remove butler files
	butlerSurveyFile := path.Clean(filepath.Join(tempDir, t.configName))
	if utils.Exists(butlerSurveyFile) {
		err := os.Remove(butlerSurveyFile)
		if err != nil {
			return errors.Wrap(err, "butler files could not be removed")
		}
	}

	logy.Debugf("pack template from %s to %s", tempDir, dest)

	err := utils.CreateDirIfNotExist(dest)
	if err != nil {
		return errors.Wrap(err, "create dest dir failed")
	}

	err = utils.MoveDir(tempDir, dest)
	if err != nil {
		return errors.Wrap(err, "move failed")
	}
	err = os.RemoveAll(tempDir)
	if err != nil {
		return errors.Wrap(err, "remove all failed")
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
			Name: "Name",
			Prompt: &survey.Input{
				Message: "What's the project name?",
				Help:    "Allowed character [a-zA-Z0-9_-]{3,30}",
			},
			Validate: survey.ComposeValidators(
				survey.Required,
				survey.MinLength(3),
				survey.MaxLength(30),
				projectNameValidator,
			),
		},
		{
			Name: "Description",
			Prompt: &survey.Input{
				Message: "What's the project description?",
			},
		},
		{
			Name:     "Path",
			Validate: survey.Required,
			Prompt: &survey.Input{
				Message: "What's the destination?",
				Default: t.cwd,
				Help:    "The path to your new project",
			},
		},
	}

	return qs
}

// getTemplateQuestions return all required prompts
func (t *Templating) getTemplateQuestions() []*survey.Question {
	qs := []*survey.Question{
		{
			Name:     "Template",
			Validate: survey.Required,
			Prompt: &survey.Select{
				Message: "Please select a template",
				Options: t.getTemplateOptions(),
				Help:    "You can add additional templates in your config",
			},
		},
	}

	return qs
}

// skip returns an error when a directory should be skipped or true with a file
func (t *Templating) skip(path string, info os.FileInfo) (bool, error) {
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

// StartCommandSurvey ask the user for the template
func (t *Templating) StartCommandSurvey() error {
	var cd = &CommandData{}
	err := survey.Ask(t.getTemplateQuestions(), cd)
	if err != nil {
		return errors.Wrap(err, "template command survey")
	}
	t.CommandData = cd
	return nil
}

// startProjectSurvey ask the user for project details
func (t *Templating) startProjectSurvey() error {
	err := survey.Ask(t.getQuestions(), t.CommandData)
	if err != nil {
		return errors.Wrap(err, "command survey")
	}
	dest, err := filepath.Abs(t.CommandData.Path)
	if err != nil {
		return errors.Wrap(err, "dest path failed")
	}
	t.CommandData.Path = dest
	return nil
}

func (t *Templating) confirmPackTemplate(msg string) (bool, error) {
	packTemplate := false
	prompt := &survey.Confirm{
		Message: msg,
	}

	err := survey.AskOne(prompt, &packTemplate, nil)
	if err != nil {
		return false, errors.Wrap(err, "confirm failed")
	}

	return packTemplate, nil
}

func (t *Templating) startTemplateSurvey() error {
	questions, err := BuildSurvey(t.templateConfig)
	if err != nil {
		return errors.Wrap(err, "build survey from template config")
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
	for i, hook := range t.templateConfig.AfterHooks {
		ctx := logy.WithFields(logy.Fields{
			"cmd":  hook.Cmd,
			"args": hook.Args,
		})

		if strings.TrimSpace(hook.Enabled) != "" {
			dat, err := parseStringAsTemplateCondition(t.TemplateData, t.templateFuncMap, hook.Cmd, hook.Enabled)
			if err != nil {
				return errors.Wrap(err, "parse hook template")
			}
			if !dat {
				ctx.Debug("skipped")
				continue
			}
		}

		cmd := exec.Command(hook.Cmd, hook.Args...)
		// inherit process env
		cmd.Env = append(mapToEnvArray(t.surveyResult, envPrefix), os.Environ()...)
		cmd.Dir = cmdDir

		var spinner *spinner.Spinner
		if hook.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
		} else {
			spinner = defaultSpinner(fmt.Sprintf("Run hook '%s'...", hook.Name))
			spinner.Start()
		}

		err := cmd.Run()
		if spinner != nil {
			spinner.Stop()
		}
		if err != nil {
			ctx.WithError(err).Error("command failed")
			if hook.Required {
				return errors.Wrapf(err, "command %d ('%s') failed", i, hook.Cmd)
			}
		}
	}

	return nil
}

// parseSurveyTemplateVariables template all survey variables
func (t *Templating) parseSurveyTemplateVariables() error {
	for k, v := range t.Variables {
		if varString, ok := v.(string); ok {
			if strings.TrimSpace(varString) == "" {
				return nil
			}
			dat, err := parseStringAsTemplate(t.TemplateData, t.templateFuncMap, k, varString)
			if err != nil {
				return errors.Wrap(err, "parse variable template")
			}
			t.Variables[k] = dat
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
	for _, question := range t.templateConfig.Questions {
		t.templateFuncMap["get"+casee.ToPascalCase(question.Name+"Question")] = (func(v Question) func() interface{} {
			return func() interface{} {
				return question
			}
		})(question)
	}

}

// walkDirectories run over all directories and collect renamed, removed items
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

	skipFile, skipDirErr := t.skip(path, info)
	if skipFile {
		return nil
	}
	if skipDirErr != nil {
		return skipDirErr
	}

	// Template directory
	newDirectory, err := parseStringAsTemplate(t.TemplateData, t.templateFuncMap, path, info.Name())
	if err != nil {
		return errors.Wrap(err, "parse template for directory")
	}

	newPath := filepath.Join(filepath.Dir(path), newDirectory)

	// when directory template expression was evaluated to empty
	if strings.TrimSpace(newPath) == "" {
		t.dirRemovings = append(t.dirRemovings, newPath)
	} else if path != newPath {
		t.dirRenamings[path] = newPath
	}

	return nil
}

// walkFiles run over all files and collect renamed items. Text files are proceed with the template engine.
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

	skipFile, skipDirErr := t.skip(path, info)
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

	// send job to workers
	t.ch <- func() {
		err := t.templater(path, info.Name(), ctx)
		if err != nil {
			t.chErr <- err
		}
	}

	return nil
}

// templater is responsible to parse files, rename or delete files and write the output back to the file.
// t.TemplateData and t.templateFuncMap are read-only
func (t *Templating) templater(path, filename string, ctx *logy.Entry) error {
	newFilename, err := parseStringAsTemplate(t.TemplateData, t.templateFuncMap, path, filename)
	if err != nil {
		return errors.Wrap(err, "parse variable template")
	}

	newPath := filepath.Join(filepath.Dir(path), newFilename)

	// when filename condition was evaluated to false
	if strings.TrimSpace(newPath) == "" {
		err := os.Remove(path)
		if err != nil {
			ctx.WithError(err).Error("delete")
			return err
		}
		return nil
	}

	dat, err := ioutil.ReadFile(path)

	if err != nil {
		ctx.WithError(err).Error("read")
		return err
	}

	// Template file content
	tmpl, err := template.New(newPath).
		Delims(startContentDelim, endContentDelim).
		Funcs(t.templateFuncMap).
		Parse(string(dat))

	if err != nil {
		ctx.WithError(err).Error("parse")
		return err
	}

	f, err := os.Create(newPath)

	if err != nil {
		ctx.WithError(err).Error("create")
		return err
	}

	defer f.Close()

	err = tmpl.Execute(f, t.TemplateData)

	if err != nil {
		ctx.WithError(err).Error("template")
		return err
	}

	// remove old file when the name was changed
	if path != newPath {
		ctx.Debug("delete due to different filename")
		err := os.Remove(path)
		if err != nil {
			ctx.WithError(err).Error("delete")
			return err
		}
	}

	return nil
}

// Run the command
func (t *Templating) Run() (err error) {
	tempDir, err := ioutil.TempDir("", "butler")
	if err != nil {
		err = errors.Wrap(err, "create temp folder failed")
		return err
	}
	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		return errors.Wrap(err, "temp abs failed")
	}

	// remove template artifacts when a panic or error occur
	// when the user abort the process at the last step we will
	// assign an "errManualTerminiation" to indicate an intentional action.
	defer func() {
		r := recover()

		if r != nil {
			logy.Errorf("recover: %s", r)
		}

		if err != nil && err != errManualTermination {
			logy.WithError(err)
		}

		err := t.cleanTemplate(tempDir)
		if err != nil {
			logy.WithError(err).Error("remove template failed")
		}

		logy.Debug("remove template artifacts")
	}()

	tpl := t.getTemplateByName(t.CommandData.Template)

	if tpl == nil {
		err = errors.Errorf("template %s could not be found", t.CommandData.Template)
		return err
	}

	/**
	* Clone task
	 */
	t.TaskTracker.Track("Clone")
	cloneSpinner := defaultSpinner("Cloning repository...")
	cloneSpinner.Start()

	if utils.Exists(tpl.URL) {
		err = t.unpackLocalGitRepository(tpl.URL, tempDir)
	} else {
		err = t.unpackGitRepository(tpl.URL, tempDir)
	}

	t.TaskTracker.UnTrack("Clone")
	cloneSpinner.Stop()

	if err != nil {
		logy.WithError(err).Error("clone")
		return err
	}

	surveyFilePath := path.Join(tempDir, t.configName)
	ctx := logy.WithFields(logy.Fields{
		"path": surveyFilePath,
	})

	// template config isn't required
	if utils.Exists(surveyFilePath) {
		templateConfig, err := ReadSurveyConfig(surveyFilePath)
		if err != nil {
			ctx.WithError(err).Error("read survey config")
			return err
		}

		// check compatibility
		if templateConfig.ButlerVersion != "" {
			butlerVersions, err := semver.ParseRange(templateConfig.ButlerVersion)
			if err != nil {
				err := fmt.Errorf(
					"could not parse required butler version '%s'",
					templateConfig.ButlerVersion,
				)
				ctx.WithError(err).Error("invalid semver")
				return err
			}
			if !butlerVersions(t.butlerVersion) {
				err := fmt.Errorf(
					"the required butler version '%s' does not match with your current version '%s'",
					t.butlerVersion.String(),
					templateConfig.ButlerVersion,
				)
				ctx.WithError(err).Error("template requirement")
				return err
			}
		}

		if templateConfig.Deprecated {
			ctx.Infof("template is deprecated")
		}

		// overwrite local variables with template variables
		for k, v := range templateConfig.Variables {
			if _, ok := t.Variables[k]; ok {
				ctx.Debugf("overwrite local variable '%s' with template variable", k)
			}
			t.Variables[k] = v
		}

		t.templateConfig = templateConfig

		err = t.startProjectSurvey()
		if err != nil {
			ctx.WithError(err).Error("start project survey")
			return err
		}

		t.TemplateData = &TemplateData{
			t.CommandData,
			time.Now().Format(time.RFC3339),
			time.Now().Year(),
			t.Variables,
		}

		err = t.startTemplateSurvey()
		if err != nil {
			ctx.WithError(err).Error("start template survey")
			return err
		}

		t.generateTempFuncs()
		t.parseSurveyTemplateVariables()

	} else {
		err := t.startProjectSurvey()
		if err != nil {
			ctx.WithError(err).Error("start project survey")
			return err
		}
		t.TemplateData = &TemplateData{
			t.CommandData,
			time.Now().Format(time.RFC3339),
			time.Now().Year(),
			t.Variables,
		}
	}

	// spinner progress
	templatingSpinner := defaultSpinner("Processing templates...")
	templatingSpinner.Start()

	// start multiple routines
	t.startN(runtime.NumCPU())

	/**
	* Templating task
	 */
	t.TaskTracker.Track("Template")

	logy.Debugf("dir walk in path '%s'", tempDir)

	// iterate through all directorys
	walkDirErr := filepath.Walk(tempDir, t.walkDirectories)

	if walkDirErr != nil {
		logy.WithError(walkDirErr).Error("walk dir")
		err = walkDirErr
		return err
	}

	// rename and remove changed dirs from walk
	for oldPath, newPath := range t.dirRenamings {
		os.Rename(oldPath, newPath)
		err = os.RemoveAll(oldPath)
		if err != nil {
			logy.WithError(err).Error("remove all")
			return err
		}
	}

	// remove directories which are evaluated to empty string from walk
	for _, path := range t.dirRemovings {
		err = os.RemoveAll(path)
		if err != nil {
			logy.WithError(err).Error("remove all")
			return err
		}
	}

	logy.Debugf("file walk in path '%s'", tempDir)

	walkErr := filepath.Walk(tempDir, t.walkFiles)

	if walkErr != nil {
		err = walkErr
		return err
	}

	go t.stop()

	templatingSpinner.Stop()

	t.TaskTracker.UnTrack("Template")

	/**
	* Let's collect all template errors
	* It's blocked until chErr is closed
	 */
	var errCount int
	for range t.chErr {
		errCount++
	}

	var confirmMsg string
	if errCount == 0 {
		confirmMsg = fmt.Sprintf("Do you really want to checkout to '%s' ?", t.CommandData.Path)
	} else if errCount == 1 {
		confirmMsg = fmt.Sprintf("%s Do you really want to checkout to '%s' ?", fmt.Sprintf("We found %d error.", errCount), t.CommandData.Path)
	} else {
		confirmMsg = fmt.Sprintf("%s Do you really want to checkout to '%s' ?", fmt.Sprintf("We found %d errors.", errCount), t.CommandData.Path)
	}

	confirmed, err := t.confirmPackTemplate(confirmMsg)
	if err != nil {
		return err
	}

	if confirmed {
		err = t.packTemplate(tempDir, t.CommandData.Path)
		if err != nil {
			logy.WithError(err).Error("pack template failed")
			return err
		}
	} else {
		err = errManualTermination
		return err
	}

	/**
	* Template Hook task
	 */
	t.TaskTracker.Track("After hooks")

	if t.surveyResult != nil {
		logy.Debug("execute template hooks")
		err = t.runSurveyTemplateHooks(t.CommandData.Path)
		if err != nil {
			logy.WithError(err).Error("template hooks failed")
			return err
		}
	} else {
		logy.Debug("skip template hooks")
	}

	t.TaskTracker.UnTrack("After hooks")

	/**
	* Git hook task
	 */
	commandGitHook := githook.New(
		githook.WithCwd(t.cwd),
		githook.WithCommandData(
			&githook.CommandData{
				Path:  t.CommandData.Path,
				Hooks: githook.Hooks,
			},
		),
	)

	err = commandGitHook.Run()
	if err != nil {
		logy.WithError(err).Error("could not create git hooks")
		return err
	}

	t.TaskTracker.UnTrack("Git Hooks")

	return err
}

// startN starts n loops.
// This is called "Bounded Parallelism Pattern" we can limit these allocations by bounding the number of files read in parallel.
// If we would parallize all files and walk over a directory with many large files, this may allocate more memory than is available on the machine.
// Therefore this way is really fast and will never be a bottleneck.
func (t *Templating) startN(n int) {
	for i := 0; i < n; i++ {
		t.wg.Add(1)
		go t.start()
	}
}

// start loop.
// we start a fixed number of workers to distribute the work
func (t *Templating) start() {
	defer t.wg.Done()
	for fn := range t.ch {
		fn()
	}
}

// stop loop.
// after finishing the walk we can safely close the channel
// and unblock the "range" so that the workGroup can be finished
// chErr is closed to continue with the templating process
func (t *Templating) stop() {
	close(t.ch)
	t.wg.Wait()
	close(t.chErr)
}

func parseStringAsTemplate(templateData *TemplateData, funcMap template.FuncMap, name, text string) (string, error) {
	tpl, err := template.New(name).
		Delims(startNameDelim, endNameDelim).
		Funcs(funcMap).
		Parse(text)

	if err != nil {
		return "", errors.Wrap(err, "parse template as string")
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, templateData)
	if err != nil {
		return "", errors.Wrap(err, "execute template as string")
	}

	return buf.String(), err
}

func parseStringAsTemplateCondition(templateData *TemplateData, funcMap template.FuncMap, name, text string) (bool, error) {
	tpl, err := template.New(name).
		Delims(startNameDelim, endNameDelim).
		Funcs(funcMap).
		Parse("{if " + text + "}true{end}")

	if err != nil {
		return false, errors.Wrap(err, "parse template as condition")
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, templateData)
	if err != nil {
		return false, errors.Wrap(err, "execute template as condition")
	}

	return buf.String() == "true", err
}

// defaultSpinner create a spinner with good default settings
func defaultSpinner(suffix string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	s.Suffix = " " + suffix
	return s
}

// mapToEnvArray convert a map[string]interface{} to string arrays, compatible version to pass it to exec.Cmd
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

// projectNameValidator check if string is a valid project name
// project name is a combination of alpha-numeric,-,_ characters
func projectNameValidator(val interface{}) error {
	if str, ok := val.(string); ok {
		reg, err := regexp.Compile("([^a-zA-Z0-9_-]+)")
		if err != nil {
			return err
		}
		if reg.MatchString(str) {
			return errors.New("invalid name")
		}
	}
	return nil
}
