package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-resty/resty/v2"
)

type WorklogResponse struct {
	Values []struct {
		Id int `json:"worklogId"`
	} `json:"values"`
	Since    int64 `json:"since"`
	Until    int64 `json:"until"`
	LastPage bool  `json:"lastPage"`
}

func GetAllWorklogs(config *config.Config, user string, startDate string, endDate string, timezone string) ([]Worklog, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load time zone: %v", err)
	}
	start, _ := time.ParseInLocation("2006-01-02", startDate, loc)
	end, _ := time.ParseInLocation("2006-01-02", endDate, loc)
	end = end.AddDate(0, 0, 1) // add one day and then check if the worklog starts BEFORE that
	httpClient := resty.New()
	var worklogIds []int
	lastPage := false
	since := start.UnixMilli()
	for lastPage != true {
		updatedWorklogURL, _ := url.JoinPath(config.JiraUrl, "rest/api/3/worklog/updated")
		updatedWorklogURL += "?since=" + strconv.FormatInt(since, 10)

		resp, err := httpClient.R().
			SetBasicAuth(config.JiraUser, config.JiraToken).
			SetHeader("Content-Type", "application/json; charset=utf-8").
			Get(updatedWorklogURL)
		if err != nil || resp.StatusCode() != 200 {
			return nil, fmt.Errorf(
				"could not get list of worklogs updated since %v from Jira. Error: '%v'. HTTP status code: %v. HTTP response %v", strconv.FormatInt(since, 10), err, resp.StatusCode(), string(resp.Body()),
			)
		}
		var worklogResp WorklogResponse
		err = json.Unmarshal(resp.Body(), &worklogResp)
		if err != nil {
			return nil, err
		}
		lastPage = worklogResp.LastPage
		since = worklogResp.Until
		for _, value := range worklogResp.Values {
			worklogIds = append(worklogIds, value.Id)
		}
	}

	worklogListURL, _ := url.JoinPath(config.JiraUrl, "rest/api/3/worklog/list")
	worklogListURL += "?expand=properties"

	var worklogs []Worklog

	const maxWorklogsPerRequest = 1000
	for offset := 0; offset < len(worklogIds); offset += maxWorklogsPerRequest {
		batchEnd := offset + maxWorklogsPerRequest
		if batchEnd > len(worklogIds) {
			batchEnd = len(worklogIds)
		}
		worklogListBody, err := json.Marshal(map[string][]int{"ids": worklogIds[offset:batchEnd]})
		if err != nil {
			return nil, err
		}
		resp, err := httpClient.R().
			SetBasicAuth(config.JiraUser, config.JiraToken).
			SetHeader("Content-Type", "application/json; charset=utf-8").
			SetBody(worklogListBody).
			Post(worklogListURL)
		if err != nil {
			return nil, fmt.Errorf("could not get worklogs from Jira: %v", err)
		}
		if resp.StatusCode() != 200 {
			return nil, fmt.Errorf(
				"could not get worklogs from Jira. HTTP status code: %v. HTTP response %v", resp.StatusCode(), string(resp.Body()),
			)
		}
		var batch []Worklog
		if err := json.Unmarshal(resp.Body(), &batch); err != nil {
			return nil, err
		}
		worklogs = append(worklogs, batch...)
	}

	var filteredWorklogs []Worklog
	for _, worklog := range worklogs {
		worklogStarted, _ := time.Parse("2006-01-02T15:04:05.000-0700", worklog.Started)

		if (
		// the tmetric user saved in Jira should contain the display name of the user in tmetric
		worklog.Tmetric.UserDisplayName == user ||
			(
			// but if there is no tmetric user saved in Jira, compare the Author to
			// the display name of the user in tmetric
			worklog.Tmetric.UserDisplayName == "" && worklog.Author.Name == user)) &&
			(worklogStarted.Equal(start) || worklogStarted.After(start)) &&
			worklogStarted.Before(end) {
			filteredWorklogs = append(filteredWorklogs, worklog)
		}
	}

	return filteredWorklogs, nil
}
