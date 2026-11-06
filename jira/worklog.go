package jira

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-resty/resty/v2"
)

type Worklog struct {
	// the API does not request the Key in the body, but in the URL.
	// Just to make the code simpler we will have the struct containing the key,
	// it will send it also in the JSON body, that should not have any negative effects
	Key string `json:"key"`
	Comment
	Started          string `json:"started"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}

type Comment struct {
	Type    string        `json:"type"`
	Version int           `json:"version"`
	Content []CommentNode `json:"content,omitempty"`
}

type CommentNode struct {
	Type    string        `json:"type"`
	Text    string        `json:"text,omitempty"`
	Content []CommentNode `json:"content,omitempty"`
}

func (worklog Worklog) Save(config config.Config) error {
	entryJSON, err := json.Marshal(worklog)
	if err != nil {
		return fmt.Errorf("error marshalling Jira worklog entry to JSON: %v", err)
	}
	httpClient := resty.New()
	worklogURL, _ := url.JoinPath(config.JiraUrl, "/rest/api/3/issue", worklog.Key, "worklog")
	resp, err := httpClient.R().
		SetBasicAuth(config.JiraUser, config.JiraToken).
		SetHeader("Content-Type", "application/json; charset=utf-8").
		SetBody(entryJSON).
		Post(worklogURL)

	if err != nil || resp.StatusCode() != 201 {
		return fmt.Errorf(
			"could not save worklog entry in Jira. Error: '%v'. HTTP status code: %v. HTTP response %v", err, resp.StatusCode(), "",
		)
	}
	return nil
}
