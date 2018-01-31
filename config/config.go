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

// Template represents the project template with informations about location
// and name
type Template struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Config represents the butler config
type Config struct {
	Templates []Template             `json:"templates"`
	Variables map[string]interface{} `json:"variables"`
}

// downloadConfig download the full file from web
func downloadConfig(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

// ParseConfig returns the yaml parsed config
func ParseConfig(filename string) *Config {
	cfg := &Config{
		Templates: []Template{},
		Variables: map[string]interface{}{},
	}
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		logy.Warnf("%s could not be found", filename)
	}

	cfgExt := &Config{
		Templates: []Template{},
		Variables: map[string]interface{}{},
	}

	err = yaml.Unmarshal(dat, &cfg)
	if err != nil {
		logy.Fatalf("could not unmarshal %s", err.Error())
	}

	// check for external configUrl in env
	if defaultConfigURL != "" {
		logy.Infof("loading config from %s=%s", butlerConfigURLEnv, defaultConfigURL)

		u, err := url.ParseRequestURI(defaultConfigURL)
		if err != nil {
			logy.Fatalf("invalid url in %+v", butlerConfigURLEnv)
		}

		dat, err := downloadConfig(u.String())
		if err != nil {
			logy.Fatalf("%s could not be downloaded from %+v", filename, defaultConfigURL)
		}
		err = yaml.Unmarshal(dat, &cfgExt)
		if err != nil {
			logy.Fatalf("could not unmarshal %s", err.Error())
		}

		cfg = mergeConfigs(cfg, cfgExt)
	}

	return cfg
}

// merge extend b with a, whereby b take precedence
func mergeConfigs(a, b *Config) *Config {
	// merge variables
	for k, v := range b.Variables {
		a.Variables[k] = v
	}

	// merge templates
	for _, v := range b.Templates {
		found := false
		for j, v2 := range a.Templates {
			if v.Name == v2.Name {
				a.Templates[j] = v2
				found = true
				break
			}
		}
		if !found {
			a.Templates = append(a.Templates, v)
		}
	}

	return a
}
