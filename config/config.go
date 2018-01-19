package config

import (
	"io/ioutil"
	"net/http"
	"os"

	logy "github.com/apex/log"
	yaml "gopkg.in/yaml.v2"
)

var (
	butlerConfigURLEnv = "BUTLER_CONFIG_URL"
	defaultConfigURL   = os.Getenv(butlerConfigURLEnv)
)

type Template struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Config struct {
	Templates []Template        `json:"templates"`
	Variables map[string]string `json:"variables"`
}

func downloadConfig() ([]byte, error) {
	resp, err := http.Get(defaultConfigURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func ParseConfig() *Config {
	cfg := &Config{}
	dat, err := ioutil.ReadFile("butler.yml")
	if err != nil {
		logy.Info("butler.yml could not be found")

		if defaultConfigURL == "" {
			logy.Fatalf("environment Variable %s was not set", butlerConfigURLEnv)
		}

		logy.Infof("downloading defaut config butler.yml from %+v", defaultConfigURL)

		dat, err = downloadConfig()
		if err != nil {
			logy.Fatalf("butler.yml could not be downloaded from %+v", defaultConfigURL)
		}
	}

	err = yaml.Unmarshal(dat, &cfg)
	if err != nil {
		panic(err)
	}

	return cfg
}
