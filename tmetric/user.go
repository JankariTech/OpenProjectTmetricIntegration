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
	"regexp"
)

type User struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	ActiveAccountId int    `json:"activeAccountId"`
}

// UserV2 API V2 has a different structure for the user, see:
// https://app.tmetric.com/api-docs/v2/#/AccountMembers/accountmembers-get-api-accounts-accountid-members
type UserV2 struct {
	AccountMemberId int `json:"accountMemberId"`
	UserProfileId   int `json:"userProfileId"`
	UserProfile     struct {
		UserProfileId   int    `json:"userProfileId"`
		ActiveAccountId int    `json:"activeAccountId"`
		UserName        string `json:"userName"`
	} `json:"userProfile"`
	AccountMemberScope struct {
		GroupMembership []struct {
			Name string `json:"name"`
			Id   int    `json:"id"`
		} `json:"groupMembership"`
	} `json:"accountMemberScope"`
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

// FindUserByName searches for users that match the given search and returns the first match.
func FindUserByName(config *config.Config, userMe User, search string) (User, error) {
	httpClient := resty.New()

	resp, err := httpClient.R().
		SetAuthToken(config.TmetricToken).
		Get(fmt.Sprintf("%vaccounts/%v/members", config.TmetricAPIBaseUrl, userMe.ActiveAccountId))
	if err != nil || resp.StatusCode() != 200 {
		return User{}, fmt.Errorf(
			"cannot get members for tmetric account '%v'. Error: '%v'. HTTP status code: %v",
			userMe.ActiveAccountId, err, resp.StatusCode(),
		)
	}
	var users []UserV2
	err = json.Unmarshal(resp.Body(), &users)
	if err != nil {
		return User{}, fmt.Errorf("error parsing users response: %v\n", err)
	}
	for _, user := range users {
		if matched, _ := regexp.MatchString(".*"+search+".*", user.UserProfile.UserName); matched {
			return User{
				Id:              user.UserProfile.UserProfileId,
				Name:            user.UserProfile.UserName,
				ActiveAccountId: user.UserProfile.ActiveAccountId,
			}, nil
		}
	}
	return User{}, fmt.Errorf("cannot find a user in tmetric with a name matching '%v'", search)

}
