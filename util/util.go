package util

import (
	"log"
	"regexp"
)

func NormalizeProjectName(example string) string {
	reg, err := regexp.Compile("[^a-zA-Z0-9-]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(example, "")
}
