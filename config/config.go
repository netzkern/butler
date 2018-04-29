package config

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path"

	logy "github.com/apex/log"
	"github.com/kelseyhightower/envconfig"
	"github.com/netzkern/butler/utils"
	yaml "gopkg.in/yaml.v2"
)

type (
	// Template represents the project template with informations about location
	// and name
	Template struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	ConfluencePage struct {
		Name     string           `json:"name"`
		Children []ConfluencePage `json:"children"`
	}
	ConfluenceTemplate struct {
		Name  string           `json:"name"`
		Pages []ConfluencePage `json:"pages"`
	}
	Confluence struct {
		Templates []ConfluenceTemplate `json:"templates"`
	}
	// Config represents the butler config
	Config struct {
		Templates            []Template             `json:"templates"`
		Variables            map[string]interface{} `json:"variables"`
		ConfigURL            string                 `split_words:"true"`
		ConfluenceURL        string                 `split_words:"true"`
		ConfluenceAuthMethod string                 `split_words:"true"`
		ConfluenceBasicAuth  []string               `split_words:"true"`
		Confluence           Confluence             `json:"confluence"`
	}
)

// downloadConfig download the full file from web
func downloadConfig(path string) ([]byte, error) {
	resp, err := http.Get(path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func ParseConfigFile(filename string) (*Config, error) {
	dat, err := ioutil.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Templates: []Template{},
		Variables: map[string]interface{}{},
	}

	err = yaml.Unmarshal(dat, &cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// ParseConfig returns the yaml parsed config
func ParseConfig(filename string) *Config {
	ctx := logy.WithFields(logy.Fields{
		"config": filename,
	})

	usr, err := user.Current()

	if err != nil {
		ctx.Warnf("couldn't retrieve current user, see %s", err)
	}

	cfg := &Config{
		Templates: []Template{},
		Variables: map[string]interface{}{},
	}

	var homeCfg *Config

	// find user config
	if usr != nil {
		homePath := path.Join(usr.HomeDir, filename)

		if _, err = os.Stat(homePath); !os.IsNotExist(err) {
			homeCtx := logy.WithFields(logy.Fields{
				"config": homePath,
			})

			homeCfg, err = ParseConfigFile(homePath)

			if err != nil {
				homeCtx.Warnf(
					"couldn't load user config file from, see %s",
					err.Error())
			} else {
				cfg = mergeConfigs(cfg, homeCfg)
			}
		}
	}

	// find local config
	if _, err = os.Stat(filename); !os.IsNotExist(err) {
		localCfg, err := ParseConfigFile(filename)

		if err != nil {
			ctx.Warnf("couldn't load config from, see %s", err.Error())
		} else {
			cfg = mergeConfigs(cfg, localCfg)
		}
	}

	err = envconfig.Process("butler", cfg)

	if err != nil {
		ctx.Fatalf("could not inject env variables %s", err.Error())
	}

	// find config in ENV
	if cfg.ConfigURL != "" {
		cfgExt := &Config{
			Templates: []Template{},
			Variables: map[string]interface{}{},
		}

		ctx.WithField("url", cfg.ConfigURL).
			Debugf("loading external config")

		var dat []byte

		if utils.Exists(cfg.ConfigURL) {
			if dat, err = ioutil.ReadFile(cfg.ConfigURL); err != nil {
				ctx.WithField("url", cfg.ConfigURL).
					Fatalf("could not read config from file system")
			}
		} else if dat, err = downloadConfig(cfg.ConfigURL); err != nil {
			ctx.WithField("url", cfg.ConfigURL).
				Errorf("could not read config from external url")
		}

		err = yaml.Unmarshal(dat, cfgExt)

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
