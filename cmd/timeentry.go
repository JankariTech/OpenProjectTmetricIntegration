package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"strconv"
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

type Tag struct {
	Id         int    `json:"id"`
	Name       string `json:"name"`
	IsWorkType bool   `json:"isWorkType"`
}

type TimeEntry struct {
	Id        int     `json:"id"`
	StartTime string  `json:"startTime"`
	EndTime   string  `json:"endTime"`
	Task      Task    `json:"task"`
	Project   Project `json:"project"`
	Note      string  `json:"note"`
	Tags      []Tag   `json:"tags"`
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

func (timeEntry *TimeEntry) getPossibleWorkTypes(config Config, user TmetricUser) ([]Tag, error) {
	httpClient := resty.New()

	resp, err := httpClient.R().
		SetAuthToken(config.tmetricToken).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("projectId", strconv.Itoa(timeEntry.Project.Id)).
		Get(fmt.Sprintf(
			`%vaccounts/%v/timeentries/tags`,
			config.tmetricAPIV3BaseUrl,
			user.ActiveAccountId,
		))
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"could not get tags from t-metric\n"+
				"Error : '%v'. HTTP-Status-Code: %v",
			err, resp.StatusCode(),
		)
	}
	var tags []Tag
	err = json.Unmarshal(resp.Body(), &tags)
	if err != nil {
		return nil, fmt.Errorf("error parsing response: %v", err)
	}
	var workTypes []Tag
	for _, tag := range tags {
		if tag.IsWorkType {
			workTypes = append(workTypes, tag)
		}
	}
	return workTypes, nil
}
