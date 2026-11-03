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

package jira

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-resty/resty/v2"
)

type Issue struct {
	Key    string `json:"key"`
	Id     int    `json:"string"`
	Fields struct {
		Summary string `json:"summary"`
	} `json:"fields"`
}

func GetIssue(IssueId string, config *config.Config) (Issue, error) {
	jiraHttpClient := resty.New()

	wpURL, _ := url.JoinPath(config.JiraUrl, "/rest/api/3/issue/", IssueId)
	resp, err := jiraHttpClient.R().
		SetBasicAuth(config.JiraUser, config.JiraToken).
		Get(wpURL)

	if err != nil || resp.StatusCode() != 200 {
		return Issue{}, fmt.Errorf("could not find issue in %v", config.JiraUrl)
	}

	var issue Issue
	err = json.Unmarshal(resp.Body(), &issue)
	if err != nil {
		return Issue{}, fmt.Errorf(
			"error parsing issue response or no issue found: %v", err,
		)
	}
	return issue, err
}
