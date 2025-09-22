package tmetric

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-resty/resty/v2"
)

// ClientV2 represents a client that is returned by the Tmetric API V2
// see https://app.tmetric.com/api-docs/v2/#/Clients/clients-get-api-accounts-accountid-clients
type ClientV2 struct {
	Id   int    `json:"clientId"`
	Name string `json:"clientName"`
}

// TagV2 represents a tag that is returned by the Tmetric API V2
// see https://app.tmetric.com/api-docs/v2/#/Tags/tags-get-api-accounts-accountid-tags
type TagV2 struct {
	Id         int    `json:"tagId"`
	Name       string `json:"tagName"`
	IsWorkType bool   `json:"isWorkType"`
}

// ProjectV2 represents a project that is returned by the Tmetric API V2
// see https://app.tmetric.com/api-docs/v2/#/ProjectsV2/projectsv2-get-api-accounts-accountid-projects
type ProjectV2 struct {
	Id   int    `json:"projectId"`
	Name string `json:"projectName"`
}

type Team struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func GetAllTeams(config *config.Config, tmetricUser User) ([]Team, error) {
	httpClient := resty.New()
	tmetricUrl, _ := url.JoinPath(
		config.TmetricAPIV3BaseUrl, "accounts/", strconv.Itoa(tmetricUser.ActiveAccountId), "/teams/managed",
	)
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(tmetricUrl)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read teams from tmetric. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}
	var teams []Team
	err = json.Unmarshal(resp.Body(), &teams)
	if err != nil {
		return nil, fmt.Errorf("error parsing teams response: %v\n", err)
	}
	return teams, nil
}

func getTeamByName(config *config.Config, tmetricUser User, name string) (Team, error) {
	teams, err := GetAllTeams(config, tmetricUser)
	if err != nil {
		return Team{}, err
	}
	for _, team := range teams {
		if team.Name == name {
			return team, nil
		}
	}
	return Team{}, fmt.Errorf("could not find any team with name '%v'", name)
}

func getAllProjects(config *config.Config, tmetricUser User) ([]ProjectV2, error) {
	httpClient := resty.New()
	tmetricUrl, _ := url.JoinPath(
		config.TmetricAPIBaseUrl, "accounts/", strconv.Itoa(tmetricUser.ActiveAccountId), "/projects",
	)
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(tmetricUrl)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read projects from tmetric. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}
	var projects []ProjectV2
	err = json.Unmarshal(resp.Body(), &projects)
	if err != nil {
		return nil, fmt.Errorf("error parsing project response: %v\n", err)
	}
	return projects, nil
}

func getProjectByName(config *config.Config, tmetricUser User, name string) (ProjectV2, error) {
	projects, err := getAllProjects(config, tmetricUser)
	if err != nil {
		return ProjectV2{}, err
	}
	for _, project := range projects {
		if project.Name == name {
			return project, nil
		}
	}
	return ProjectV2{}, fmt.Errorf("could not find any project in tmetric with name '%v'", name)
}

func GetAllWorkTypes(config *config.Config, tmetricUser User) ([]Tag, error) {
	httpClient := resty.New()
	tmetricUrl, _ := url.JoinPath(
		config.TmetricAPIBaseUrl, "accounts/", strconv.Itoa(tmetricUser.ActiveAccountId), "/tags",
	)
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(tmetricUrl)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read tags from tmetric. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}
	var tags []TagV2
	err = json.Unmarshal(resp.Body(), &tags)
	if err != nil {
		return nil, fmt.Errorf("error parsing tags response: %v\n", err)
	}
	var workTypes []Tag
	for _, tag := range tags {
		if tag.IsWorkType {
			workTypes = append(workTypes, Tag{
				Id:         tag.Id,
				Name:       tag.Name,
				IsWorkType: tag.IsWorkType,
			})
		}
	}
	return workTypes, err
}

func getWorkTypeByName(config *config.Config, tmetricUser User, name string) (Tag, error) {
	worktypes, err := GetAllWorkTypes(config, tmetricUser)
	if err != nil {
		return Tag{}, err
	}
	for _, tag := range worktypes {
		if tag.Name == name {
			return tag, nil
		}
	}
	return Tag{}, fmt.Errorf("could not find any work type with name '%v'", name)
}

