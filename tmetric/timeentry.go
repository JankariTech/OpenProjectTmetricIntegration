/*
Copyright © 2024 JankariTech Pvt. Ltd. info@jankaritech.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package tmetric

import (
	"encoding/json"
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/JankariTech/OpenProjectTmetricIntegration/openproject"
	"github.com/go-resty/resty/v2"
	"regexp"
	"strconv"
	"time"
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

func NewDummyTimeEntry(workPackage openproject.WorkPackage, openProjectUrl string, projectId int) *DummyTimeEntry {
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

func (timeEntry *TimeEntry) Delete(config config.Config, user User) error {
	httpClient := resty.New()
	_, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Delete(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries/%v`,
				config.TmetricAPIV3BaseUrl,
				user.ActiveAccountId,
				timeEntry.Id,
			),
		)
	return err
}

func (timeEntry *TimeEntry) Update(config config.Config, user User) error {
	entryJSON, err := json.Marshal(timeEntry)
	if err != nil {
		return fmt.Errorf("error marshalling tmetric entry to JSON: %v", err)
	}
	httpClient := resty.New()
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		SetHeader("Content-Type", "application/json").
		SetBody(entryJSON).
		Put(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries/%v`,
				config.TmetricAPIV3BaseUrl,
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

func (timeEntry *TimeEntry) GetPossibleWorkTypes(config config.Config, user User) ([]Tag, error) {
	httpClient := resty.New()

	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("projectId", strconv.Itoa(timeEntry.Project.Id)).
		Get(fmt.Sprintf(
			`%vaccounts/%v/timeentries/tags`,
			config.TmetricAPIV3BaseUrl,
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

// GetIssueIdAsInt returns the issue id as an integer
// the issue Id in tmetric is a string e.g. #1234, but for OpenProject we need the integer to construct the URLs
func (timeEntry *TimeEntry) GetIssueIdAsInt() (int, error) {
	issueIdStr := regexp.MustCompile(`#(\d+)`).
		FindStringSubmatch(timeEntry.Task.ExternalLink.IssueId)[1]
	return strconv.Atoi(issueIdStr)
}

/*
this is the only way to create an external task in tmetric.
This task is needed to have an issueId of OpenProject assigned to a time entry.
*/
func CreateDummyTimeEntry(
	workPackage openproject.WorkPackage, tmetricUser *User, config *config.Config,
) (*TimeEntry, error) {
	dummyTimeEntry := NewDummyTimeEntry(workPackage, config.OpenProjectUrl, config.TmetricDummyProjectId)
	dummyTimerString, _ := json.Marshal(dummyTimeEntry)
	httpClient := resty.New()
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		SetHeader("Content-Type", "application/json").
		SetBody(dummyTimerString).
		Post(fmt.Sprintf(
			`%vaccounts/%v/timer/issue`,
			config.TmetricAPIBaseUrl,
			tmetricUser.ActiveAccountId,
		))
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"could not create dummy time entry. Is 'tmetric.dummyProjectId' set correctly in the config?\n"+
				"Error : '%v'. HTTP-Status-Code: %v",
			err, resp.StatusCode(),
		)
	}

	resp, err = httpClient.R().
		SetAuthToken(config.TmetricToken).
		SetQueryString("userId=" + strconv.Itoa(tmetricUser.Id)).
		Get(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries/latest`,
				config.TmetricAPIV3BaseUrl,
				tmetricUser.ActiveAccountId,
			),
		)

	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"could not find latest time entry. Error : '%v'. HTTP-Status-Code: %v",
			err, resp.StatusCode(),
		)
	}
	latestTimeEntry := TimeEntry{}
	err = json.Unmarshal(resp.Body(), &latestTimeEntry)

	if err != nil || latestTimeEntry.Note != "to-delete-only-created-to-create-an-external-task" {
		return nil, fmt.Errorf(
			"could not find dummy time entry",
		)
	}
	return &latestTimeEntry, nil
}

func (timeEntry *TimeEntry) GetWorkType() (string, error) {
	for _, tag := range timeEntry.Tags {
		if tag.IsWorkType {
			return tag.Name, nil
		}
	}
	return "", fmt.Errorf("no work type found")
}

func (timeEntry *TimeEntry) ConvertToOpenProjectTimeEntry(activityId int) (openproject.TimeEntry, error) {
	opTimeEntry := openproject.TimeEntry{
		Ongoing: false,
	}
	opTimeEntry.Comment.Raw = timeEntry.Note
	issueId, err := timeEntry.GetIssueIdAsInt()
	if err != nil {
		return openproject.TimeEntry{}, err
	}
	opTimeEntry.Links.WorkPackage.Href = fmt.Sprintf("/api/v3/work_packages/%d", issueId)
	opTimeEntry.Links.Activity.Href = fmt.Sprintf("/api/v3/time_entries/activities/%d", activityId)
	iso8601Duration, spentOn, err := timeEntry.getIso8601Duration()
	if err != nil {
		return openproject.TimeEntry{}, err
	}
	opTimeEntry.Hours = iso8601Duration
	opTimeEntry.SpentOn = spentOn
	return opTimeEntry, err
}

func (timeEntry *TimeEntry) getIso8601Duration() (string, string, error) {
	startTimeParsed, err := time.Parse("2006-01-02T15:04:05", timeEntry.StartTime)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse startTime: %v", err)
	}
	endTimeParsed, err := time.Parse("2006-01-02T15:04:05", timeEntry.EndTime)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse endTime: %v", err)
	}
	duration := endTimeParsed.Sub(startTimeParsed)
	if duration < 0 {
		return "", "", fmt.Errorf("end time is before start time")
	}

	iso8601Duration := fmt.Sprintf(
		"P%dDT%dH%dM%dS",
		int(duration.Hours()/24),
		int(duration.Hours())%24,
		int(duration.Minutes())%60,
		int(duration.Seconds())%60,
	)
	spentOn := startTimeParsed.Format("2006-01-02")
	return iso8601Duration, spentOn, nil
}
