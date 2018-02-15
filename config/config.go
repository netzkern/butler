package config

import (
	"io/ioutil"
	"net/http"
	"net/url"

	logy "github.com/apex/log"
	"github.com/kelseyhightower/envconfig"
	yaml "gopkg.in/yaml.v2"
)

// Template represents the project template with informations about location
// and name
type Template struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Config represents the butler config
type Config struct {
	Templates            []Template             `json:"templates"`
	Variables            map[string]interface{} `json:"variables"`
	ConfigURL            string                 `split_words:"true"`
	ConfluenceURL        string                 `split_words:"true"`
	ConfluenceAuthMethod string                 `split_words:"true"`
	ConfluenceBasicAuth  []string               `split_words:"true"`
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
	ctx := logy.WithFields(logy.Fields{
		"config": filename,
	})

	cfg := &Config{
		Templates: []Template{},
		Variables: map[string]interface{}{},
	}
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		ctx.Warnf("%s could not be found", filename)
	}

	cfgExt := &Config{
		Templates: []Template{},
		Variables: map[string]interface{}{},
	}

	err = yaml.Unmarshal(dat, &cfg)
	if err != nil {
		ctx.Fatalf("could not unmarshal %s", err.Error())
	}

	err = envconfig.Process("butler", cfg)
	if err != nil {
		ctx.Fatalf("could not inject env variables %s", err.Error())
	}

	// check for external configUrl in env
	if cfg.ConfigURL != "" {
		ctx.WithField("url", cfg.ConfigURL).
			Debugf("loading external config")

		u, err := url.ParseRequestURI(cfg.ConfigURL)
		if err != nil {
			ctx.WithField("url", cfg.ConfigURL).
				Fatalf("invalid url in BUTLER_CONFIG_URL")
		}

		dat, err := downloadConfig(u.String())

		if err != nil {
			ctx.WithField("url", cfg.ConfigURL).
				Fatalf("%s could not be downloaded from %+v", filename, cfg.ConfigURL)
		}
		err = yaml.Unmarshal(dat, &cfgExt)
		if err != nil {
			ctx.WithField("url", cfg.ConfigURL).
				Fatalf("could not unmarshal external config %s", err.Error())
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
