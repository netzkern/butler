package commands

import (
	"encoding/json"
	"fmt"

	log "github.com/netzkern/butler/logger"
	"github.com/parnurzeal/gorequest"
)

// https://docs.atlassian.com/atlassian-confluence/REST/6.5.0

type (
	Confluence struct {
		Space  Space
		Auth   BasicAuth
		Logger *log.Logger
	}
	BasicAuth struct {
		Username, Password string
	}
	Space struct {
		Name string `json:"name"`
	}
)

func (c *Confluence) Run() error {
	mJSON, _ := json.Marshal(c.Space)

	request := gorequest.New().SetBasicAuth(c.Auth.Username, c.Auth.Password)
	resp, _, errs := request.Post("http://example.com").
		Send(string(mJSON)).
		End()

	if len(errs) > 0 {
		c.Logger.Errorf("could not create Space %+v", errs)
		return fmt.Errorf("could not create Space")
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("could not create Space Status %v", resp.StatusCode)
	}

	return nil
}