func getClientByName(config *config.Config, tmetricUser User, name string) (Client, error) {
	httpClient := resty.New()
	tmetricUrl, _ := url.JoinPath(
		config.TmetricAPIBaseUrl, "accounts/", strconv.Itoa(tmetricUser.ActiveAccountId), "/clients",
	)
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(tmetricUrl)
	if err != nil || resp.StatusCode() != 200 {
		return Client{}, fmt.Errorf(
			"cannot read clients from tmetric. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}
	var clients []ClientV2
	err = json.Unmarshal(resp.Body(), &clients)
	if err != nil {
		return Client{}, fmt.Errorf("error parsing clients response: %v\n", err)
	}
	for _, client := range clients {
		if client.Name == name {
			return Client{
				Id:   client.Id,
				Name: client.Name,
			}, nil
		}
	}
	return Client{}, fmt.Errorf("could not find any client with name '%v'", name)
}

func GetAllTimeEntries(config *config.Config, tmetricUser User, startDate string, endDate string) ([]TimeEntry, error) {
	httpClient := resty.New()
	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(
			fmt.Sprintf(
				`%vaccounts/%v/timeentries?userId=%v&startDate=%v&endDate=%v`,
				config.TmetricAPIV3BaseUrl,
				tmetricUser.ActiveAccountId,
				tmetricUser.Id,
				startDate,
				endDate,
			),
		)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read timeentries. Error: '%v'. HTTP status code: %v", err, resp.StatusCode(),
		)
	}

	var timeEntries []TimeEntry
	err = json.Unmarshal(resp.Body(), &timeEntries)
	if err != nil {
		return nil, fmt.Errorf("error parsing time entries response: %v\n", err)
	}

	var timeEntriesOfTheSelectedClient []TimeEntry
	for _, entry := range timeEntries {
		if entry.Project.Client.Id == config.ClientIdInTmetric {
			timeEntriesOfTheSelectedClient = append(timeEntriesOfTheSelectedClient, entry)
		}
	}
	sort.Slice(timeEntriesOfTheSelectedClient, func(i, j int) bool {
		return timeEntriesOfTheSelectedClient[i].StartTime < timeEntriesOfTheSelectedClient[j].StartTime
	})

	return timeEntriesOfTheSelectedClient, nil
}

func GetEntriesNotTransferredToOpenProject(timeEntries []TimeEntry, TmetricTagTransferredToOpenProject string) []TimeEntry {
	var filteredEntries []TimeEntry
	for _, entry := range timeEntries {
		hasTransferredTag := false
		for _, tag := range entry.Tags {
			if tag.Name == TmetricTagTransferredToOpenProject {
				hasTransferredTag = true
				break
			}
		}
		if !hasTransferredTag {
			filteredEntries = append(filteredEntries, entry)
		}
	}

	return filteredEntries
}

func GetEntriesWithoutWorkType(timeEntries []TimeEntry) []TimeEntry {
	// get all entries that belong to the client and do not have any work-type set
	var entriesWithoutWorkType []TimeEntry
	for _, entry := range timeEntries {
		workTypeFound := false
		for _, tag := range entry.Tags {
			if tag.IsWorkType {
				workTypeFound = true
				break
			}
		}
		if !workTypeFound {
			entriesWithoutWorkType = append(entriesWithoutWorkType, entry)
		}
	}
	return entriesWithoutWorkType
}

func GetEntriesWithoutLinkToOpenProject(config *config.Config, timeEntries []TimeEntry) []TimeEntry {
	var entriesWithoutLink []TimeEntry
	for _, entry := range timeEntries {
		_, err := entry.GetIssueIdAsInt()
		if entry.Task.Id == 0 ||
			err != nil ||
			entry.Task.ExternalLink.IssueId == "" ||
			(!strings.HasPrefix(entry.Task.ExternalLink.Link, config.TmetricExternalTaskLink+"work_packages")) {
			entriesWithoutLink = append(entriesWithoutLink, entry)
		}
	}
	return entriesWithoutLink
}

func GetAllAssignedWorkTypes(timeEntries []TimeEntry) []string {
	workTypeSet := make(map[string]struct{})
	for _, entry := range timeEntries {
		for _, tag := range entry.Tags {
			if tag.IsWorkType {
				workTypeSet[tag.Name] = struct{}{}
			}
		}
	}

	var workTypes []string
	for workType := range workTypeSet {
		workTypes = append(workTypes, workType)
	}

	return workTypes
}
