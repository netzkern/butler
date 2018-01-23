package template

import (
	"io/ioutil"

	logy "github.com/apex/log"
	yaml "gopkg.in/yaml.v2"
)

type Question struct {
	Type     string      `json:"type"`
	Name     string      `json:"name"`
	Default  interface{} `json:"default"`
	Options  []string    `json:"options"`
	Message  string      `json:"message"`
	Required bool        `json:"required"`
	Help     string      `json:"help"`
}

type Survey struct {
	Questions []Question `json:"questions"`
}

// ReadSurveyConfig read the config and return a survey
func ReadSurveyConfig(path string) (*Survey, error) {
	survey := &Survey{}
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		logy.Warn("butler-survey.yml could not be found")
		return survey, err
	}

	err = yaml.Unmarshal(dat, &survey)
	if err != nil {
		logy.Errorf("could not unmarshal %s", err.Error())
		return survey, err
	}

	return survey, nil
}
