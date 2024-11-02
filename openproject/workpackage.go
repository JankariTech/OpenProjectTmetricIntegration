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
)

type WorkPackage struct {
	Subject string `json:"subject"`
	Id      int    `json:"id"`
}

func NewWorkPackage(id int, subject string) WorkPackage {
	return WorkPackage{
		Id:      id,
		Subject: subject,
	}
}

func GetWorkpackage(workPackageId string, config *config.Config) (WorkPackage, error) {
	openProjectHttpClient := resty.New()

	wpURL, _ := url.JoinPath(config.OpenProjectUrl, "/api/v3/work_packages/", workPackageId)
	resp, err := openProjectHttpClient.R().
		SetBasicAuth("apikey", config.OpenProjectToken).
		Get(wpURL)

	if err != nil || resp.StatusCode() != 200 {
		return WorkPackage{}, fmt.Errorf("could not find WP in %v", config.OpenProjectUrl)
	}

	var workPackage WorkPackage
	err = json.Unmarshal(resp.Body(), &workPackage)
	if err != nil {
		return WorkPackage{}, fmt.Errorf(
			"error parsing work packages response or no work packages found: %v", err,
		)
	}
	return workPackage, err
}

func (w *WorkPackage) GetAllowedActivities(config config.Config) ([]Activity, error) {
	httpClient := resty.New()
	openProjectUrl, _ := url.JoinPath(config.OpenProjectUrl, "/api/v3/time_entries/form")
	body := fmt.Sprintf(`{"_links":{"workPackage":{"href":"/api/v3/work_packages/%d"}}}`, w.Id)
	resp, err := httpClient.R().
		SetBasicAuth("apikey", config.OpenProjectToken).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(openProjectUrl)
	if err != nil || resp.StatusCode() != 200 {
		return []Activity{}, fmt.Errorf(
			"could not fetch allowed activities for work package '%v' from '%v'.\n"+
				"Are 'Time and costs' activated for the project?\n"+
				"Error: '%v'. HTTP-Status Code: %v",
			w.Id,
			config.OpenProjectUrl,
			err,
			resp.StatusCode(),
		)
	}

	activitiesJSON := gjson.GetBytes(resp.Body(), "_embedded.schema.activity._embedded.allowedValues")
	validationErrorJSON := gjson.GetBytes(resp.Body(), "_embedded.validationErrors.workPackage.message")
	var activities []Activity
	err = json.Unmarshal([]byte(activitiesJSON.String()), &activities)
	if err != nil {
		return []Activity{}, fmt.Errorf(
			"error parsing work packages response or no work packages found: %v", err,
		)
	}
	if validationErrorJSON.Exists() {
		return []Activity{}, fmt.Errorf(
			"work packages '%v' not found. Error: %v", w.Id, validationErrorJSON.String(),
		)
	}
	return activities, nil
}
