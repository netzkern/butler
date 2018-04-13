package template

import (
	"io/ioutil"

	logy "github.com/apex/log"
	validator "gopkg.in/go-playground/validator.v9"
	yaml "gopkg.in/yaml.v2"
)

// Question represents a question in the yml file
type Question struct {
	Type     string      `json:"type" validate:"required"`
	Name     string      `json:"name" validate:"required"`
	Default  interface{} `json:"default"`
	Options  []string    `json:"options"`
	Message  string      `json:"message" validate:"required"`
	Required bool        `json:"required"`
	Help     string      `json:"help"`
}

// Hook represent a hook in the yml file
type Hook struct {
	Name     string   `json:"name" validate:"required"`
	Cmd      string   `json:"cmd" validate:"required"`
	Args     []string `json:"args"`
	Verbose  bool     `json:"verbose"`
	Enabled  string   `json:"enabled"`
	Required bool     `json:"required"`
}

// Survey represents in the yml file
type Survey struct {
	Questions     []Question             `yaml:"questions" validate:"required,dive"`
	AfterHooks    []Hook                 `yaml:"afterHooks"`
	Variables     map[string]interface{} `yaml:"variables"`
	ButlerVersion string                 `yaml:"butlerVersion"`
	Deprecated    bool                   `yaml:"deprecated"`
}

// ReadSurveyConfig reads the config and return a new survey
func ReadSurveyConfig(path string) (*Survey, error) {
	survey := &Survey{}
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		logy.Warnf("survey config could not be found in %s", path)
		return survey, err
	}

	err = yaml.Unmarshal(dat, &survey)

	if err != nil {
		logy.Errorf("survey config could not be unmarshaled %s", err.Error())
		return survey, err
	}

	if err = validate(survey); err != nil {
		logy.WithError(err).Error("invalid template configuration")
		return nil, err
	}

	return survey, nil
}

func validate(cfg interface{}) error {
	validate := validator.New()
	validate.RegisterStructValidation(questionStructHasOptions, Question{})
	return validate.Struct(cfg)
}

func questionStructHasOptions(sl validator.StructLevel) {
	question := sl.Current().Interface().(Question)

	if (question.Type == "select" || question.Type == "multiselect") && len(question.Options) == 0 {
		sl.ReportError(question.Options, "options", "foptions", "optionsRequired", "")
	}
}
