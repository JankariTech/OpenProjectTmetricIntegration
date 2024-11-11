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
	"github.com/go-resty/resty/v2"
	"os"
)

type User struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	ActiveAccountId int    `json:"activeAccountId"`
}

func NewUser() User {
	config := config.NewConfig()

	httpClient := resty.New()

	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(config.TmetricAPIV3BaseUrl + "user")
	if err != nil || resp.StatusCode() != 200 {
		fmt.Fprintf(os.Stderr, "cannot reach tmetric server\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "status code: %v\n", resp.StatusCode())
		}
		os.Exit(1)
	}
	var user User
	err = json.Unmarshal(resp.Body(), &user)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing response: %v\n", err)
		os.Exit(1)
	}

	return user
}
