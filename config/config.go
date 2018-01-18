package config

import (
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type Template struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Config struct {
	Templates []Template        `json:"templates"`
	Logger    string            `json:"logger"`
	Variables map[string]string `json:"variables"`
}

func ParseConfig() *Config {
	cfg := &Config{}
	dat, err := ioutil.ReadFile("butler.yml")
	if err != nil {
		fmt.Println(fmt.Errorf("butler: butler.yml could not be found"))
		os.Exit(1)
	}

	err = yaml.Unmarshal(dat, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
