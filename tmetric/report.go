package tmetric

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-resty/resty/v2"
)

type ReportItem struct {
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	User          string `json:"user"`
	IssueId       string `json:"issueId"`
	WorkpackageId string
}

type Report struct {
	ReportItems []ReportItem
	Duration    time.Duration
}

func (reportItem *ReportItem) getParsedTime(startTime bool) (time.Time, error) {
	stringToParse := ""
	if startTime {
		stringToParse = reportItem.StartTime
	} else {
		stringToParse = reportItem.EndTime
	}

	timeParsed, err := time.Parse("2006-01-02T15:04:05Z", stringToParse)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse time: %v", err)
	}
	return timeParsed, nil
}

func (reportItem *ReportItem) getDuration() (time.Duration, error) {
	startTimeParsed, err := reportItem.getParsedTime(true)
	if err != nil {
		return 0, err
	}
	endTimeParsed, err := reportItem.getParsedTime(false)
	if err != nil {
		return 0, err
	}

	duration := endTimeParsed.Sub(startTimeParsed)
	if duration < 0 {
		return 0, fmt.Errorf("end time is before start time")
	}

	return duration, nil
}

func GetDetailedReport(
	config *config.Config, tmetricUser User, clientName string, tagName string, groupName string, startDate string, endDate string, projects []string,
) (Report, error) {
	client, err := getClientByName(config, tmetricUser, clientName)
	if err != nil {
		return Report{}, err
	}

	team, err := getTeamByName(config, tmetricUser, groupName)
	if err != nil {
		return Report{}, err
	}

	var projectsIds []string // we need a slice of strings for the URL parameters, so let's declare it a string slice
	for _, projectName := range projects {
		project, err := getProjectByName(config, tmetricUser, projectName)
		if err != nil {
			return Report{}, err
		}

		projectsIds = append(projectsIds, strconv.Itoa(project.Id))
	}

	httpClient := resty.New()
	tmetricUrl, _ := url.JoinPath(config.TmetricAPIBaseUrl, "reports/detailed")
	request := httpClient.R()

	if tagName != "" {
		workType, err := getWorkTypeByName(config, tmetricUser, tagName)
		if err != nil {
			return Report{}, err
		}
		request.SetQueryParam("TagList", strconv.Itoa(workType.Id))
	}

	// for this API we have to add one day to actually get the data also for today
	endTime, _ := time.Parse("2006-01-02", endDate)
	endDate = endTime.AddDate(0, 0, 1).Format("2006-01-02")

	resp, err := request.
		SetAuthToken(config.TmetricToken).
		SetQueryParam("AccountId", strconv.Itoa(tmetricUser.ActiveAccountId)).
		SetQueryParam("ClientList", strconv.Itoa(client.Id)).
		SetQueryParam("GroupList", strconv.Itoa(team.Id)).
		SetQueryParamsFromValues(url.Values{
			"ProjectList": projectsIds,
		}).
		SetQueryParam("StartDate", startDate).
		SetQueryParam("EndDate", endDate).
		Get(tmetricUrl)
	if err != nil || resp.StatusCode() != 200 {
		return Report{}, fmt.Errorf(
			"cannot read report from tmetric. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}

	var reportItems []ReportItem
	err = json.Unmarshal(resp.Body(), &reportItems)
	if err != nil {
		return Report{}, fmt.Errorf("error parsing report response: %v\n", err)
	}
	var report Report
	for _, item := range reportItems {
		item.WorkpackageId = strings.Trim(item.IssueId, "#") // remove leading '#' from issue id
		report.ReportItems = append(report.ReportItems, item)
		itemDuration, _ := item.getDuration()
		report.Duration += itemDuration
	}
	return report, nil
}
