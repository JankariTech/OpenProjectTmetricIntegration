package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
)

type ExternalLink struct {
	Caption string `json:"caption"`
	Link    string `json:"link"`
	IssueId string `json:"issueId"`
}

type Task struct {
	Id           int          `json:"id"`
	Name         string       `json:"name"`
	ExternalLink ExternalLink `json:"externalLink"`
}

type Client struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Client Client `json:"client"`
}

type TimeEntry struct {
	Id        int     `json:"id"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	Task      Task    `json:"task"`
	Project   Project `json:"project"`
	Note      string  `json:"note"`
}

type DummyTimeEntry struct {
	IsStarted   bool   `json:"isStarted"`
	ShowIssueId bool   `json:"showIssueId"`
	Duration    int    `json:"duration"`
	IssueId     string `json:"issueId"`
	IssueName   string `json:"issueName"`
	IssueUrl    string `json:"issueUrl"`
	ServiceUrl  string `json:"serviceUrl"`
	ServiceType string `json:"serviceType"`
	Description string `json:"description"`
	TagNames    []string
	ProjectId   int `json:"projectId"`
}

func newDummyTimeEntry(workPackage WorkPackage, openProjectUrl string, projectId int) *DummyTimeEntry {
	return &DummyTimeEntry{
		IsStarted:   true,
		ShowIssueId: true,
		Duration:    0,
		IssueId:     fmt.Sprintf("#%v", workPackage.Id),
		IssueName:   workPackage.Subject,
		IssueUrl:    fmt.Sprintf("/work_packages/%v", workPackage.Id),
		ServiceUrl:  openProjectUrl,
		ServiceType: "OpenProject",
		Description: "to-delete-only-created-to-create-an-external-task",
		ProjectId:   projectId,
	}
}

func (timeEntry *TimeEntry) delete(config Config, user TmetricUser) error {
	httpClient := resty.New()
	_, err := httpClient.R().
		SetAuthToken(config.tmetricToken).
		Delete(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries/%v`,
				config.tmetricAPIV3BaseUrl,
				user.ActiveAccountId,
				timeEntry.Id,
			),
		)
	return err
}

func (timeEntry *TimeEntry) update(config Config, user TmetricUser) error {
	entryJSON, err := json.Marshal(timeEntry)
	if err != nil {
		return fmt.Errorf("error marshalling tmetric entry to JSON: %v", err)
	}
	httpClient := resty.New()
	resp, err := httpClient.R().
		SetAuthToken(config.tmetricToken).
		SetHeader("Content-Type", "application/json").
		SetBody(entryJSON).
		Put(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries/%v`,
				config.tmetricAPIV3BaseUrl,
				user.ActiveAccountId,
				timeEntry.Id,
			),
		)

	if err != nil || resp.StatusCode() != 200 {
		return fmt.Errorf(
			"could not update time entry. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}
	return nil
}
