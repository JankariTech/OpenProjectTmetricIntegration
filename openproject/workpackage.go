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
	"net/url"
)

type WorkPackage struct {
	Subject string `json:"subject"`
	Id      int    `json:"id"`
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
