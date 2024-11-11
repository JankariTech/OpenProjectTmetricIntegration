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

package openproject

import (
	"encoding/json"
	"fmt"
	"github.com/JankariTech/OpenProjectTmetricIntegration/config"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
	"net/url"
	"strconv"
)

func GetAllTimeEntries(config *config.Config, user User, startDate string, endDate string) ([]TimeEntry, error) {
	httpClient := resty.New()
	openProjectUrl, _ := url.JoinPath(config.OpenProjectUrl, "/api/v3/time_entries")
	var userString string

	// use ether "me" in the request or the id of the user
	if user == (User{}) {
		userString = "me"
	} else {
		userString = strconv.Itoa(user.Id)
	}
	resp, err := httpClient.R().
		SetBasicAuth("apikey", config.OpenProjectToken).
		SetHeader("Content-Type", "application/json").
		SetQueryParam("pageSize", "3000").
		SetQueryParam("sortBy", "[[\"updated_at\",\"desc\"]]").
		// the operator is '<>d' and means between the dates
		SetQueryParam("filters", fmt.Sprintf(
			`[{"user":{"operator":"=","values":["%v"]}},{"spent_on":{"operator":"\u003c\u003ed","values":["%v","%v"]}}]`,
			userString,
			startDate,
			endDate),
		).
		Get(openProjectUrl)
	if err != nil || resp.StatusCode() != 200 {
		return nil, fmt.Errorf(
			"cannot read timeentries from OpenProject for user '%v'. Error: '%v'. HTTP status code: %v",
			user.Name,
			err,
			resp.StatusCode(),
		)
	}
	var timeEntries []TimeEntry
	timeEntriesJSON := gjson.GetBytes(resp.Body(), "_embedded.elements")
	err = json.Unmarshal([]byte(timeEntriesJSON.String()), &timeEntries)
	if err != nil {
		return []TimeEntry{}, fmt.Errorf(
			"error parsing time entries response: %v", err,
		)
	}
	return timeEntries, nil
}
