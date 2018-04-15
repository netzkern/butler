package template

import (
	"fmt"

	survey "gopkg.in/AlecAivazis/survey.v1"
)

// BuildSurvey generates a list of survey questions based on the template survey
func BuildSurvey(s *Survey) ([]*survey.Question, error) {
	var qs []*survey.Question
	for _, question := range s.Questions {
		switch question.Type {
		case "input":
			p := &survey.Input{
				Message: question.Message,
				Help:    question.Help,
			}
			if question.Default != nil {
				defaultValue, ok := question.Default.(string)
				if !ok {
					return nil, fmt.Errorf("default value must be a string on input questions")
				}
				p.Default = defaultValue
			}
			sqs := &survey.Question{
				Name:   question.Name,
				Prompt: p,
			}
			if question.Required {
				sqs.Validate = survey.Required
			}
			qs = append(qs, sqs)
		case "password":
			p := &survey.Password{
				Message: question.Message,
				Help:    question.Help,
			}
			sqs := &survey.Question{
				Name:   question.Name,
				Prompt: p,
			}
			if question.Required {
				sqs.Validate = survey.Required
			}
			qs = append(qs, sqs)
		case "confirm":
			p := &survey.Confirm{
				Message: question.Message,
				Help:    question.Help,
			}
			sqs := &survey.Question{
				Name:   question.Name,
				Prompt: p,
			}
			if question.Default != nil {
				defaultValue, ok := question.Default.(bool)
				if !ok {
					return nil, fmt.Errorf("default value must be a boolean on confirm questions")
				}
				p.Default = defaultValue
			}
			if question.Required {
				sqs.Validate = survey.Required
			}
			qs = append(qs, sqs)
		case "select":
			p := &survey.Select{
				Message: question.Message,
				Options: question.Options,
				Help:    question.Help,
			}
			if question.Default != nil {
				defaultValue, ok := question.Default.(string)
				if !ok {
					return nil, fmt.Errorf("default value must be a string on select questions")
				}
				p.Default = defaultValue
			}
			sqs := &survey.Question{
				Name:   question.Name,
				Prompt: p,
			}
			if question.Required {
				sqs.Validate = survey.Required
			}
			qs = append(qs, sqs)
		case "multiselect":
			p := &survey.MultiSelect{
				Message: question.Message,
				Options: question.Options,
				Help:    question.Help,
			}
			defaults := []string{}
			if question.Default != nil {
				defaultValue, ok := question.Default.([]interface{})
				if !ok {
					return nil, fmt.Errorf("default value must be an array of strings on multiselect questions")
				}
				for _, v := range defaultValue {
					s, ok := v.(string)
					if ok {
						defaults = append(defaults, s)
					}
				}
			}
			p.Default = defaults
			sqs := &survey.Question{
				Name:   question.Name,
				Prompt: p,
			}
			if question.Required {
				sqs.Validate = survey.Required
			}
			qs = append(qs, sqs)
		default:
			return nil, fmt.Errorf("invalid prompt type %s", question.Type)
		}
	}

	return qs, nil
}
