package template

import (
	"io/ioutil"

	logy "github.com/apex/log"
	yaml "gopkg.in/yaml.v2"
)

// Question represents a question in the yml file
type Question struct {
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Default  interface{} `json:"default"`
	Options  []string    `json:"options"`
	Message  string      `json:"message"`
	Required bool        `json:"required"`
	Help     string      `json:"help"`
}

// Hook represent a hook in the yml file
type Hook struct {
	Cmd     string   `json:"cmd"`
	Args    []string `json:"args"`
	Enabled string   `json:"enabled"`
}

// Survey represents in the yml file
type Survey struct {
	Questions  []Question        `yaml:"questions"`
	AfterHooks []Hook            `yaml:"afterHooks"`
	Variables  map[string]string `yaml:"variables"`
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

	return survey, nil
}
