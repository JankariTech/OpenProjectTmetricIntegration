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
	Key              string  `json:"key"`
	Comment          Comment `json:"comment"`
	Started          string  `json:"started"`
	TimeSpentSeconds int     `json:"timeSpentSeconds"`
	Author           Author  `json:"author"`
	UpdateAuthor     Author  `json:"updateAuthor"`
	IssueId          string  `json:"issueId"`
	Tmetric          Tmetric `json:"-"`
	TextComment      string  `json:"-"`
}

type Tmetric struct {
	UserDisplayName string `json:"userdisplayname"`
	UserId          int    `json:"userid"`
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

type Author struct {
	AccountId string `json:"accountId"`
	Email     string `json:"emailAddress"`
	Name      string `json:"displayName"`
}

type WorklogProperty struct {
	Key   string  `json:"key"`
	Value Tmetric `json:"value"`
}

// MarshalJSON puts the tmetric properties into the correct JSON structure
func (worklog *Worklog) MarshalJSON() ([]byte, error) {
	type Alias Worklog

	return json.Marshal(struct {
		Alias
		Properties []WorklogProperty `json:"properties"`
	}{
		Alias: Alias(*worklog),
		Properties: []WorklogProperty{
			{
				Key: "tmetric",
				Value: Tmetric{
					UserDisplayName: worklog.Tmetric.UserDisplayName,
					UserId:          worklog.Tmetric.UserId,
				},
			},
		},
	})
}

func (worklog *Worklog) UnmarshalJSON(b []byte) error {
	type Alias Worklog

	var tmpStruct struct {
		Alias
		Properties []WorklogProperty `json:"properties"`
	}

	err := json.Unmarshal(b, &tmpStruct)
	if err != nil {
		return err
	}
	worklog.Author = tmpStruct.Author
	worklog.UpdateAuthor = tmpStruct.UpdateAuthor
	worklog.Comment = tmpStruct.Comment
	worklog.IssueId = tmpStruct.IssueId
	worklog.Started = tmpStruct.Started
	worklog.TimeSpentSeconds = tmpStruct.TimeSpentSeconds
	for _, property := range tmpStruct.Properties {
		if property.Key == "tmetric" {
			worklog.Tmetric = property.Value
		}
	}
	worklog.TextComment = getTextFromComment(tmpStruct.Comment.Content)

	return nil
}

func getTextFromComment(content []CommentNode) string {
	result := ""
	for _, contentNode := range content {
		if contentNode.Type == "text" {
			result += contentNode.Text
		}
		if len(contentNode.Content) > 0 {
			result += getTextFromComment(contentNode.Content)
		}
	}
	return result + " "
}

func (worklog *Worklog) Save(config config.Config) error {
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
