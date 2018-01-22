package config

import (
	"io/ioutil"
	"net/http"
	"net/url"
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

// downloadConfig download full file from web
func downloadConfig(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// ParseConfig returns the yaml parsed config
func ParseConfig() *Config {
	cfg := &Config{
		Templates: []Template{},
		Variables: map[string]string{},
	}
	dat, err := ioutil.ReadFile("butler.yml")
	if err != nil {
		logy.Warn("butler.yml could not be found")
	}

	cfgExt := &Config{
		Templates: []Template{},
		Variables: map[string]string{},
	}

	err = yaml.Unmarshal(dat, &cfg)
	if err != nil {
		logy.Fatalf("could not unmarshal %s", err.Error())
	}

	// check config in env
	if defaultConfigURL != "" {
		logy.Infof("loading config from %s=%s", butlerConfigURLEnv, defaultConfigURL)

		u, err := url.ParseRequestURI(defaultConfigURL)
		if err != nil {
			logy.Fatalf("invalid url in %+v", butlerConfigURLEnv)
		}

		dat, err := downloadConfig(u.String())
		if err != nil {
			logy.Fatalf("butler.yml could not be downloaded from %+v", defaultConfigURL)
		}
		err = yaml.Unmarshal(dat, &cfgExt)
		if err != nil {
			logy.Fatalf("could not unmarshal %s", err.Error())
		}

		// merge variables
		for k, v := range cfgExt.Variables {
			cfg.Variables[k] = v
		}

		// merge templates
		for _, v := range cfgExt.Templates {
			found := false
			for j, v2 := range cfg.Templates {
				if v.Name == v2.Name {
					cfg.Templates[j] = v2
					found = true
					break
				}
			}
			if !found {
				cfg.Templates = append(cfg.Templates, v)
				break
			}
		}
	}

	return cfg
}
